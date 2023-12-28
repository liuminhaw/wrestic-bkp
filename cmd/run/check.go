/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package run

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

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check BackupName",
	Short: "Checking repository integrity and consistency",
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
			log.Fatalf("repository check: %v\n", err)
		}
		checkConf, err := config.ReadBackup(backupName)
		if err != nil {
			if errors.Is(err, restic.ErrConfigBackupNameNotFound) {
				fmt.Printf("backup name %s not found in config file\n", backupName)
				os.Exit(1)
			}
			log.Fatalf("repository check: %v\n", err)
		}

		backupRepo, err := config.CreateRepositoryStruct(checkConf.Config)
		if err != nil {
			log.Fatalf("repository check :%v\n", err)
		}
		if err := backupRepo.Check(); err != nil {
			log.Fatalf("repository check: %v\n", err)
		}
	},
}

func init() {
	RunCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
