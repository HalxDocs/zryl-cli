package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// splitLines converts raw output into clean line slices
func splitLines(s string) []string {
	lines := []string{}
	current := ""

	for _, r := range s {
		if r == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			continue
		}
		current += string(r)
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// Run represents a single captured execution
type Run struct {
	Command    string    `json:"command"`
	Stdout     []string  `json:"stdout"`
	Stderr     []string  `json:"stderr"`
	ExitCode   int       `json:"exit_code"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	DurationMs int64     `json:"duration_ms"`

	OS   string `json:"os"`
	Arch string `json:"arch"`
	Cwd  string `json:"cwd"`
}

// uploadRun sends the run to Supabase and returns the run ID
func uploadRun(run Run) (string, error) {
	supabaseURL := os.Getenv("ZRYL_SUPABASE_URL")
	supabaseKey := os.Getenv("ZRYL_SUPABASE_SERVICE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return "", fmt.Errorf("missing Supabase environment variables")
	}

	payload, err := json.Marshal(run)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		supabaseURL+"/rest/v1/runs",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var inserted []struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&inserted); err != nil {
		return "", err
	}

	if len(inserted) == 0 || inserted[0].ID == "" {
		return "", fmt.Errorf("no run ID returned from Supabase")
	}

	return inserted[0].ID, nil
}

func main() {
	// Load .env if present (dev only)
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Println("usage: zryl <command> [args...]")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	cmd := exec.Command(command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	end := time.Now()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			fmt.Fprintln(os.Stderr, "zryl: failed to execute command:", err)
			os.Exit(1)
		}
	}

	stdoutLines := splitLines(stdout.String())
	stderrLines := splitLines(stderr.String())

	for _, line := range stdoutLines {
		fmt.Println(line)
	}
	for _, line := range stderrLines {
		fmt.Fprintln(os.Stderr, line)
	}

	cwd, _ := os.Getwd()
	duration := end.Sub(start)

	run := Run{
		Command:    strings.Join(append([]string{command}, args...), " "),
		Stdout:     stdoutLines,
		Stderr:     stderrLines,
		ExitCode:   exitCode,
		StartedAt:  start,
		FinishedAt: end,
		DurationMs: duration.Milliseconds(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Cwd:        cwd,
	}

	runID, uploadErr := uploadRun(run)

	baseURL := os.Getenv("ZRYL_BASE_URL")
	if baseURL == "" {
		baseURL = "https://zryl.vercel.app"
	}

	if uploadErr != nil {
		fmt.Fprintln(os.Stderr, "[zryl] upload failed:", uploadErr)
	} else {
		fmt.Printf("[zryl] link: %s/run/%s\n", baseURL, runID)
	}

	fmt.Printf(
		"[zryl] env: %s/%s · cwd=%s · duration=%s\n",
		runtime.GOOS,
		runtime.GOARCH,
		filepath.Base(cwd),
		duration,
	)

	fmt.Printf(
		"\n[zryl] run captured (%s → %s, exit=%d)\n",
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		exitCode,
	)

	os.Exit(exitCode)
}
