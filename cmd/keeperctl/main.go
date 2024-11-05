/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import "GophKeeper/cmd/keeperctl/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
