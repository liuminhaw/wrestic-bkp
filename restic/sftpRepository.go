package restic

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

const sshConfigSetupMsg string = `
Provided host not found in ssh config file, please add before executing command

ssh config example:
===================================================
Host custom-config-name
    Hostname [ip-address|domain name] 
    User username
    Port 22
    Identityfile /path/to/ssh/key/file
    ServerAliveInterval 60
    ServerAliveCountMax 240 		

`

var (
	ErrSshConfigNotFound = errors.New("user ssh config file not found")
)

type SftpBackupRepository struct {
	Password    string
	Destination string
	Sources     []string
	Excludes    []string
	ConfigHost  string
}

func (r SftpBackupRepository) Init() ([]byte, error) {
	// Set repository password
	os.Setenv(passwordEnv, r.Password)

	// Check if ConfigHost setting exist in ssh config file
	foundHost, err := checkSshHost(r.ConfigHost)
	if err != nil {
		return nil, fmt.Errorf("sftpBackupRepository init: %w", err)
	}
	if !foundHost {
		fmt.Print(sshConfigSetupMsg)
		return nil, fmt.Errorf("host setting %s not found in ssh config file", r.ConfigHost)
	}

	commandArg := []string{"init", "-r", fmt.Sprintf("sftp:%s:%s", r.ConfigHost, r.Destination)}
	output, err := execOutput(commandArg)
	if err != nil {
		return output, fmt.Errorf("sftpBackupRepository init: %w", err)
	}

	return output, nil
}

func (r SftpBackupRepository) Backup() error {
	os.Setenv(passwordEnv, r.Password)
	os.Setenv(resticProgressFPS, resticProgressFPSValue)

	// Check if ConfigHost setting exist in ssh config file
	foundHost, err := checkSshHost(r.ConfigHost)
	if err != nil {
		return fmt.Errorf("sftpBackupRepository snapshots: %w", err)
	}
	if !foundHost {
		fmt.Print(sshConfigSetupMsg)
		return fmt.Errorf("host setting %s not found in ssh config file", r.ConfigHost)
	}

	commandArg := []string{"backup", "-r", fmt.Sprintf("sftp:%s:%s", r.ConfigHost, r.Destination)}
	commandArg = append(commandArg, r.Sources...)

	// Add exclude option
	for _, exclude := range r.Excludes {
		excludeOpt := fmt.Sprintf("--exclude=%s", exclude)
		commandArg = append(commandArg, excludeOpt)
	}

	err = execStream(commandArg, true)
	if err != nil {
		return fmt.Errorf("sftpBackupRepository backup: %w", err)
	}

	// Check repository integrity and consistency after backup
	if err := r.Check(); err != nil {
		return fmt.Errorf("sftpBackupRepository backup: %w", err)
	}

	return nil
}

func (r SftpBackupRepository) Snapshots() ([]byte, error) {
	os.Setenv(passwordEnv, r.Password)

	// Check if ConfigHost setting exist in ssh config file
	foundHost, err := checkSshHost(r.ConfigHost)
	if err != nil {
		return nil, fmt.Errorf("sftpBackupRepository snapshots: %w", err)
	}
	if !foundHost {
		fmt.Print(sshConfigSetupMsg)
		return nil, fmt.Errorf("host setting %s not found in ssh config file", r.ConfigHost)
	}

	commandArg := []string{"snapshots", "-r", fmt.Sprintf("sftp:%s:%s", r.ConfigHost, r.Destination)}
	output, err := execOutput(commandArg)
	if err != nil {
		return output, fmt.Errorf("sftpBackupRepository snapshots: %w", err)
	}

	return output, nil
}

func (r SftpBackupRepository) Check() error {
	os.Setenv(passwordEnv, r.Password)

	// Check if ConfigHost setting exist in ssh config file
	foundHost, err := checkSshHost(r.ConfigHost)
	if err != nil {
		return fmt.Errorf("sftpBackupRepository check: %w", err)
	}
	if !foundHost {
		fmt.Print(sshConfigSetupMsg)
		return fmt.Errorf("host setting %s not found in ssh config file", r.ConfigHost)
	}

	commandArg := []string{"check", "-r", fmt.Sprintf("sftp:%s:%s", r.ConfigHost, r.Destination)}
	err = execStream(commandArg, false)
	if err != nil {
		return fmt.Errorf("sftpBackupRepository check: %w", err)
	}

	return nil
}

// checkSshHost find if configHost is set in user's ssh config file with syntax 'Host configHost'
// Return true is found, and return false otherwise
func checkSshHost(configHost string) (bool, error) {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("check ssh host: %w", err)
	}
	// Check ssh config file existence
	sshConfig := fmt.Sprintf("%s/.ssh/config", home)
	if _, err := os.Stat(sshConfig); errors.Is(err, fs.ErrNotExist) {
		return false, ErrSshConfigNotFound
	}

	// Check if input configHost exist in ssh config file
	searchHost := fmt.Sprintf("Host %s", configHost)
	sshConfigFile, err := os.Open(sshConfig)
	if err != nil {
		return false, fmt.Errorf("check ssh host: open file: %w", err)
	}
	defer sshConfigFile.Close()

	scanner := bufio.NewScanner(sshConfigFile)
	for scanner.Scan() {
		// line := scanner.Text()
		if strings.Contains(scanner.Text(), searchHost) {
			return true, nil
		}
	}

	return false, nil
}
