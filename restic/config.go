package restic

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrConfigNotFound           = errors.New("config file not found")
	ErrConfigBackupNameNotFound = errors.New("backup name in config not found")
)

type Config struct {
	Repository ConfigRepository `yaml:"repository"`
	Backups    []Backup         `yaml:"backups"`
}

type BackupTypeConfig interface {
	Validate() error
	String() string
}

type ConfigRepository struct {
	Password string `yaml:"password"`
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

func CreateRepositoryStruct(cRepo ConfigRepository, bConf BackupTypeConfig) (ResticRepository, error) {
	switch v := bConf.(type) {
	case *LocalBackupConfig:
		return LocalBackupRepository{
			Password:    cRepo.Password,
			Destination: v.Destination,
		}, nil
	default:
		fmt.Printf("type of bConf: %T\n", v)
		return nil, errors.New("no matched concrete type")
	}
}

// NewConfig read in file and return restic Config type struct
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

// ReadBackup find Backup struct with given name and returns it.
// Return ErrConfigBackupNameNotFound error if no matching name found in config
func (c *Config) ReadBackup(name string) (Backup, error) {
	for _, backup := range c.Backups {
		if backup.Name == name {
			return backup, nil
		}
	}

	return Backup{}, ErrConfigBackupNameNotFound
}

func (c *Config) BackupNames() []string {
	names := []string{}
	for _, backup := range c.Backups {
		names = append(names, backup.Name)
	}

	return names
}

func (c *Config) IsValidName(name string) bool {
	for _, backup := range c.Backups {
		if backup.Name == name {
			return true
		}
	}
	return false
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
