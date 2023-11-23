/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"fmt"
	"strings"

	"github.com/liuminhaw/wrestic-bkp/restic"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "command related to config file",
	Long:  ``,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("config called")
	// },
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ValidConfigNames(config *restic.Config) string {
	var builder strings.Builder
	builder.WriteString("[ ")
	for _, backup := range config.Backups {
		builder.WriteString(fmt.Sprintf("'%s' ", backup.Name))
	}
	builder.WriteString("]")
	return builder.String()
}
