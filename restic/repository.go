package restic

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const passwordEnv string = "RESTIC_PASSWORD"

type ResticRepository interface {
	Init() ([]byte, error)
	Backup() error
}

type LocalBackupRepository struct {
	Password    string
	Destination string
	Sources     []string
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

	commandArg := []string{"backup", "-r", r.Destination}
	commandArg = append(commandArg, r.Sources...)

	var stderr bytes.Buffer
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("localBackupRepository backup: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("localBackupRepository backup: command start: %w", err)
	}

	// Read from the pipe
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // Print each line of output
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println(stderr.String())
		return fmt.Errorf("localBackupRepository backup: command wait: %w", err)
	}

	return nil
}
