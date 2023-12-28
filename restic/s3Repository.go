package restic

import (
	"fmt"
	"os"
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
	r.initCredential()

	commandArg := []string{"init", "-r", fmt.Sprintf("s3:s3.amazonaws.com/%s", r.Destination)}
	output, err := execOutput(commandArg)
	if err != nil {
		return output, fmt.Errorf("s3BackupRepository init: %w", err)
	}

	return output, nil
}

func (r S3BackupRepository) Backup() error {
	os.Setenv(passwordEnv, r.Password)
	os.Setenv(resticProgressFPS, resticProgressFPSValue)

	// Set S3 credential
	r.initCredential()

	commandArg := []string{"backup", "-r", fmt.Sprintf("s3:s3.amazonaws.com/%s", r.Destination)}
	commandArg = append(commandArg, r.Sources...)

	// Add exclude option
	for _, exclude := range r.Excludes {
		excludeOpt := fmt.Sprintf("--exclude=%s", exclude)
		commandArg = append(commandArg, excludeOpt)
	}

	err := execStream(commandArg, true)
	if err != nil {
		return fmt.Errorf("s3BackupRepository backup: %w", err)
	}

	// Check repository integrity and consistency after backup
	if err := r.Check(); err != nil {
		return fmt.Errorf("s3BackupRepository backup: %w", err)
	}

	return nil
}

func (r S3BackupRepository) Snapshots() ([]byte, error) {
	os.Setenv(passwordEnv, r.Password)

	r.initCredential()

	commandArg := []string{"snapshots", "-r", fmt.Sprintf("s3:s3.amazonaws.com/%s", r.Destination)}
	output, err := execOutput(commandArg)
	if err != nil {
		return output, fmt.Errorf("s3BackupRepository snapshots: %w", err)
	}

	return output, nil
}

func (r S3BackupRepository) Check() error {
	os.Setenv(passwordEnv, r.Password)

	r.initCredential()

	commandArg := []string{"check", "-r", fmt.Sprintf("s3:s3.amazonaws.com/%s", r.Destination)}
	err := execStream(commandArg, false)
	if err != nil {
		return fmt.Errorf("s3BackupRepository snapshots: %w", err)
	}

	return nil
}

func (r S3BackupRepository) initCredential() {
	os.Setenv(awsAccessKeyIdEnv, r.AccessKeyId)
	os.Setenv(awsSecretAccessKeyEnv, r.SecretAccessKey)
}
