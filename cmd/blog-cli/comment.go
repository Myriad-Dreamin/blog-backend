package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/Myriad-Dreamin/blog-backend/pkg/commentmoderation"
	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
	"github.com/spf13/cobra"

	_ "github.com/mattn/go-sqlite3"
)

const defaultOwnerEmail = commentmoderation.DefaultOwnerEmail

type cliConfig struct {
	dataDir    string
	dbPath     string
	ownerEmail string
}

func newCommentCmd(cfg *cliConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Review and moderate article comments",
	}
	cmd.AddCommand(newCommentListCmd(cfg))
	cmd.AddCommand(newCommentShowCmd(cfg))
	cmd.AddCommand(newCommentDraftCmd(cfg))
	cmd.AddCommand(newCommentReviewCmd(cfg))
	cmd.AddCommand(newCommentAuthorizeCmd(cfg))
	cmd.AddCommand(newCommentRejectCmd(cfg))
	cmd.AddCommand(newCommentExportCmd(cfg))
	return cmd
}

func newCommentListCmd(cfg *cliConfig) *cobra.Command {
	var state string
	var format string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments by moderation state",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := cfg.openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			comments, err := commentmoderation.ListComments(cmd.Context(), db, commentmoderation.ListOptions{
				State: state,
				Limit: limit,
			})
			if err != nil {
				return err
			}

			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), comments)
			case "table":
				printCommentTable(cmd.OutOrStdout(), comments)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&state, "state", commentmoderation.StatePending, "comment state: pending, approved, or all")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum comments to list; use 0 for no limit")
	return cmd
}

func newCommentShowCmd(cfg *cliConfig) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := cfg.openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			comment, err := commentmoderation.GetComment(cmd.Context(), db, args[0])
			if err != nil {
				return err
			}

			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), comment)
			case "markdown":
				printCommentMarkdown(cmd.OutOrStdout(), comment)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown or json")
	return cmd
}

func newCommentDraftCmd(cfg *cliConfig) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "draft <id>",
		Short: "Generate owner and author notification drafts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := cfg.buildDraftBundle(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), bundle)
			case "markdown":
				printDraftMarkdown(cmd.OutOrStdout(), bundle)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown or json")
	return cmd
}

func newCommentReviewCmd(cfg *cliConfig) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "review <id>",
		Short: "Show a complete moderation packet for one comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := cfg.buildDraftBundle(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), bundle)
			case "markdown":
				printReviewMarkdown(cmd.OutOrStdout(), bundle)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown or json")
	return cmd
}

func newCommentAuthorizeCmd(cfg *cliConfig) *cobra.Command {
	return newCommentStateCmd(cfg, "authorize", true)
}

func newCommentRejectCmd(cfg *cliConfig) *cobra.Command {
	return newCommentStateCmd(cfg, "reject", false)
}

func newCommentStateCmd(cfg *cliConfig, name string, authorized bool) *cobra.Command {
	var format string
	var export bool

	cmd := &cobra.Command{
		Use:   name + " <id>",
		Short: commentStateShort(name, authorized),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := cfg.openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			id := args[0]
			if err := commentmoderation.SetCommentAuthorized(cmd.Context(), db, id, authorized); err != nil {
				return err
			}
			if export {
				if err := commentmoderation.ExportSnapshots(cmd.Context(), db, cfg.dataDir); err != nil {
					return err
				}
			}

			result := map[string]any{
				"id":         id,
				"authorized": authorized,
				"exported":   export,
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), result)
			case "text":
				state := "pending"
				if authorized {
					state = "approved"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "comment %s marked %s\n", id, state)
				if export {
					fmt.Fprintf(cmd.OutOrStdout(), "comment snapshots exported to %s\n", cfg.dataDir)
				}
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	cmd.Flags().BoolVar(&export, "export", true, "rewrite comment snapshot JSON files after updating")
	return cmd
}

func newCommentExportCmd(cfg *cliConfig) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export comment snapshot JSON files",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := cfg.openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			if err := commentmoderation.ExportSnapshots(cmd.Context(), db, cfg.dataDir); err != nil {
				return err
			}
			result := map[string]any{
				"dataDir": cfg.dataDir,
				"files": []string{
					filepath.Join(cfg.dataDir, "article-comments.json"),
					filepath.Join(cfg.dataDir, "article-email-comments.json"),
				},
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), result)
			case "text":
				fmt.Fprintf(cmd.OutOrStdout(), "comment snapshots exported to %s\n", cfg.dataDir)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	return cmd
}

func (cfg *cliConfig) openDB() (*sql.DB, error) {
	dbPath := cfg.dbPath
	if dbPath == "" {
		dbPath = filepath.Join(cfg.dataDir, "blog.db")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (cfg *cliConfig) buildDraftBundle(ctx context.Context, id string) (commentmoderation.DraftBundle, error) {
	db, err := cfg.openDB()
	if err != nil {
		return commentmoderation.DraftBundle{}, err
	}
	defer db.Close()

	comment, err := commentmoderation.GetComment(ctx, db, id)
	if err != nil {
		return commentmoderation.DraftBundle{}, err
	}
	comments, err := commentmoderation.ListComments(ctx, db, commentmoderation.ListOptions{State: commentmoderation.StateAll})
	if err != nil {
		return commentmoderation.DraftBundle{}, err
	}
	return commentmoderation.BuildDraftBundle(comment, comments, cfg.ownerEmail)
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func printCommentTable(w io.Writer, comments []dto.ArticleComment) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATE\tCREATED\tARTICLE\tEMAIL\tCONTENT")
	for _, comment := range comments {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			comment.Id,
			commentState(comment),
			commentmoderation.FormatCreatedAt(comment.CreatedAt),
			comment.ArticleId,
			oneLine(comment.Email, 36),
			oneLine(comment.Content, 96),
		)
	}
	tw.Flush()
}

func printCommentMarkdown(w io.Writer, comment dto.ArticleComment) {
	fmt.Fprintf(w, "# Comment %s\n\n", comment.Id)
	fmt.Fprintf(w, "- State: %s\n", commentState(comment))
	fmt.Fprintf(w, "- Article: %s\n", comment.ArticleId)
	fmt.Fprintf(w, "- Created: %s\n", commentmoderation.FormatCreatedAt(comment.CreatedAt))
	fmt.Fprintf(w, "- Email: %s\n\n", comment.Email)
	fmt.Fprintln(w, "## Content")
	fmt.Fprintln(w)
	fmt.Fprintln(w, fencedText(comment.Content))
}

func printReviewMarkdown(w io.Writer, bundle commentmoderation.DraftBundle) {
	printCommentMarkdown(w, bundle.Comment)
	fmt.Fprintln(w)
	printDraftMarkdown(w, bundle)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "## Decision Commands")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Approve: `target/blog-cli comment authorize %s`\n", bundle.Comment.Id)
	fmt.Fprintf(w, "- Keep pending: `target/blog-cli comment reject %s`\n", bundle.Comment.Id)
}

func printDraftMarkdown(w io.Writer, bundle commentmoderation.DraftBundle) {
	fmt.Fprintln(w, "## Links")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Canonical: %s\n", bundle.Links.Canonical)
	fmt.Fprintf(w, "- CN: %s\n", bundle.Links.CN)
	fmt.Fprintln(w)
	printEmailDraftMarkdown(w, "Owner Notification", bundle.Owner)
	fmt.Fprintln(w)
	printEmailDraftMarkdown(w, "Author Authorization", bundle.Author)
}

func printEmailDraftMarkdown(w io.Writer, title string, draft commentmoderation.EmailDraft) {
	fmt.Fprintf(w, "## %s\n\n", title)
	fmt.Fprintf(w, "- To: %s\n", draft.To)
	if len(draft.Bcc) > 0 {
		fmt.Fprintf(w, "- Bcc: %s\n", strings.Join(draft.Bcc, ", "))
	}
	fmt.Fprintf(w, "- Subject: %s\n", draft.Subject)
	fmt.Fprintf(w, "- Gmail: %s\n\n", draft.GmailURL)
	fmt.Fprintln(w, "```text")
	fmt.Fprintln(w, draft.Body)
	fmt.Fprintln(w, "```")
}

func commentState(comment dto.ArticleComment) string {
	if comment.Authorized {
		return "approved"
	}
	return "pending"
}

func commentStateShort(name string, authorized bool) string {
	if authorized {
		return "Mark one comment as approved"
	}
	if name == "reject" {
		return "Keep one comment unapproved without deleting it"
	}
	return "Update one comment state"
}

func fencedText(text string) string {
	if strings.Contains(text, "```") {
		return "````text\n" + text + "\n````"
	}
	return "```text\n" + text + "\n```"
}

func oneLine(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	if limit <= 0 || len(text) <= limit {
		return text
	}
	if limit <= 3 {
		return text[:limit]
	}
	return text[:limit-3] + "..."
}

func init() {
	cobra.EnableCommandSorting = false
}
