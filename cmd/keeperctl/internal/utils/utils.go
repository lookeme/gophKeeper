package utils

import (
	"GophKeeper/cmd/keeperctl/internal/client"
	"fmt"
	"golang.org/x/term"
	"strings"
	"syscall"
)

func LoginCycle(username string, fmClient *client.FileManagerClient) bool {
	for {
		fmt.Print("Enter password: ")
		bytePassword, _ := term.ReadPassword(syscall.Stdin)
		fmt.Println()
		password := strings.TrimSpace(string(bytePassword))
		err := fmClient.GetAuthToken(username, password)
		if err != nil {
			fmt.Print("error during authentication: " + err.Error() + "\n")
			continue
		}
		fmt.Println("You are logged in as " + username)
		return true
	}
}
