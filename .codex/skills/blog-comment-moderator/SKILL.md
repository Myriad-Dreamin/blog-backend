---
name: blog-comment-moderator
description: Review and moderate comments for the github.com/Myriad-Dreamin/blog-backend project by using the repository's Go blog-cli workflow. Use when Codex is asked to list pending blog comments, inspect a comment, decide whether a blog comment should be authorized, generate owner/author notification drafts, approve or reject comments, or replace the old Astro comment moderation frontend.
---

# Blog Comment Moderator

## Overview

Use this skill from the `blog-backend` repository root to moderate blog comments through `make download-data`, `target/blog-cli`, and the Makefile moderation targets.

Important data model:

- Local moderation review uses downloaded JSON snapshots. Prefer the private `.data/article-email-comments.json` snapshot because it contains all comments; `.data/article-comments.json` is the public/frontend snapshot and should contain only approved comments.
- The SQLite database is the cloud/server runtime state, not the local source of truth.
- Do not trust local `.data/blog.db` for current pending comments; it may be stale.
- State changes must run against the cloud DB, then local snapshots must be refreshed.
- Comment visibility and moderation are separate: `authorized=true,rejected=false` is approved/public; `authorized=false,rejected=false` is pending review; `authorized=false,rejected=true` is rejected/kept unpublished.
- Only approved comments are exported to the frontend/public snapshot. Rejected comments remain in the private snapshot for audit/history.

## Operating Rules

- Run commands from the repository root.
- Build the CLI first when `target/blog-cli` is missing or stale:

```bash
make target/blog-cli
```

- Use `--format json` when parsing output programmatically; use markdown/table only for user-facing summaries.
- Start by refreshing local snapshots:

```bash
make download-data
```

- If you have not refreshed snapshots in the current turn/session, explicitly tell the user the local comment source may be stale and should be updated with `make download-data` before relying on review results.
- For local read-only review, use the snapshot-backed CLI commands. The comment CLI defaults to `--source snapshot`; use `--source db` only when intentionally running against a known-good server DB context.
- Do not hand-write SQLite update statements for moderation. Use the Makefile targets `make comment-authorize ID=<id>`, `make comment-reject ID=<id>`, or `make comment-delete ID=<id>` so the command runs against the cloud data dir.
- Do not send email automatically or open Gmail. Use the CLI-generated drafts and Gmail URLs as artifacts to show the user.
- When generating email drafts, always include the Gmail compose URL. If URLs are long, put clickable links in a markdown artifact and link that artifact in the final response; do not omit the compose links.
- Treat `delete` as a moderation alias for `reject`: it marks `rejected=true` and `authorized=false`, not physical deletion.
- Treat `reject` as "keep unpublished"; it does not physically delete comments.
- Ask before mutating comment state unless the user explicitly requested applying decisions. If the user explicitly says approve/authorize/reject/delete specific comment IDs, execute the matching Makefile targets directly and report the result.

## Reporting Rules

- Default reports are moderation reports: list pending comments first, oldest to newest. If rejected comments are relevant, report them separately from pending.
- Do not automatically expand approved comments. If useful, mention only the approved count.
- Show approved comment details only when the user explicitly asks for all comments, approved comments, or a specific approved comment id.
- When multiple states are relevant, phrase the result as pending first, then rejected, then approved summary.

## Workflow

1. Refresh local snapshots:

```bash
make download-data
```

2. List pending comments from the downloaded private snapshot:

```bash
target/blog-cli comment list --state pending --format json
```

or:

```bash
make comment-list
```

3. Inspect each target comment with the review packet:

```bash
target/blog-cli comment review <id> --format markdown
```

Use JSON if the next step needs structured fields:

```bash
target/blog-cli comment review <id> --format json
```

or:

```bash
make comment-review ID=<id>
```

4. Decide:

- Approve relevant, good-faith comments that are safe to publish.
- Reject/delete comments that are spam, abusive, malicious, doxxing, privacy-leaking, irrelevant, unreadable, or fake/test submissions. In this workflow, delete means reject/keep unpublished (`authorized=false,rejected=true`), not physical removal.
- If the comment contains links, code, or unusual markup, inspect it carefully before recommending approval.
- If only `.data/article-comments.json` is available, pending/rejected comments may be absent because the public snapshot is expected to contain only approved comments. Do not rely on it for moderation decisions unless `.data/article-email-comments.json` is unavailable and you explicitly state the limitation.
- If unsure, present the concern and keep the comment pending.

5. Apply only after user confirmation or explicit instruction. Use the Makefile targets so mutation happens against the cloud DB and local snapshots are refreshed afterward:

```bash
make comment-authorize ID=<id>
make comment-reject ID=<id>
make comment-delete ID=<id>
```

The remote command exports `article-comments.json` and `article-email-comments.json` on the server by default, and the Makefile target then runs `make download-data`.

6. Report the result:

- Comment id and final state.
- Any email drafts/Gmail URLs the user needs.
- Any comments left pending and why.

## Command Reference

```bash
target/blog-cli comment list --state pending --format table
target/blog-cli comment list --state rejected --format table
target/blog-cli comment list --source snapshot --state pending --format table
target/blog-cli comment list --source db --state pending --format table
target/blog-cli comment show <id> --format markdown
target/blog-cli comment draft <id> --format markdown
target/blog-cli comment review <id> --format markdown
make comment-list
make comment-review ID=<id>
make comment-authorize ID=<id>
make comment-reject ID=<id>
make comment-delete ID=<id>
```

Useful global flags:

- `--data-dir <dir>`: defaults to `./.data`.
- `--db <path>`: defaults to `<data-dir>/blog.db`.
- `--comments <path>`: defaults to `<data-dir>/article-comments.json`.
- `--owner-email <addr>`: defaults to `Kamiya <camiyoru@gmail.com>`.
