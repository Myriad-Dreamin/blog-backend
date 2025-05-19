package main

import (
	"github.com/Myriad-Dreamin/blog-backend/pkg/sqlite"
	"github.com/spf13/cobra"
)

var cmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "backup site",
	RunE: func(cmd *cobra.Command, args []string) error {
		return sqlite.BackupBlog(nil)
	},
}
