package main

import (
	"bytes"
	"encoding/json"
	"flag"
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

//
// ---------- helpers ----------
//

func splitLines(s string) []string {
	lines := []string{}
	current := ""

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
			continue
		}
		current += string(r)
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func detectShell() (string, []string) {
	if runtime.GOOS == "windows" {
		// Prefer PowerShell if available
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell", []string{"-NoProfile", "-Command"}
		}
		return "cmd", []string{"/C"}
	}

	// Unix-like systems
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell, []string{"-c"}
	}

	return "/bin/sh", []string{"-c"}
}

//
// ---------- models ----------
//

type Run struct {
	Command    string    `json:"command"`
	Stdout     []string  `json:"stdout"`
	Stderr     []string  `json:"stderr"`
	ExitCode   int       `json:"exit_code"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	DurationMs int64     `json:"duration_ms"`
	OS         string    `json:"os"`
	Arch       string    `json:"arch"`
	Cwd        string    `json:"cwd"`
}

//
// ---------- upload ----------
//

func uploadRun(run Run) (string, error) {
	supabaseURL := os.Getenv("ZRYL_SUPABASE_URL")
	serviceKey := os.Getenv("ZRYL_SUPABASE_SERVICE_KEY")

	if supabaseURL == "" || serviceKey == "" {
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

	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
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
		return "", fmt.Errorf(
			"upload failed (%s): %s",
			resp.Status,
			strings.TrimSpace(string(body)),
		)
	}

	var inserted []struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&inserted); err != nil {
		return "", err
	}

	if len(inserted) == 0 {
		return "", fmt.Errorf("no run ID returned")
	}

	return inserted[0].ID, nil
}

//
// ---------- main ----------
//

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Println("usage: zryl run [--no-upload] [--json] <command>")
		os.Exit(1)
	}

	runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	noUpload := runFlags.Bool("no-upload", false, "do not upload run")
	jsonOut := runFlags.Bool("json", false, "print run as JSON")

	_ = runFlags.Parse(os.Args[2:])
	cmdArgs := runFlags.Args()

	if len(cmdArgs) == 0 {
		fmt.Println("zryl run: missing command")
		os.Exit(1)
	}

	// Shell detection
	shell, shellArgs := detectShell()
	fullCommand := strings.Join(cmdArgs, " ")

	cmd := exec.Command(shell, append(shellArgs, fullCommand)...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	end := time.Now()

	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			fmt.Fprintln(os.Stderr, "zryl: failed to execute command:", err)
			os.Exit(1)
		}
	}

	stdoutLines := splitLines(stdout.String())
	stderrLines := splitLines(stderr.String())

	// Print command output exactly
	for _, l := range stdoutLines {
		fmt.Println(l)
	}
	for _, l := range stderrLines {
		fmt.Fprintln(os.Stderr, l)
	}

	cwd, _ := os.Getwd()
	duration := end.Sub(start)

	run := Run{
		Command:    fullCommand,
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

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(run)
		os.Exit(exitCode)
	}

	if !*noUpload {
		runID, err := uploadRun(run)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[zryl] upload failed:", err)
		} else {
			baseURL := os.Getenv("ZRYL_BASE_URL")
			if baseURL == "" {
				baseURL = "https://zryl.vercel.app"
			}
			fmt.Printf("[zryl] link: %s/run/%s\n", baseURL, runID)
		}
	}

	fmt.Printf(
		"[zryl] env: %s/%s · cwd=%s · duration=%s\n",
		runtime.GOOS,
		runtime.GOARCH,
		filepath.Base(cwd),
		duration,
	)

	fmt.Printf(
		"[zryl] run captured (%s → %s, exit=%d)\n",
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		exitCode,
	)

	os.Exit(exitCode)
}
