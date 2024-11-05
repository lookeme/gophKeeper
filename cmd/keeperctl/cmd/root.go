// Package cmd /*
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "My CLI Application",
	Long: `My CLI Application is a versatile tool for interacting with server functionalities.
It allows users to perform various tasks such as uploading, downloading, and encrypting files, 
as well as listing files and their different versions.

The CLI is designed to be easy to use and highly configurable, ensuring efficient task completion.

Usage:
  gophkeeper [command]

Available Commands:
  upload        Upload a file to the server
  download      Download a file from the server
  encrypt       Encrypt a specified file
  list-files    List all files on the server
  list-versions List different versions of a specified file
  help          Help about any command

Flags:
  -h, --help     Show help for the root command
  -v, --version  Display the version of the CLI application

Use "gophkeeper [command] --help" for more information about a command.

Examples:
  gophkeeper upload --address "localhost:8080" --user "admin" --path "/path/to/file.txt"
  gophkeeper download --address "localhost:8080" --user "admin" --file "file.txt"
  gophkeeper encrypt --path "/path/to/file.txt"
  gophkeeper list-files
  gophkeeper list-versions --file "file.txt"
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("App Name:", viper.GetString("listen_address"))
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().String("config", "", "config file (default is ./config.yaml)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	rootCmd.PersistentFlags().String("user", "", "provide user name")
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("user"))
	rootCmd.PersistentFlags().String("path", "", "path to file to upload")
	viper.BindPFlag("path", rootCmd.PersistentFlags().Lookup("path"))
}
func initConfig() {
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	//viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
	fmt.Println("Using config file:", viper.ConfigFileUsed())
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
