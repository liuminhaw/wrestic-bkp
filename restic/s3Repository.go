package restic

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

	var stderr bytes.Buffer
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("s3BackupRepository backup: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("s3BackupRepository backup: command start: %w", err)
	}

	// Read from the pipe
	linesCount := 0
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
			linesCount = 0
		}
		fmt.Println(str) // Print each line of output
		linesCount++
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println(stderr.String())
		return fmt.Errorf("s3BackupRepository backup: command wait: %w", err)
	}

	return nil
}

func (r S3BackupRepository) Snapshots() ([]byte, error) {
	os.Setenv(passwordEnv, r.Password)

	r.initCredential()

	commandArg := []string{"snapshots", "-r", fmt.Sprintf("s3:s3.amazonaws.com/%s", r.Destination)}

	var stderr bytes.Buffer
	cmd := exec.Command("restic", commandArg...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return stderr.Bytes(), fmt.Errorf("s3BackupRepository snapshots: %w", err)
	}

	return output, nil
}

func (r S3BackupRepository) initCredential() {
	os.Setenv(awsAccessKeyIdEnv, r.AccessKeyId)
	os.Setenv(awsSecretAccessKeyEnv, r.SecretAccessKey)
}
