/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"GophKeeper/cmd/keeperctl/internal/client"
	"GophKeeper/cmd/keeperctl/internal/utils"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// listFilesCmd represents the listFiles command
var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List files uploaded by a specified user",
	Long: `The list-user-files command allows you to list all the files that have been uploaded by a specific user.
You need to provide the username to fetch the list of their uploaded files.

Usage:
  list-user-files [flags]

Flags:
  -u, --user string   Username whose uploaded files you want to list

Example:
  list-user-files --user "admin"

The command performs the following steps:
1. Retrieves the username from the provided flag or configuration.
2. Connects to the server to fetch the list of files uploaded by the specified user.
3. Displays the list of uploaded files.
`,
	Run: func(cmd *cobra.Command, args []string) {
		address := viper.GetString("listen_address")
		username := viper.GetString("user")
		fmClient := client.NewFMClient(address)
		if utils.LoginCycle(username, fmClient) {
			ctx := context.Background()
			md := metadata.New(map[string]string{"authorization": fmClient.CashedToken})
			ctx = metadata.NewOutgoingContext(ctx, md)
			files, err := fmClient.ListUserFiles(ctx)
			if err != nil {
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
			fmt.Fprintln(w, "NAME\tSIZE (bytes)\tVERSION")
			for _, object := range files.Objects {
				if object.FileName == "" {
					continue
				}
				fmt.Fprintf(w, "%s\t%d\t%s\n", object.FileName, object.Size, object.VersionID)
			}

			w.Flush()

		}
	},
}

func init() {
	rootCmd.AddCommand(listFilesCmd)
}
