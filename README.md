# minimum-template
gin extension

## Comment Snapshot Cron

Install the hourly backend-to-frontend comment snapshot sync:

```bash
make install-comments-cron
```

The cron job runs `scripts/sync-comments-snapshot.sh` at `0 * * * *`. It
downloads backend comment snapshots, compares `article-comments.json` with the
frontend snapshot, and only when comments changed commits and pushes the
frontend snapshot files, runs `make deploy-frontend`, and starts a detached
Paseo reminder agent from `/home/kamiyoru/work/ts/meow-launcher`.

Useful commands:

```bash
make comments-sync
scripts/sync-comments-snapshot.sh --dry-run
scripts/install-comments-cron.sh --print
make uninstall-comments-cron
```
