package commentmoderation

import (
	"context"
	"database/sql"
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

func TestPublicComments(t *testing.T) {
	comments := []dto.ArticleComment{
		{Id: "1", Email: "Alice <alice@example.com>"},
		{Id: "2", Email: "not an email"},
	}

	publicComments := PublicComments(comments)
	if publicComments[0].Email != "Alice" {
		t.Fatalf("public email = %q, want Alice", publicComments[0].Email)
	}
	if publicComments[1].Email != "" {
		t.Fatalf("invalid public email = %q, want empty", publicComments[1].Email)
	}
	if comments[0].Email != "Alice <alice@example.com>" {
		t.Fatalf("PublicComments mutated input: %q", comments[0].Email)
	}
}

func TestListAndSetCommentAuthorized(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	pending, err := ListComments(ctx, db, ListOptions{State: StatePending})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].Id != "1" {
		t.Fatalf("pending = %#v, want only comment 1", pending)
	}

	if err := SetCommentAuthorized(ctx, db, "1", true); err != nil {
		t.Fatal(err)
	}
	approved, err := ListComments(ctx, db, ListOptions{State: StateApproved})
	if err != nil {
		t.Fatal(err)
	}
	if len(approved) != 2 {
		t.Fatalf("approved len = %d, want 2: %#v", len(approved), approved)
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
  created_at INTEGER
);
INSERT INTO comments (article_id, email, content, authorized, created_at) VALUES
  ('a', 'Alice <alice@example.com>', 'pending', false, 10),
  ('a', 'Bob <bob@example.com>', 'approved', true, 20);
`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
