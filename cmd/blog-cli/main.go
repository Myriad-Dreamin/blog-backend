package main

import (
	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var rootCmd = &cobra.Command{Use: "blog-cli"}
	rootCmd.AddCommand(cmdClick)
	rootCmd.AddCommand(cmdComment)
	rootCmd.AddCommand(cmdBackup)
	rootCmd.Execute()
}
