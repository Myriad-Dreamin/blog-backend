package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cfg := &cliConfig{}

	rootCmd := &cobra.Command{
		Use:          "blog-cli",
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().StringVar(&cfg.dataDir, "data-dir", "./.data", "directory containing blog data files")
	rootCmd.PersistentFlags().StringVar(&cfg.dbPath, "db", "", "path to blog sqlite database (defaults to <data-dir>/blog.db)")
	rootCmd.PersistentFlags().StringVar(&cfg.ownerEmail, "owner-email", defaultOwnerEmail, "comment notification owner email")

	rootCmd.AddCommand(newCommentCmd(cfg))
	return rootCmd
}
