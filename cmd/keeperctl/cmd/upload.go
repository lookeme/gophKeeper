/*
Copyright © 2024 NAME HERE <loolmeup42@gmail.com>
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
	"path/filepath"
)

const smallFileChunk = 1024 * 1024

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to the server",
	Long: `The upload command allows you to upload a specified file to a server.
You need to provide the server address, username, and the path to the file you wish to upload.
Ensure that these parameters are configured correctly in your environment.

Usage:
  upload [flags]

Flags:
  -a, --address string   Server address to connect to
  -u, --user string      Username for authentication
  -p, --path string      Path to the file to upload

Example:
  upload --address "localhost:8080" --user "admin" --path "/path/to/file.txt"
  upload --user tester --path ./large2.pptx
  otherwise the command will use the values configured in the yaml config file.

The command performs the following steps:
1. Establishes a connection to the server.
2. Authenticates the user.
3. Uploads the specified file to the server.
4. Closes the connection after completion.`,
	Run: func(cmd *cobra.Command, args []string) {
		address := viper.GetString("listen_address")
		username := viper.GetString("user")
		path := viper.GetString("path")
		if path == "" {
			fmt.Println("filepath is required")
			return
		}
		fmClient := client.NewFMClient(address)
		if utils.LoginCycle(username, fmClient) {
			uploadFile(path, fmClient)

		}
		defer fmClient.Close()
	},
}

func uploadFile(filePath string, fmClient *client.FileManagerClient) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
		return
	}
	defer file.Close()
	md := metadata.New(map[string]string{"authorization": fmClient.CashedToken})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stat, _ := file.Stat()
	size := stat.Size()
	stream, err := fmClient.UploadFile(ctx)
	if err != nil {
		log.Fatalf("failed to create upload stream: %v", err)
	}
	// Send file in chunks
	buf := make([]byte, smallFileChunk)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatalf("error reading file: %v", err)
	}
	fileName := filepath.Base(filePath)
	err = stream.Send(&pb.FileChunk{
		ChunkSize: smallFileChunk,
		Chunk:     buf[:n],
		Filename:  fileName,
	})
	if err != nil {
		log.Fatalf("error sending chunk: %v", err)
	}
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading file: %v", err)
		}
		err = stream.Send(&pb.FileChunk{
			Filename: fileName,
			FileSize: size,
			Chunk:    buf[:n]})
		if err != nil {
			log.Fatalf("error sending chunk: %v", err)
		}
	}
	// Close and receive the server's response
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	log.Printf("Upload Status: %v", res.Message)
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
