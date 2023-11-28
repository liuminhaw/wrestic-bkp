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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init BackupName",
	Short: "Init repository for BackupName",
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
			log.Fatalf("repository init: %v\n", err)
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
			log.Fatalf("repository init: %v\n", err)
		}
		output, err := backupRepo.Init()
		if err != nil {
			fmt.Print(string(output))
			fmt.Printf("wrestic init: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func init() {
	RunCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
