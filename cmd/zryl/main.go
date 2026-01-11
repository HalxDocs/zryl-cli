package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
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

func main() {
	// Ensure a command was provided
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
			fmt.Println("zryl: failed to execute command:", err)
			os.Exit(1)
		}
	}

	// Normalize output
	stdoutLines := splitLines(stdout.String())
	stderrLines := splitLines(stderr.String())

	// Print output exactly as the command produced it
	for _, line := range stdoutLines {
		fmt.Println(line)
	}
	for _, line := range stderrLines {
		fmt.Fprintln(os.Stderr, line)
	}

	fmt.Printf(
		"\n[zryl] run captured (%s → %s, exit=%d)\n",
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		exitCode,
	)

	// Exit with the same exit code as the command
	os.Exit(exitCode)
}
