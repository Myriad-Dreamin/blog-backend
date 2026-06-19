package commentmoderation

import (
	"context"
	"database/sql"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"
	"github.com/Myriad-Dreamin/blog-backend/pkg/iou"
)

const (
	StateAll      = "all"
	StatePending  = "pending"
	StateApproved = "approved"

	DefaultOwnerEmail = "Kamiya <camiyoru@gmail.com>"
)

var (
	commentMentionRE = regexp.MustCompile(`\[comment:(.+?)\]`)
	userMentionRE    = regexp.MustCompile(`\[user:(.+?)\]`)
)

type ListOptions struct {
	State string
	Limit int
}

type CommentLinks struct {
	Canonical string `json:"canonical"`
	CN        string `json:"cn"`
}

type EmailDraft struct {
	To       string   `json:"to"`
	Bcc      []string `json:"bcc,omitempty"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	GmailURL string   `json:"gmailUrl"`
}

type DraftBundle struct {
	Comment dto.ArticleComment `json:"comment"`
	Links   CommentLinks       `json:"links"`
	Owner   EmailDraft         `json:"owner"`
	Author  EmailDraft         `json:"author"`
}

func NormalizeState(state string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "", StatePending:
		return StatePending, nil
	case StateAll:
		return StateAll, nil
	case StateApproved, "authorized":
		return StateApproved, nil
	default:
		return "", fmt.Errorf("unknown comment state %q", state)
	}
}

func ListComments(ctx context.Context, db *sql.DB, opts ListOptions) ([]dto.ArticleComment, error) {
	state, err := NormalizeState(opts.State)
	if err != nil {
		return nil, err
	}

	query := "SELECT id, article_id, content, email, authorized, created_at FROM comments"
	var args []any
	switch state {
	case StatePending:
		query += " WHERE authorized = ?"
		args = append(args, false)
	case StateApproved:
		query += " WHERE authorized = ?"
		args = append(args, true)
	}
	query += " ORDER BY created_at DESC, id DESC"
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanComments(rows)
}

func GetComment(ctx context.Context, db *sql.DB, id string) (dto.ArticleComment, error) {
	row := db.QueryRowContext(ctx, "SELECT id, article_id, content, email, authorized, created_at FROM comments WHERE id = ?", id)
	var comment dto.ArticleComment
	if err := row.Scan(&comment.Id, &comment.ArticleId, &comment.Content, &comment.Email, &comment.Authorized, &comment.CreatedAt); err != nil {
		return dto.ArticleComment{}, err
	}
	return comment, nil
}

func SetCommentAuthorized(ctx context.Context, db *sql.DB, id string, authorized bool) error {
	result, err := db.ExecContext(ctx, "UPDATE comments SET authorized = ? WHERE id = ?", authorized, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func ExportSnapshots(ctx context.Context, db *sql.DB, dataDir string) error {
	comments, err := ListComments(ctx, db, ListOptions{State: StateAll})
	if err != nil {
		return err
	}
	return ExportComments(dataDir, comments)
}

func ExportComments(dataDir string, comments []dto.ArticleComment) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	if err := iou.WriteJsonToFile(filepath.Join(dataDir, "article-email-comments.json"), comments); err != nil {
		return err
	}
	return iou.WriteJsonToFile(filepath.Join(dataDir, "article-comments.json"), PublicComments(comments))
}

func PublicComments(comments []dto.ArticleComment) []dto.ArticleComment {
	publicComments := append([]dto.ArticleComment(nil), comments...)
	for i := range publicComments {
		addr, err := mail.ParseAddress(publicComments[i].Email)
		if err != nil {
			publicComments[i].Email = ""
			continue
		}
		publicComments[i].Email = addr.Name
	}
	return publicComments
}

func BuildDraftBundle(target dto.ArticleComment, allComments []dto.ArticleComment, ownerEmail string) (DraftBundle, error) {
	if _, err := mail.ParseAddress(ownerEmail); err != nil {
		return DraftBundle{}, fmt.Errorf("invalid owner email: %w", err)
	}

	links := CommentLinks{
		Canonical: fmt.Sprintf("https://www.myriad-dreamin.com/article/%s/#comment-%s", target.ArticleId, target.Id),
		CN:        fmt.Sprintf("https://cn.myriad-dreamin.com/article/%s/#comment-%s", target.ArticleId, target.Id),
	}
	commentRef := buildCommentRef(target, links)

	ownerDraft := EmailDraft{
		To:      ownerEmail,
		Bcc:     ReplyRecipients(target, allComments),
		Subject: fmt.Sprintf(`Receiving the Comment #%s from Article "%s" on PoeMagie`, target.Id, target.ArticleId),
		Body: fmt.Sprintf(`Hello,
You received the comment for the article "%s":

%s

If you have any questions, please feel free to contact me.

Best regards,
Myriad Dreamin
i.myriad-dreamin.com`, target.ArticleId, commentRef),
	}

	greeting := "Hello,"
	if name := DisplayName(target.Email); name != "" {
		greeting = fmt.Sprintf("Hello, %s,", name)
	}
	authorDraft := EmailDraft{
		To:      target.Email,
		Subject: fmt.Sprintf(`Comment #%s to Article "%s" on PoeMagie`, target.Id, target.ArticleId),
		Body: fmt.Sprintf(`%s
I would like to authorize the following comment for the article "%s":

%s

If you would like to cancel this authorization, please reply to this email with
the word "Cancel" in the "email body".

If you have any questions, please feel free to contact me.

Best regards,
Myriad Dreamin
i.myriad-dreamin.com`, greeting, target.ArticleId, commentRef),
	}

	ownerDraft.GmailURL = GmailComposeURL(ownerDraft)
	authorDraft.GmailURL = GmailComposeURL(authorDraft)

	return DraftBundle{
		Comment: target,
		Links:   links,
		Owner:   ownerDraft,
		Author:  authorDraft,
	}, nil
}

func ReplyRecipients(target dto.ArticleComment, allComments []dto.ArticleComment) []string {
	commentsByID := make(map[string]dto.ArticleComment, len(allComments))
	emailsByArticleAndName := map[string]map[string][]string{}
	for _, comment := range allComments {
		commentsByID[comment.Id] = comment

		name := DisplayName(comment.Email)
		if name == "" {
			continue
		}
		if emailsByArticleAndName[comment.ArticleId] == nil {
			emailsByArticleAndName[comment.ArticleId] = map[string][]string{}
		}
		emailsByArticleAndName[comment.ArticleId][name] = append(emailsByArticleAndName[comment.ArticleId][name], comment.Email)
	}

	targetIdentity := emailIdentity(target.Email)
	seen := map[string]struct{}{}
	var recipients []string
	appendRecipient := func(email string) {
		identity := emailIdentity(email)
		if identity == "" || identity == targetIdentity {
			return
		}
		if _, ok := seen[identity]; ok {
			return
		}
		seen[identity] = struct{}{}
		recipients = append(recipients, email)
	}

	for _, match := range commentMentionRE.FindAllStringSubmatch(target.Content, -1) {
		if comment, ok := commentsByID[strings.TrimSpace(match[1])]; ok {
			appendRecipient(comment.Email)
		}
	}
	for _, match := range userMentionRE.FindAllStringSubmatch(target.Content, -1) {
		for _, email := range emailsByArticleAndName[target.ArticleId][strings.TrimSpace(match[1])] {
			appendRecipient(email)
		}
	}
	return recipients
}

func DisplayName(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	if addr, err := mail.ParseAddress(email); err == nil && strings.TrimSpace(addr.Name) != "" {
		return strings.TrimSpace(addr.Name)
	}
	if idx := strings.Index(email, "<"); idx >= 0 {
		return strings.TrimSpace(email[:idx])
	}
	return email
}

func GmailComposeURL(draft EmailDraft) string {
	composeURL := url.URL{
		Scheme: "https",
		Host:   "mail.google.com",
		Path:   "/mail/",
	}
	query := url.Values{}
	query.Set("view", "cm")
	query.Set("fs", "1")
	query.Set("to", draft.To)
	if len(draft.Bcc) > 0 {
		query.Set("bcc", strings.Join(draft.Bcc, ","))
	}
	query.Set("su", draft.Subject)
	query.Set("body", draft.Body)
	composeURL.RawQuery = query.Encode()
	return composeURL.String()
}

func FormatCreatedAt(createdAt int64) string {
	if createdAt <= 0 {
		return ""
	}
	return time.UnixMilli(createdAt).Format(time.RFC3339)
}

func scanComments(rows *sql.Rows) ([]dto.ArticleComment, error) {
	comments := make([]dto.ArticleComment, 0)
	for rows.Next() {
		var comment dto.ArticleComment
		if err := rows.Scan(&comment.Id, &comment.ArticleId, &comment.Content, &comment.Email, &comment.Authorized, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func buildCommentRef(comment dto.ArticleComment, links CommentLinks) string {
	return fmt.Sprintf(`The link to the article is:
- Canonical Url: %s
- Asia Mirror (such as China, Japan): %s

Here is the comment:

>>>>>>>>>>>>>>>>>>>>>>>>>>
%s
<<<<<<<<<<<<<<<<<<<<<<<<<<`, links.Canonical, links.CN, comment.Content)
}

func emailIdentity(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	if addr, err := mail.ParseAddress(email); err == nil {
		return strings.ToLower(addr.Address)
	}
	return strings.ToLower(email)
}
