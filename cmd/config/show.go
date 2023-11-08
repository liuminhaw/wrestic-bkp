/*
Copyright © 2023 Min-Haw, Liu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/liuminhaw/wrestic-bkp/restic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	ErrCommandNotFound = errors.New("Command not found")
)

// checkCmd represents the check command
var showCmd = &cobra.Command{
	Use:   "show [backup name]",
	Short: "show configuration",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		// Optionally run one of the validators provided by cobra
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}
		// Run the custom validation logic
		if len(args) == 0 {
			return nil
		}

		config, err := restic.NewConfig(viper.ConfigFileUsed())
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if isValidConfigName(config, args[0]) {
			return nil
		}

		return fmt.Errorf("given name '%s' not found in config backup names: %v", args[0], validConfigNames(config))
	},
	Run: func(cmd *cobra.Command, args []string) {
		var backupName string
		if len(args) == 1 {
			backupName = args[0]
		}

		err := CmdCheck("restic")
		if err != nil {
			fmt.Println("restic should be installed before using this program")
			os.Exit(1)
		}

		backups, err := restic.NewConfig(viper.ConfigFileUsed())
		if err != nil {
			log.Fatalf("check: %v\n", err)
		}
		// Now you can use the config struct, for example, print the backup names
		for _, backup := range backups.Backups {
			data, err := yaml.Marshal(backup)
			if err != nil {
				log.Fatalf("config check: %v\n", err)
			}
			if backupName == "" || backupName == backup.Name {
				fmt.Printf("--- config backup:\n%s\n\n", string(data))
			}
			// fmt.Printf("Name: %s\n", backup.Name)
			// fmt.Printf("Type: %s\n", backup.Type)
			// fmt.Println(backup.Config.String())
		}
	},
}

func init() {
	ConfigCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// CmdCheck checks if given command is available in system PATH
func CmdCheck(command string) error {
	_, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("command not found: %s", command)
	}

	return nil
}

func isValidConfigName(config *restic.Config, name string) bool {
	for _, backup := range config.Backups {
		if backup.Name == name {
			return true
		}
	}
	return false
}

func validConfigNames(config *restic.Config) string {
	var builder strings.Builder
	builder.WriteString("[ ")
	for _, backup := range config.Backups {
		builder.WriteString(fmt.Sprintf("'%s' ", backup.Name))
	}
	builder.WriteString("]")
	return builder.String()
}
