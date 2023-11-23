/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package repository

import (
	"github.com/spf13/cobra"
)

// repositoryCmd represents the repository command
var RepositoryCmd = &cobra.Command{
	Use:   "repository",
	Short: "Execute actions on restic repository",
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
