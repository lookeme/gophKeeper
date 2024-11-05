/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"GophKeeper/cmd/keeperctl/internal/client"
	"GophKeeper/cmd/keeperctl/internal/utils"
	"GophKeeper/internal/proto/gkeeper/pb"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"os"
)

// downloadCmd represents the dowload command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads a file from the server",
	Long: `Downloads a specified file from the server using the given username for authentication.
The file will be saved to the specified local path.

Parameters:
- username: The username of the account used to authenticate the download request.
- filename: The name of the file to be downloaded.
- path: The local file system path where the downloaded file should be saved.

Examples:
  # Download a file using username 'johndoe' and save it to '/tmp/myfile.txt'
  gophkeeper download --username johndoe --filename myfile.txt --path /tmp/myfile.txt
`,
	Run: func(cmd *cobra.Command, args []string) {
		username := viper.GetString("user")
		address := viper.GetString("listen_address")
		file := viper.GetString("filename")
		path := viper.GetString("path")
		if path == "" {
			fmt.Println("filepath is required")
			return
		}
		fmClient := client.NewFMClient(address)
		if utils.LoginCycle(username, fmClient) {
			err := downloadFile(fmClient, file, path)
			if err != nil {
				_ = fmt.Errorf("error downloading file: %v", err)
				return
			}

		}
		defer fmClient.Close()
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().StringP("filename", "f", "", "file to download")
	viper.BindPFlag("filename", downloadCmd.PersistentFlags().Lookup("filename"))
}

func downloadFile(fmClient *client.FileManagerClient, filename string, path string) error {
	req := &pb.DownloadRequest{
		Filename: filename,
	}
	md := metadata.New(map[string]string{"authorization": fmClient.CashedToken})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := fmClient.DownloadFile(ctx, req)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for {
		resp, recErr := stream.Recv()
		if recErr == io.EOF {
			log.Println("Download complete.")
			break
		}
		if recErr != nil {
			return recErr
		}

		if _, writeErr := file.Write(resp.Chunk); writeErr != nil {
			return writeErr
		}
	}
	return nil
}
