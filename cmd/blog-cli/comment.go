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

const (
	commentSourceSnapshot = "snapshot"
	commentSourceDB       = "db"
)

type cliConfig struct {
	dataDir      string
	dbPath       string
	commentsPath string
	ownerEmail   string
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
	var source string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments by moderation state",
		RunE: func(cmd *cobra.Command, args []string) error {
			comments, err := cfg.listComments(cmd.Context(), source, commentmoderation.ListOptions{
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
				cfg.writeSnapshotRefreshReminder(cmd, source, format)
				printCommentTable(cmd.OutOrStdout(), comments)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&state, "state", commentmoderation.StatePending, "comment state: pending, approved, rejected, or all")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 20, "maximum comments to list; use 0 for no limit")
	cmd.Flags().StringVar(&source, "source", commentSourceSnapshot, "comment source: snapshot or db")
	return cmd
}

func newCommentShowCmd(cfg *cliConfig) *cobra.Command {
	var format string
	var source string

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show one comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			comment, err := cfg.getComment(cmd.Context(), source, args[0])
			if err != nil {
				return err
			}

			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), comment)
			case "markdown":
				cfg.writeSnapshotRefreshReminder(cmd, source, format)
				printCommentMarkdown(cmd.OutOrStdout(), comment)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown or json")
	cmd.Flags().StringVar(&source, "source", commentSourceSnapshot, "comment source: snapshot or db")
	return cmd
}

func newCommentDraftCmd(cfg *cliConfig) *cobra.Command {
	var format string
	var source string

	cmd := &cobra.Command{
		Use:   "draft <id>",
		Short: "Generate owner and author notification drafts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := cfg.buildDraftBundle(cmd.Context(), source, args[0])
			if err != nil {
				return err
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), bundle)
			case "markdown":
				cfg.writeSnapshotRefreshReminder(cmd, source, format)
				printDraftMarkdown(cmd.OutOrStdout(), bundle)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown or json")
	cmd.Flags().StringVar(&source, "source", commentSourceSnapshot, "comment source: snapshot or db")
	return cmd
}

func newCommentReviewCmd(cfg *cliConfig) *cobra.Command {
	var format string
	var source string

	cmd := &cobra.Command{
		Use:   "review <id>",
		Short: "Show a complete moderation packet for one comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := cfg.buildDraftBundle(cmd.Context(), source, args[0])
			if err != nil {
				return err
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), bundle)
			case "markdown":
				cfg.writeSnapshotRefreshReminder(cmd, source, format)
				printReviewMarkdown(cmd.OutOrStdout(), bundle)
				return nil
			default:
				return fmt.Errorf("unknown format %q", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown or json")
	cmd.Flags().StringVar(&source, "source", commentSourceSnapshot, "comment source: snapshot or db")
	return cmd
}

func newCommentAuthorizeCmd(cfg *cliConfig) *cobra.Command {
	return newCommentStateCmd(cfg, "authorize", commentmoderation.StateApproved)
}

func newCommentRejectCmd(cfg *cliConfig) *cobra.Command {
	return newCommentStateCmd(cfg, "reject", commentmoderation.StateRejected)
}

func newCommentStateCmd(cfg *cliConfig, name string, state string) *cobra.Command {
	var format string
	var export bool

	cmd := &cobra.Command{
		Use:   name + " <id>",
		Short: commentStateShort(name, state),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := cfg.openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			id := args[0]
			if err := commentmoderation.SetCommentState(cmd.Context(), db, id, state); err != nil {
				return err
			}
			if export {
				if err := commentmoderation.ExportSnapshots(cmd.Context(), db, cfg.dataDir); err != nil {
					return err
				}
			}

			result := map[string]any{
				"id":       id,
				"state":    state,
				"exported": export,
			}
			switch format {
			case "json":
				return writeJSON(cmd.OutOrStdout(), result)
			case "text":
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
	if err := commentmoderation.EnsureRejectedColumn(context.Background(), db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (cfg *cliConfig) listComments(ctx context.Context, source string, opts commentmoderation.ListOptions) ([]dto.ArticleComment, error) {
	source, err := normalizeCommentSource(source)
	if err != nil {
		return nil, err
	}

	switch source {
	case commentSourceSnapshot:
		comments, err := cfg.loadSnapshotComments()
		if err != nil {
			return nil, err
		}
		return commentmoderation.ListCommentsFromSlice(comments, opts)
	case commentSourceDB:
		db, err := cfg.openDB()
		if err != nil {
			return nil, err
		}
		defer db.Close()
		return commentmoderation.ListComments(ctx, db, opts)
	default:
		return nil, fmt.Errorf("unknown comment source %q", source)
	}
}

func (cfg *cliConfig) getComment(ctx context.Context, source string, id string) (dto.ArticleComment, error) {
	source, err := normalizeCommentSource(source)
	if err != nil {
		return dto.ArticleComment{}, err
	}

	switch source {
	case commentSourceSnapshot:
		comments, err := cfg.loadSnapshotComments()
		if err != nil {
			return dto.ArticleComment{}, err
		}
		return commentmoderation.GetCommentFromSlice(comments, id)
	case commentSourceDB:
		db, err := cfg.openDB()
		if err != nil {
			return dto.ArticleComment{}, err
		}
		defer db.Close()
		return commentmoderation.GetComment(ctx, db, id)
	default:
		return dto.ArticleComment{}, fmt.Errorf("unknown comment source %q", source)
	}
}

func (cfg *cliConfig) buildDraftBundle(ctx context.Context, source string, id string) (commentmoderation.DraftBundle, error) {
	comment, err := cfg.getComment(ctx, source, id)
	if err != nil {
		return commentmoderation.DraftBundle{}, err
	}
	comments, err := cfg.listComments(ctx, source, commentmoderation.ListOptions{State: commentmoderation.StateAll})
	if err != nil {
		return commentmoderation.DraftBundle{}, err
	}
	return commentmoderation.BuildDraftBundle(comment, comments, cfg.ownerEmail)
}

func (cfg *cliConfig) loadSnapshotComments() ([]dto.ArticleComment, error) {
	return commentmoderation.LoadSnapshotComments(commentmoderation.SnapshotOptions{
		PublicPath: cfg.publicCommentsPath(),
		EmailPath:  filepath.Join(cfg.dataDir, commentmoderation.EmailCommentsFilename),
	})
}

func (cfg *cliConfig) publicCommentsPath() string {
	if cfg.commentsPath != "" {
		return cfg.commentsPath
	}
	return filepath.Join(cfg.dataDir, commentmoderation.PublicCommentsFilename)
}

func (cfg *cliConfig) writeSnapshotRefreshReminder(cmd *cobra.Command, source string, format string) {
	if strings.EqualFold(format, "json") {
		return
	}
	source, err := normalizeCommentSource(source)
	if err != nil || source != commentSourceSnapshot {
		return
	}
	fmt.Fprintf(
		cmd.ErrOrStderr(),
		"Reminder: refresh the local comment snapshot before reviewing: make download-data (source: %s)\n\n",
		cfg.publicCommentsPath(),
	)
}

func normalizeCommentSource(source string) (string, error) {
	source = strings.ToLower(strings.TrimSpace(source))
	switch source {
	case "", commentSourceSnapshot:
		return commentSourceSnapshot, nil
	case commentSourceDB:
		return commentSourceDB, nil
	default:
		return "", fmt.Errorf("unknown comment source %q", source)
	}
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
	fmt.Fprintf(w, "- Approve on cloud DB: `make comment-authorize ID=%s`\n", bundle.Comment.Id)
	fmt.Fprintf(w, "- Reject on cloud DB: `make comment-reject ID=%s`\n", bundle.Comment.Id)
	fmt.Fprintln(w, "- Refresh local snapshot after a decision: `make download-data`")
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
	if comment.Rejected {
		return "rejected"
	}
	if comment.Authorized {
		return "approved"
	}
	return "pending"
}

func commentStateShort(name string, state string) string {
	if state == commentmoderation.StateApproved {
		return "Mark one comment as approved"
	}
	if name == "reject" {
		return "Mark one comment as rejected without deleting it"
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
