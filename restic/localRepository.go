package restic

import (
	"fmt"
	"os"
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

	commandArg := []string{"init", "-r", r.Destination}
	output, err := execOutput(commandArg)
	if err != nil {
		return output, fmt.Errorf("localBackupRepository init: %w", err)
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

	err := execStream(commandArg, true)
	if err != nil {
		return fmt.Errorf("localBackupRepository backup: %w", err)
	}

	// Check repository integrity and consistency after backup
	if err := r.Check(); err != nil {
		return fmt.Errorf("localBackupRepository backup: %w", err)
	}

	return nil
}

func (r LocalBackupRepository) Snapshots() ([]byte, error) {
	os.Setenv(passwordEnv, r.Password)

	commandArg := []string{"snapshots", "-r", r.Destination}
	output, err := execOutput(commandArg)
	if err != nil {
		return output, fmt.Errorf("localBackupRepository snapshots: %w", err)
	}

	return output, nil
}

func (r LocalBackupRepository) Check() error {
	os.Setenv(passwordEnv, r.Password)
	os.Setenv(resticProgressFPS, resticProgressFPSValue)

	commandArg := []string{"check", "-r", r.Destination}

	err := execStream(commandArg, false)
	if err != nil {
		return fmt.Errorf("localBackupRepository check: %w", err)
	}

	return nil
}
