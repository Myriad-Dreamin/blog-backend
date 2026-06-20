package commentmoderation

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Myriad-Dreamin/blog-backend/pkg/dto"

	_ "github.com/mattn/go-sqlite3"
)

func TestReplyRecipients(t *testing.T) {
	comments := []dto.ArticleComment{
		{Id: "1", ArticleId: "a", Email: "Alice <alice@example.com>", Content: "first"},
		{Id: "2", ArticleId: "a", Email: "Bob <bob@example.com>", Content: "second"},
		{Id: "3", ArticleId: "a", Email: "Carol <carol@example.com>", Content: "third"},
		{Id: "4", ArticleId: "b", Email: "Alice <other@example.com>", Content: "other article"},
	}
	target := dto.ArticleComment{
		Id:        "5",
		ArticleId: "a",
		Email:     "Alice <alice@example.com>",
		Content:   "reply [comment:2] and [user:Carol] and self [user:Alice]",
	}

	recipients := ReplyRecipients(target, comments)
	want := []string{"Bob <bob@example.com>", "Carol <carol@example.com>"}
	if len(recipients) != len(want) {
		t.Fatalf("recipients len = %d, want %d: %#v", len(recipients), len(want), recipients)
	}
	for i := range want {
		if recipients[i] != want[i] {
			t.Fatalf("recipients[%d] = %q, want %q", i, recipients[i], want[i])
		}
	}
}

func TestReplyRecipientsHandlesLegacyAnonymousDisplayNames(t *testing.T) {
	comments := []dto.ArticleComment{
		{Id: "2", ArticleId: "a", Email: "legacy@example.com", Content: "old bare email"},
		{Id: "3", ArticleId: "a", Email: "not an email", Content: "bad legacy email"},
	}
	target := dto.ArticleComment{
		Id:        "5",
		ArticleId: "a",
		Email:     "Alice <alice@example.com>",
		Content:   "reply [user:Anonymous #2] and [user:Anonymous #3]",
	}

	recipients := ReplyRecipients(target, comments)
	want := []string{"legacy@example.com"}
	if len(recipients) != len(want) {
		t.Fatalf("recipients len = %d, want %d: %#v", len(recipients), len(want), recipients)
	}
	for i := range want {
		if recipients[i] != want[i] {
			t.Fatalf("recipients[%d] = %q, want %q", i, recipients[i], want[i])
		}
	}
}

func TestPublicComments(t *testing.T) {
	comments := []dto.ArticleComment{
		{Id: "1", Email: "Alice <alice@example.com>", Authorized: true},
		{Id: "2", Email: "not an email", Authorized: true},
		{Id: "3", Email: "spam@example.com", Authorized: false},
		{Id: "4", Email: "rejected@example.com", Authorized: true, Rejected: true},
		{Id: "5", Email: "foo@example.com", Authorized: true},
		{Id: "6", Email: "\"Bad\x7fName\" <bad@example.com>", Authorized: true},
	}

	publicComments := PublicComments(comments)
	if len(publicComments) != 4 {
		t.Fatalf("public comments len = %d, want 4: %#v", len(publicComments), publicComments)
	}
	if publicComments[0].Email != "Alice" {
		t.Fatalf("public email = %q, want Alice", publicComments[0].Email)
	}
	if publicComments[1].Email != "Anonymous #2" {
		t.Fatalf("invalid public email = %q, want Anonymous #2", publicComments[1].Email)
	}
	if publicComments[2].Email != "Anonymous #5" {
		t.Fatalf("bare public email = %q, want Anonymous #5", publicComments[2].Email)
	}
	if publicComments[3].Email != "Anonymous #6" {
		t.Fatalf("bad display name public email = %q, want Anonymous #6", publicComments[3].Email)
	}
	if publicComments[0].Id != "1" || publicComments[1].Id != "2" || publicComments[2].Id != "5" || publicComments[3].Id != "6" {
		t.Fatalf("public comment ids = %#v, want comments 1, 2, 5, and 6", publicComments)
	}
	if comments[0].Email != "Alice <alice@example.com>" {
		t.Fatalf("PublicComments mutated input: %q", comments[0].Email)
	}
}

func TestValidateCommentEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{name: "valid", email: "Alice <alice@example.com>"},
		{name: "trim spaces", email: "  Alice <alice@example.com>  "},
		{name: "bare address", email: "alice@example.com", wantErr: true},
		{name: "empty display name", email: `"" <alice@example.com>`, wantErr: true},
		{name: "invalid address", email: "not an email", wantErr: true},
		{name: "non-printable email", email: "Alice <alice@example.com>\x7f", wantErr: true},
		{name: "non-printable display name", email: "\"Alice\x7f\" <alice@example.com>", wantErr: true},
		{name: "reserved display name character", email: `"Alice [admin]" <alice@example.com>`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateCommentEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCommentEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestListAndSetCommentState(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	pending, err := ListComments(ctx, db, ListOptions{State: StatePending})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].Id != "1" {
		t.Fatalf("pending = %#v, want only comment 1", pending)
	}

	rejected, err := ListComments(ctx, db, ListOptions{State: StateRejected})
	if err != nil {
		t.Fatal(err)
	}
	if len(rejected) != 1 || rejected[0].Id != "3" {
		t.Fatalf("rejected = %#v, want only comment 3", rejected)
	}

	if err := SetCommentState(ctx, db, "1", StateApproved); err != nil {
		t.Fatal(err)
	}
	approved, err := ListComments(ctx, db, ListOptions{State: StateApproved})
	if err != nil {
		t.Fatal(err)
	}
	if len(approved) != 2 {
		t.Fatalf("approved len = %d, want 2: %#v", len(approved), approved)
	}

	if err := SetCommentState(ctx, db, "1", StateRejected); err != nil {
		t.Fatal(err)
	}
	pending, err = ListComments(ctx, db, ListOptions{State: StatePending})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending after reject = %#v, want none", pending)
	}
	rejected, err = ListComments(ctx, db, ListOptions{State: StateRejected})
	if err != nil {
		t.Fatal(err)
	}
	if len(rejected) != 2 || rejected[0].Id != "1" || rejected[1].Id != "3" {
		t.Fatalf("rejected after update = %#v, want comments 1 and 3", rejected)
	}
}

func TestLoadSnapshotCommentsPrefersPrivateAllComments(t *testing.T) {
	dir := t.TempDir()
	publicPath := filepath.Join(dir, PublicCommentsFilename)
	emailPath := filepath.Join(dir, EmailCommentsFilename)

	writeTestComments(t, publicPath, []dto.ArticleComment{
		{Id: "1", ArticleId: "a", Email: "Alice", Content: "approved", Authorized: true, CreatedAt: 10},
	})
	writeTestComments(t, emailPath, []dto.ArticleComment{
		{Id: "1", ArticleId: "a", Email: "Alice <alice@example.com>", Content: "approved", Authorized: true, CreatedAt: 10},
		{Id: "2", ArticleId: "a", Email: "Bob <bob@example.com>", Content: "older pending", Authorized: false, CreatedAt: 20},
		{Id: "3", ArticleId: "a", Email: "Carol <carol@example.com>", Content: "rejected", Authorized: false, Rejected: true, CreatedAt: 30},
	})

	comments, err := LoadSnapshotComments(SnapshotOptions{
		PublicPath: publicPath,
		EmailPath:  emailPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	comment, err := GetCommentFromSlice(comments, "1")
	if err != nil {
		t.Fatal(err)
	}
	if comment.Email != "Alice <alice@example.com>" {
		t.Fatalf("private email = %q, want full private email", comment.Email)
	}
	comment, err = GetCommentFromSlice(comments, "2")
	if err != nil {
		t.Fatal(err)
	}
	if comment.Email != "Bob <bob@example.com>" {
		t.Fatalf("pending private email = %q, want Bob <bob@example.com>", comment.Email)
	}

	pending, err := ListCommentsFromSlice(comments, ListOptions{State: StatePending})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].Id != "2" {
		t.Fatalf("pending comments = %#v, want only comment 2", pending)
	}
	rejected, err := ListCommentsFromSlice(comments, ListOptions{State: StateRejected})
	if err != nil {
		t.Fatal(err)
	}
	if len(rejected) != 1 || rejected[0].Id != "3" {
		t.Fatalf("rejected comments = %#v, want only comment 3", rejected)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec(`
CREATE TABLE comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  article_id TEXT,
	  email TEXT,
	  content TEXT,
	  authorized BOOLEAN NOT NULL DEFAULT FALSE,
	  rejected BOOLEAN NOT NULL DEFAULT FALSE,
	  created_at INTEGER
	);
	INSERT INTO comments (article_id, email, content, authorized, rejected, created_at) VALUES
	  ('a', 'Alice <alice@example.com>', 'pending', false, false, 10),
	  ('a', 'Bob <bob@example.com>', 'approved', true, false, 20),
	  ('a', 'Carol <carol@example.com>', 'rejected', false, true, 30);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func writeTestComments(t *testing.T, path string, comments []dto.ArticleComment) {
	t.Helper()

	data, err := json.Marshal(comments)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
