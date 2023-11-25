/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package test

import (
	"errors"
	"fmt"
	"log"
	"os"

	conf "github.com/liuminhaw/wrestic-bkp/cmd/config"
	"github.com/liuminhaw/wrestic-bkp/restic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate BackupName",
	Short: "Generate testing files for BackupName",
	Long: `Read settings from BackupName configuration and generate random
source files and destination directory for testing`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		readConfig()

		config, err := restic.NewConfig(viper.ConfigFileUsed())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if !config.IsValidName(args[0]) {
			return fmt.Errorf("given name '%s' not found in config names: %v", args[0], conf.ValidConfigNames(config))
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		backupName := args[0]

		config, err := restic.NewConfig(viper.ConfigFileUsed())
		if err != nil {
			log.Fatalf("test generate: load config: %v\n", err)
		}
		backupConf, err := config.ReadBackup(backupName)
		if err != nil {
			if errors.Is(err, restic.ErrConfigBackupNameNotFound) {
				fmt.Printf("backup %s not found in config file\n", backupName)
				os.Exit(1)
			}
			log.Fatalf("test generate: read backupName: %v\n", err)
		}

		backupTest, err := config.CreateTestStruct(backupConf.Config)
		if err != nil {
			log.Fatalf("test generate: create repo struct: %v\n", err)
		}
		if err := backupTest.TestGenerate(); err != nil {
			log.Fatalf("test generate: %v\n", err)
		}
	},
}

func init() {
	TestCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
