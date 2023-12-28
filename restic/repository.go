package restic

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
)

const (
	resticCmd              string = "restic"
	passwordEnv            string = "RESTIC_PASSWORD"
	resticProgressFPS      string = "RESTIC_PROGRESS_FPS"
	resticProgressFPSValue string = "2"
)

type ResticRepository interface {
	Init() ([]byte, error)
	Backup() error
	Snapshots() ([]byte, error)
	Check() error
}

func execOutput(cmdArgs []string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.Command(resticCmd, cmdArgs...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return stderr.Bytes(), fmt.Errorf("execOutput: %w", err)
	}

	return output, nil
}

// execStream runs restic command with cmdArgs and stream output to stdout
// useLinesCount determines if cleaning screen operation will clean only the line match regex pattern
// or all lines before match regex pattern. Set to true for all lines cleaning
// and false for matched line clean
func execStream(cmdArgs []string, useLinesCount bool) error {
	var stderr bytes.Buffer
	cmd := exec.Command(resticCmd, cmdArgs...)
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("execStream: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("execStream: command start: %w", err)
	}

	// Read from the pipe
	linesCount := 0
	if !useLinesCount {
		linesCount = 1
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		str := scanner.Text()
		// Check if string start with pattern `[x:xx]`
		pattern := regexp.MustCompile(`^\[\d+:\d\d\]`)
		if pattern.MatchString(str) {
			for i := 0; i < linesCount; i++ {
				fmt.Print("\033[A")
			}
			fmt.Print("\033[K")
			if useLinesCount {
				linesCount = 0
			}
		}
		fmt.Println(str) // Print each line of output
		if useLinesCount {
			linesCount++
		}
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println(stderr.String())
		return fmt.Errorf("execStream: command wait: %w", err)
	}

	return nil
}

func countStringLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}

	if len(s) > 0 && s[len(s)-1] != '\n' {
		count++
	}

	return count
}
