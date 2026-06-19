---
name: blog-comment-moderator
description: Review and moderate comments for the github.com/Myriad-Dreamin/blog-backend project by using the repository's Go blog-cli workflow. Use when Codex is asked to list pending blog comments, inspect a comment, decide whether a blog comment should be authorized, generate owner/author notification drafts, approve or reject comments, or replace the old Astro comment moderation frontend.
---

# Blog Comment Moderator

## Overview

Use this skill from the `blog-backend` repository root to moderate blog comments through `target/blog-cli`. The CLI owns the fixed workflow: reading SQLite comments, resolving reply recipients, generating Gmail drafts, changing `authorized`, and exporting comment snapshots.

## Operating Rules

- Run commands from the repository root.
- Build the CLI first when `target/blog-cli` is missing or stale:

```bash
make target/blog-cli
```

- Use `--format json` when parsing output programmatically; use markdown/table only for user-facing summaries.
- Do not hand-write SQLite update statements for moderation. Use `target/blog-cli comment authorize` or `target/blog-cli comment reject`.
- Do not send email automatically or open Gmail. Use the CLI-generated drafts and Gmail URLs as artifacts to show the user.
- Treat `reject` as "keep unapproved"; it does not delete comments.
- Ask before mutating comment state unless the user explicitly requested applying decisions.

## Workflow

1. List pending comments:

```bash
target/blog-cli comment list --state pending --format json
```

2. Inspect each target comment with the review packet:

```bash
target/blog-cli comment review <id> --format markdown
```

Use JSON if the next step needs structured fields:

```bash
target/blog-cli comment review <id> --format json
```

3. Decide:

- Approve relevant, good-faith comments that are safe to publish.
- Keep pending comments that are spam, abusive, malicious, doxxing, privacy-leaking, irrelevant, or unreadable.
- If the comment contains links, code, or unusual markup, inspect it carefully before recommending approval.
- If unsure, present the concern and keep the comment pending.

4. Apply only after user confirmation or explicit instruction:

```bash
target/blog-cli comment authorize <id>
target/blog-cli comment reject <id>
```

These commands export `.data/article-comments.json` and `.data/article-email-comments.json` by default. Use `--export=false` only when the user explicitly wants to skip snapshot refresh.

5. Report the result:

- Comment id and final state.
- Any email drafts/Gmail URLs the user needs.
- Any comments left pending and why.

## Command Reference

```bash
target/blog-cli comment list --state pending --format table
target/blog-cli comment show <id> --format markdown
target/blog-cli comment draft <id> --format markdown
target/blog-cli comment review <id> --format markdown
target/blog-cli comment authorize <id>
target/blog-cli comment reject <id>
target/blog-cli comment export
```

Useful global flags:

- `--data-dir <dir>`: defaults to `./.data`.
- `--db <path>`: defaults to `<data-dir>/blog.db`.
- `--owner-email <addr>`: defaults to `Kamiya <camiyoru@gmail.com>`.
