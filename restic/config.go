package restic

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
)

type Config struct {
	Repository struct {
		Password string `yaml:"password"`
	} `yaml:"repository"`
	Backups []Backup `yaml:"backups"`
}

type BackupTypeConfig interface {
	Validate() error
	String() string
}

type Backup struct {
	Name   string           `yaml:"name"`
	Type   string           `yaml:"type"`
	Config BackupTypeConfig `yaml:"config"`
}

type LocalBackupConfig struct {
	Sources     []string `yaml:"sources"`
	Destination string   `yaml:"destination"`
}

func (c LocalBackupConfig) Validate() error {
	return nil
}

func (c LocalBackupConfig) String() string {
	var builder strings.Builder
	builder.WriteString(srcDestString(c.Sources, c.Destination))

	return builder.String()
}

type SftpBackupConfig struct {
	Host        string   `yaml:"host"`
	Sources     []string `yaml:"sources"`
	Destination string   `yaml:"destination"`
}

func (c SftpBackupConfig) Validate() error {
	return nil
}

func (c SftpBackupConfig) String() string {
	var builder strings.Builder
	builder.WriteString(srcDestString(c.Sources, c.Destination))
	builder.WriteString(fmt.Sprintf("Host: %s\n", c.Host))

	return builder.String()
}

type S3BackupConfig struct {
	AccessKeyId     string   `yaml:"accessKeyId"`
	SecretAccessKey string   `yaml:"secretAccessKey"`
	Region          string   `yaml:"region"`
	Sources         []string `yaml:"sources"`
	Destination     string   `yaml:"destination"`
}

func (c S3BackupConfig) Validate() error {
	return nil
}

func (c S3BackupConfig) String() string {
	var builder strings.Builder
	builder.WriteString(srcDestString(c.Sources, c.Destination))
	builder.WriteString(fmt.Sprintf("Access Key ID: %s\n", c.AccessKeyId))
	builder.WriteString(fmt.Sprintf("Secret Access Key: %s\n", c.SecretAccessKey))
	builder.WriteString(fmt.Sprintf("Region: %s\n", c.Region))

	return builder.String()
}

func srcDestString(sources []string, destination string) string {
	var builder strings.Builder
	builder.WriteString("Sources:\n")
	for _, source := range sources {
		builder.WriteString(fmt.Sprintf("- %s\n", source))
	}
	builder.WriteString(fmt.Sprintf("Destination: %s\n", destination))

	return builder.String()
}

func NewConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	var rawConfig struct {
		Repository struct {
			Password string `yaml:"password"`
		} `yaml:"repository"`
		Backups []struct {
			Name   string                 `yaml:"name"`
			Type   string                 `yaml:"type"`
			Config map[string]interface{} `yaml:"config"`
		} `yaml:"backups"`
	}

	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("new config: %w", err)
	}

	// Process raw backup configuration
	config := Config{Repository: rawConfig.Repository}
	for _, rawBackup := range rawConfig.Backups {
		var typedConfig BackupTypeConfig

		switch rawBackup.Type {
		case "local":
			typedConfig = &LocalBackupConfig{}
		case "sftp":
			typedConfig = &SftpBackupConfig{}
		case "s3":
			typedConfig = &S3BackupConfig{}
		default:
			return nil, fmt.Errorf("new config: unsupportedt type %s", rawBackup.Type)
		}

		configBytes, err := yaml.Marshal(rawBackup.Config)
		if err != nil {
			return nil, fmt.Errorf("new config: remarshal: %w", err)
		}

		// Re-marshal config
		err = yaml.Unmarshal(configBytes, typedConfig)
		if err != nil {
			return nil, fmt.Errorf("new config: remarshal: %w", err)
		}

		config.Backups = append(config.Backups, Backup{
			Name:   rawBackup.Name,
			Type:   rawBackup.Type,
			Config: typedConfig,
		})

	}

	return &config, nil
}

// type Config struct {
// 	Repository struct {
// 		Password string `yaml:"password"`
// 	} `yaml:"repository"`
// 	Backups []struct {
// 		Name        string   `yaml:"name"`
// 		Type        string   `yaml:"type"`
// 		Host        string   `yaml:"host,omitempty"`
// 		Sources     []string `yaml:"sources"`
// 		Destination string   `yaml:"destination"`
// 		AccessKeyId string   `yaml:"accessKeyId,omitempty"`
// 		SecretKey   string   `yaml:"secretAccessKey,omitempty"`
// 		Region      string   `yaml:"region,omitempty"`
// 	} `yaml:"backups"`
// }

// // NewConfig load data into struct from given config file
// func NewConfig(filepath string) (*Config, error) {
// 	data, err := os.ReadFile(filepath)
// 	if err != nil {
// 		return nil, ErrConfigNotFound
// 	}

// 	var config Config
// 	if err := yaml.Unmarshal(data, &config); err != nil {
// 		return nil, fmt.Errorf("new config: %w", err)
// 	}

// 	return &config, nil
// }
