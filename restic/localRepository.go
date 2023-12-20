package restic

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

type LocalBackupRepository struct {
	Password    string
	Destination string
	Sources     []string
	Excludes    []string
}

func (r LocalBackupRepository) Init() ([]byte, error) {
	// Set repository password
	os.Setenv(passwordEnv, r.Password)

	var stderr bytes.Buffer
	commandArg := []string{"init", "-r", r.Destination}
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return stderr.Bytes(), fmt.Errorf("localBackupRepository init: %w", err)
	}

	return output, nil
}

func (r LocalBackupRepository) Backup() error {
	os.Setenv(passwordEnv, r.Password)
	os.Setenv(resticProgressFPS, resticProgressFPSValue)

	commandArg := []string{"backup", "-r", r.Destination}
	commandArg = append(commandArg, r.Sources...)

	// Add exclude option
	for _, exclude := range r.Excludes {
		excludeOpt := fmt.Sprintf("--exclude=%s", exclude)
		commandArg = append(commandArg, excludeOpt)
	}

	// commandArg = append(commandArg, "--dry-run", "-vv")

	var stderr bytes.Buffer
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr

	// fmt.Printf("Command path: %s\n", cmd.Path)
	// fmt.Printf("Command args: %+v\n", cmd.Args)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("localBackupRepository backup: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("localBackupRepository backup: command start: %w", err)
	}

	// Read from the pipe
	linesCount := 0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		str := scanner.Text()
		// Check if string start with pattern `[x:xx]`
		pattern := regexp.MustCompile(`^\[\d:\d\d\]`)
		if pattern.MatchString(str) {
			for i := 0; i < linesCount; i++ {
				fmt.Print("\033[A")
			}
			fmt.Print("\033[K")
			linesCount = 0
		}
		fmt.Println(str) // Print each line of output
		linesCount++
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println(stderr.String())
		return fmt.Errorf("localBackupRepository backup: command wait: %w", err)
	}

	return nil
}

func (r LocalBackupRepository) Snapshots() ([]byte, error) {
	os.Setenv(passwordEnv, r.Password)

	commandArg := []string{"snapshots", "-r", r.Destination}

	var stderr bytes.Buffer
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return stderr.Bytes(), fmt.Errorf("localBackupRepository snapshots: %w", err)
	}

	return output, nil
}
