#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT_PATH="${ROOT_DIR}/scripts/sync-comments-snapshot.sh"
LOG_DIR="${BLOG_COMMENTS_SYNC_LOG_DIR:-"${ROOT_DIR}/.data/logs"}"
CRON_LOG="${LOG_DIR}/comments-cron.log"
SCHEDULE="${BLOG_COMMENTS_SYNC_CRON:-0 * * * *}"
SHELL_PATH="${BLOG_COMMENTS_SYNC_SHELL:-${SHELL:-/bin/zsh}}"
BEGIN_MARKER="# BEGIN blog-backend comments snapshot sync"
END_MARKER="# END blog-backend comments snapshot sync"
MODE="install"

usage() {
  cat <<'EOF'
Usage: scripts/install-comments-cron.sh [--uninstall] [--print]

Installs an idempotent crontab block that runs comments snapshot sync hourly.

Environment overrides:

  BLOG_COMMENTS_SYNC_CRON   Cron expression, default: 0 * * * *
  BLOG_COMMENTS_SYNC_SHELL  Shell used by cron, default: $SHELL or /bin/zsh
  BLOG_COMMENTS_SYNC_LOG_DIR
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --uninstall)
      MODE="uninstall"
      ;;
    --print)
      MODE="print"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

shell_quote() {
  local value="$1"
  printf "'%s'" "${value//\'/\'\\\'\'}"
}

current_crontab() {
  crontab -l 2>/dev/null || true
}

without_managed_block() {
  sed "/^${BEGIN_MARKER}$/,/^${END_MARKER}$/d"
}

managed_command() {
  local root
  local script
  local log_file
  local inner

  root="$(shell_quote "${ROOT_DIR}")"
  script="$(shell_quote "${SCRIPT_PATH}")"
  log_file="$(shell_quote "${CRON_LOG}")"
  inner="${script} >> ${log_file} 2>&1"

  printf '%s cd %s && %s -lc %s\n' \
    "${SCHEDULE}" \
    "${root}" \
    "$(shell_quote "${SHELL_PATH}")" \
    "$(shell_quote "${inner}")"
}

print_block() {
  printf '%s\n' "${BEGIN_MARKER}"
  managed_command
  printf '%s\n' "${END_MARKER}"
}

install_cron() {
  local tmp

  command -v crontab >/dev/null 2>&1 || {
    echo "crontab command not found" >&2
    exit 1
  }
  [[ -x "${SCRIPT_PATH}" ]] || chmod +x "${SCRIPT_PATH}"
  mkdir -p "${LOG_DIR}"

  tmp="$(mktemp)"
  {
    current_crontab | without_managed_block
    print_block
  } >"${tmp}"

  crontab "${tmp}"
  rm -f "${tmp}"
  echo "Installed hourly comments snapshot sync:"
  print_block
}

uninstall_cron() {
  local tmp

  command -v crontab >/dev/null 2>&1 || {
    echo "crontab command not found" >&2
    exit 1
  }

  tmp="$(mktemp)"
  current_crontab | without_managed_block >"${tmp}"
  crontab "${tmp}"
  rm -f "${tmp}"
  echo "Removed comments snapshot sync crontab block."
}

case "${MODE}" in
  install)
    install_cron
    ;;
  uninstall)
    uninstall_cron
    ;;
  print)
    print_block
    ;;
esac
