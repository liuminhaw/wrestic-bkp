/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package test

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	testConfig string = "tests/config.test.yaml"
)

// testCmd represents the test command
var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "This is for testing purpose",
	Long:  ``,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("test called")
	// },
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// readConfig read separate config file 'testConfig' for testing
func readConfig() {
	viper.SetConfigFile(testConfig)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Cannot read test config file: %s\n", testConfig)
	}
}
