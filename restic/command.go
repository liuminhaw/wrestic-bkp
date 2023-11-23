package restic

import (
	"fmt"
	"os/exec"
)

// ResticCheck checks if "restic" command is available in system path
func ResticCheck() error {
	_, err := exec.LookPath("restic")
	if err != nil {
		return fmt.Errorf("restic command check: %w", err)
	}

	return nil
}
