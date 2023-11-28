/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package run

import (
	"fmt"
	"os"

	"github.com/liuminhaw/wrestic-bkp/restic"
	"github.com/spf13/cobra"
)

// repositoryCmd represents the repository command
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Restic execution",
	Long:  ``,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("repository called")
	// },
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// repositoryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// repositoryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// requirementsCheck check needed requirements for program execution.
// Exit if any requirment is not met
func requirementsCheck() {
	if err := restic.ResticCheck(); err != nil {
		fmt.Println("restic should be installed before running this program")
		os.Exit(1)
	}
}
