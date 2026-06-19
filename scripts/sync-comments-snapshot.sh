#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${BLOG_BACKEND_ENV_FILE:-"${ROOT_DIR}/.env"}"
DRY_RUN=0

usage() {
  cat <<'EOF'
Usage: scripts/sync-comments-snapshot.sh [--dry-run]

Downloads blog comment snapshots from the backend data directory, compares the
backend public comments snapshot with the frontend snapshot, and when comments
changed:

  - copies fresh snapshots into the backend and frontend working trees
  - commits and pushes frontend snapshot files only
  - deploys the frontend snapshot through make deploy-frontend
  - starts a detached Paseo agent from the meow-launcher repository

Environment is read from .env by default. Useful overrides:

  BLOG_BACKEND_ENV_FILE
  FRONTEND_PATH
  MEOW_LAUNCHER_PATH
  BLOG_COMMENTS_SYNC_LOG_DIR
  BLOG_COMMENTS_SYNC_STATE_DIR
  FRONTEND_GIT_REMOTE
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
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

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

FRONTEND_PATH="${FRONTEND_PATH:-"${HOME}/work/ts/blog"}"
MEOW_LAUNCHER_PATH="${MEOW_LAUNCHER_PATH:-"${HOME}/work/ts/meow-launcher"}"
if [[ -z "${BACKEND_DATA_PATH:-}" && -n "${BACKEND_PATH:-}" ]]; then
  BACKEND_DATA_PATH="${BACKEND_PATH%/}/backend/data"
fi
FRONTEND_GIT_REMOTE="${FRONTEND_GIT_REMOTE:-origin}"
LOG_DIR="${BLOG_COMMENTS_SYNC_LOG_DIR:-"${ROOT_DIR}/.data/logs"}"
STATE_DIR="${BLOG_COMMENTS_SYNC_STATE_DIR:-"${ROOT_DIR}/.data/comment-sync"}"

PUBLIC_COMMENTS_FILE="article-comments.json"
EMAIL_COMMENTS_FILE="article-email-comments.json"
STATS_FILE="article-stats.json"
FRONTEND_SNAPSHOT_DIR="${FRONTEND_PATH%/}/content/snapshot"
FRONTEND_COMMENTS_PATH="${FRONTEND_SNAPSHOT_DIR}/${PUBLIC_COMMENTS_FILE}"
FRONTEND_STATS_PATH="${FRONTEND_SNAPSHOT_DIR}/${STATS_FILE}"
DATA_DIR="${ROOT_DIR}/.data"

mkdir -p "${LOG_DIR}" "${STATE_DIR}"
LOG_FILE="${LOG_DIR}/comments-sync.log"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/blog-comments-sync.XXXXXX")"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

log() {
  printf '[%s] %s\n' "$(date -Is)" "$*" | tee -a "${LOG_FILE}" >&2
}

fail() {
  log "ERROR: $*"
  exit 1
}

require_var() {
  local name="$1"
  local value="${!name:-}"
  [[ -n "${value}" ]] || fail "${name} is required"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

run() {
  log "+ $*"
  if [[ "${DRY_RUN}" -eq 1 ]]; then
    return 0
  fi
  "$@"
}

run_capture() {
  log "+ $*"
  "$@"
}

copy_backend_data_snapshots() {
  if [[ "${DRY_RUN}" -eq 1 ]]; then
    log "dry-run: would copy downloaded snapshots into ${DATA_DIR}"
    return 0
  fi

  mkdir -p "${DATA_DIR}"
  cp "${TMP_DIR}/${STATS_FILE}" "${DATA_DIR}/${STATS_FILE}"
  cp "${TMP_DIR}/${PUBLIC_COMMENTS_FILE}" "${DATA_DIR}/${PUBLIC_COMMENTS_FILE}"
  cp "${TMP_DIR}/${EMAIL_COMMENTS_FILE}" "${DATA_DIR}/${EMAIL_COMMENTS_FILE}"
}

copy_frontend_snapshots() {
  if [[ "${DRY_RUN}" -eq 1 ]]; then
    log "dry-run: would copy public snapshots into ${FRONTEND_SNAPSHOT_DIR}"
    return 0
  fi

  mkdir -p "${FRONTEND_SNAPSHOT_DIR}"
  cp "${TMP_DIR}/${STATS_FILE}" "${FRONTEND_STATS_PATH}"
  cp "${TMP_DIR}/${PUBLIC_COMMENTS_FILE}" "${FRONTEND_COMMENTS_PATH}"
}

frontend_has_staged_snapshot_changes() {
  git -C "${FRONTEND_PATH}" diff --cached --quiet -- \
    "content/snapshot/${PUBLIC_COMMENTS_FILE}" \
    "content/snapshot/${STATS_FILE}"
}

push_frontend_snapshot() {
  local branch

  run git -C "${FRONTEND_PATH}" add \
    "content/snapshot/${PUBLIC_COMMENTS_FILE}" \
    "content/snapshot/${STATS_FILE}" || return

  if frontend_has_staged_snapshot_changes; then
    log "frontend snapshot matches HEAD after copy; no snapshot commit needed"
  else
    run git -C "${FRONTEND_PATH}" commit \
      -m "chore: update comment snapshot" \
      -- \
      "content/snapshot/${PUBLIC_COMMENTS_FILE}" \
      "content/snapshot/${STATS_FILE}" || return
  fi

  if [[ "${DRY_RUN}" -eq 1 ]]; then
    log "dry-run: would push frontend snapshot changes"
    return 0
  fi

  if git -C "${FRONTEND_PATH}" rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
    run git -C "${FRONTEND_PATH}" push || return
    return 0
  fi

  branch="$(git -C "${FRONTEND_PATH}" branch --show-current)"
  [[ -n "${branch}" ]] || fail "frontend repository is not on a branch"
  run git -C "${FRONTEND_PATH}" push -u "${FRONTEND_GIT_REMOTE}" "${branch}" || return
}

deploy_frontend_snapshot() {
  if [[ "${DRY_RUN}" -eq 1 ]]; then
    log "dry-run: would run make deploy-frontend"
    return 0
  fi

  run make -C "${ROOT_DIR}" deploy-frontend
}

start_paseo_notification() {
  local status="$1"
  local detail="$2"
  local prompt_file="${TMP_DIR}/paseo-prompt.md"
  local title

  title="Blog Comments Snapshot ${status}"

  if ! command -v paseo >/dev/null 2>&1; then
    log "paseo not found on PATH; skipped notification agent"
    return 0
  fi

  if [[ ! -d "${MEOW_LAUNCHER_PATH}" ]]; then
    log "meow-launcher repository not found at ${MEOW_LAUNCHER_PATH}; skipped notification agent"
    return 0
  fi

  cat >"${prompt_file}" <<EOF
请用中文提醒 Kamiya：博客评论快照同步任务检测到 comments 已更新。

状态：${status}
详情：${detail}

后端仓库：${ROOT_DIR}
前端仓库：${FRONTEND_PATH}
前端快照：content/snapshot/${PUBLIC_COMMENTS_FILE}
触发时间：$(date -Is)

只需要提醒用户同步结果，不要修改文件。
EOF

  if [[ "${DRY_RUN}" -eq 1 ]]; then
    log "dry-run: would start paseo notification agent from ${MEOW_LAUNCHER_PATH}"
    return 0
  fi

  if ! (
    cd "${MEOW_LAUNCHER_PATH}"
    paseo run --detach --mode full-access --provider codex/gpt-5.4 --name "${title}" "$(cat "${prompt_file}")"
  ); then
    log "failed to start paseo notification agent"
  fi
}

main() {
  local lock_file="${STATE_DIR}/comments-sync.lock"
  local remote_data_path
  local changed=0
  local detail

  require_var SERVER_NAME
  require_var BACKEND_DATA_PATH
  require_cmd cmp
  require_cmd cp
  require_cmd flock
  require_cmd git
  require_cmd make
  require_cmd mktemp
  require_cmd rsync

  [[ -d "${FRONTEND_PATH}" ]] || fail "frontend repository not found: ${FRONTEND_PATH}"
  [[ -d "${FRONTEND_SNAPSHOT_DIR}" ]] || fail "frontend snapshot directory not found: ${FRONTEND_SNAPSHOT_DIR}"

  exec 9>"${lock_file}"
  if ! flock -n 9; then
    log "another comments sync is already running; exiting"
    exit 0
  fi

  remote_data_path="${SERVER_NAME}:${BACKEND_DATA_PATH%/}"
  log "downloading backend comment snapshots from ${remote_data_path}"
  run_capture rsync -az \
    "${remote_data_path}/${STATS_FILE}" \
    "${remote_data_path}/${PUBLIC_COMMENTS_FILE}" \
    "${remote_data_path}/${EMAIL_COMMENTS_FILE}" \
    "${TMP_DIR}/"

  [[ -s "${TMP_DIR}/${PUBLIC_COMMENTS_FILE}" ]] || fail "downloaded public comments snapshot is empty"
  [[ -s "${TMP_DIR}/${EMAIL_COMMENTS_FILE}" ]] || fail "downloaded email comments snapshot is empty"
  [[ -s "${TMP_DIR}/${STATS_FILE}" ]] || fail "downloaded stats snapshot is empty"

  if ! cmp -s "${TMP_DIR}/${PUBLIC_COMMENTS_FILE}" "${FRONTEND_COMMENTS_PATH}"; then
    changed=1
  fi

  copy_backend_data_snapshots

  if [[ "${changed}" -eq 0 ]]; then
    log "backend and frontend public comments snapshots match; no push, deploy, or notification needed"
    exit 0
  fi

  log "backend comments differ from frontend snapshot; updating frontend"
  copy_frontend_snapshots

  if push_frontend_snapshot && deploy_frontend_snapshot; then
    detail="前端快照已提交、推送并部署。"
    log "${detail}"
    start_paseo_notification "success" "${detail}"
    exit 0
  fi

  detail="comments 快照已下载，但推送或部署失败，请查看 ${LOG_FILE}。"
  log "${detail}"
  start_paseo_notification "failed" "${detail}"
  exit 1
}

main "$@"
