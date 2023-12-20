package restic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

const (
	awsAccessKeyIdEnv     string = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnv string = "AWS_SECRET_ACCESS_KEY"
)

type S3BackupRepository struct {
	Password        string
	Destination     string
	Sources         []string
	Excludes        []string
	AccessKeyId     string
	SecretAccessKey string
}

func (r S3BackupRepository) Init() ([]byte, error) {
	// Set repository password
	os.Setenv(passwordEnv, r.Password)

	// Set S3 credential
	os.Setenv(awsAccessKeyIdEnv, r.AccessKeyId)
	os.Setenv(awsSecretAccessKeyEnv, r.SecretAccessKey)

	var stderr bytes.Buffer
	commandArg := []string{"init", "-r", fmt.Sprintf("s3:s3.amazonaws.com/%s", r.Destination)}
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return stderr.Bytes(), fmt.Errorf("s3BackupRepository init: %w", err)
	}

	return output, nil
}

func (r S3BackupRepository) Backup() error {
	return nil
}

func (r S3BackupRepository) Snapshots() ([]byte, error) {
	return []byte("To be implement"), nil
}
