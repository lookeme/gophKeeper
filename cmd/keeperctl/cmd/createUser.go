/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"GophKeeper/cmd/keeperctl/internal/client"
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// createUserCmd represents the createUser command
var createUserCmd = &cobra.Command{
	Use:   "createUser",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("createUser called")
		address := viper.GetString("listen_address")
		fmClient := client.NewFMClient(address)
		defer fmClient.Close()
		reader := bufio.NewReader(os.Stdin)
		registerUser(reader, fmClient)
	},
}

func init() {
	rootCmd.AddCommand(createUserCmd)
}

func registerUser(reader *bufio.Reader, fmtClient *client.FileManagerClient) bool {
	for {
		fmt.Print("Enter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		fmt.Print("Enter your password: ")
		// Read the password without echoing
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			continue
		}
		password := string(passwordBytes)
		fmt.Println()
		if password == "" {
			continue
		}
		fmt.Print("Enter email: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		user, err := fmtClient.CreateUser(username, password, email)
		if err != nil {
			fmt.Print("error during registration: " + err.Error())
			break
		}
		fmt.Println("User " + user + " created")
		return true
	}
	return false
}
