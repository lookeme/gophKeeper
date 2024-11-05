/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"GophKeeper/internal/version"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "gkeeper",
	Short:   "Server side app for storing files",
	Version: version.Version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gkeeper v: " + version.Version)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "Конфигурационный файл (по умолчанию - config.yaml)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}
