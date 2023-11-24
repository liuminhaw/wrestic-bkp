/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package repository

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

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup BackupName",
	Short: "Backup from source paths to repository referencing BackupName configuration",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

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

		requirementsCheck()

		config, err := restic.NewConfig(viper.ConfigFileUsed())
		if err != nil {
			log.Fatalf("repository backup: %v\n", err)
		}
		backupConf, err := config.ReadBackup(backupName)
		if err != nil {
			if errors.Is(err, restic.ErrConfigBackupNameNotFound) {
				fmt.Printf("backup %s not found in config file\n", backupName)
				os.Exit(1)
			}
			log.Fatalf("repository init: %v\n", err)
		}

		backupRepo, err := config.CreateRepositoryStruct(backupConf.Config)
		if err != nil {
			log.Fatalf("repository backup: %v\n", err)
		}
		if err := backupRepo.Backup(); err != nil {
			log.Fatalf("repository backup: %v\n", err)
		}
	},
}

func init() {
	RepositoryCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
