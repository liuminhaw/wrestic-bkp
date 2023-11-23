package restic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type ResticRepository interface {
	Init() ([]byte, error)
}

type LocalBackupRepository struct {
	Password    string
	Destination string
}

func (r LocalBackupRepository) Init() ([]byte, error) {
	// Set repository password
	os.Setenv("RESTIC_PASSWORD", r.Password)

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
