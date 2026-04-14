#!/usr/bin/env bash

set -euo pipefail

log() {
  printf '[deploy-prod] %s\n' "$*"
}

fail() {
  printf '[deploy-prod] ERROR: %s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

require_clean_worktree() {
  if [ "${ALLOW_DIRTY:-0}" = "1" ]; then
    return
  fi

  if [ -n "$(git status --short)" ]; then
    fail "git worktree is dirty; commit or stash changes first, or set ALLOW_DIRTY=1"
  fi
}

resolve_repo_slug() {
  local remote_url
  remote_url="$(git remote get-url "$GIT_REMOTE")"
  remote_url="${remote_url%.git}"
  remote_url="${remote_url#git@github.com:}"
  remote_url="${remote_url#https://github.com/}"

  if [[ "$remote_url" != */* ]]; then
    fail "unable to parse GitHub repository from remote: $remote_url"
  fi

  printf '%s' "$remote_url"
}

github_api_get() {
  local url="$1"
  if [ -n "${GITHUB_TOKEN:-}" ]; then
    curl -fsSL \
      -H "Authorization: Bearer ${GITHUB_TOKEN}" \
      -H "Accept: application/vnd.github+json" \
      "$url"
    return
  fi

  curl -fsSL "$url"
}

wait_for_github_build() {
  local repo_slug="$1"
  local sha="$2"
  local api_url="https://api.github.com/repos/${repo_slug}/actions/runs?branch=${DEPLOY_BRANCH}&per_page=20"
  local attempt response status conclusion html_url

  for attempt in $(seq 1 "$GITHUB_POLL_MAX_ATTEMPTS"); do
    response="$(github_api_get "$api_url")"
    status="$(printf '%s' "$response" | jq -r --arg sha "$sha" --arg name "$WORKFLOW_NAME" '
      [.workflow_runs[]
        | select(.head_sha == $sha and .name == $name)][0].status // empty'
    )"
    conclusion="$(printf '%s' "$response" | jq -r --arg sha "$sha" --arg name "$WORKFLOW_NAME" '
      [.workflow_runs[]
        | select(.head_sha == $sha and .name == $name)][0].conclusion // empty'
    )"
    html_url="$(printf '%s' "$response" | jq -r --arg sha "$sha" --arg name "$WORKFLOW_NAME" '
      [.workflow_runs[]
        | select(.head_sha == $sha and .name == $name)][0].html_url // empty'
    )"

    if [ -z "$status" ]; then
      log "GitHub Actions run not visible yet for ${sha} (attempt ${attempt}/${GITHUB_POLL_MAX_ATTEMPTS})"
    else
      log "GitHub Actions status=${status} conclusion=${conclusion:-pending} ${html_url}"
      if [ "$status" = "completed" ]; then
        [ "$conclusion" = "success" ] || fail "GitHub build failed: ${html_url}"
        return
      fi
    fi

    sleep "$GITHUB_POLL_INTERVAL"
  done

  fail "timed out waiting for GitHub workflow ${WORKFLOW_NAME} for ${sha}"
}

deploy_remote() {
  local compose_cmd="${COMPOSE_BIN}"
  ssh "$DEPLOY_HOST" \
    "cd '$DEPLOY_DIR' && ${compose_cmd} pull && ${compose_cmd} up -d && ${compose_cmd} ps"
}

wait_for_remote_health() {
  local attempt health_status inspect_cmd
  inspect_cmd="docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' ${SERVICE_NAME}"

  for attempt in $(seq 1 "$DEPLOY_HEALTH_MAX_ATTEMPTS"); do
    health_status="$(ssh "$DEPLOY_HOST" "$inspect_cmd" | tr -d '\r')"
    log "Remote health=${health_status} (attempt ${attempt}/${DEPLOY_HEALTH_MAX_ATTEMPTS})"
    if [ "$health_status" = "healthy" ]; then
      ssh "$DEPLOY_HOST" "docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'"
      return
    fi
    sleep "$DEPLOY_HEALTH_POLL_INTERVAL"
  done

  fail "service ${SERVICE_NAME} did not become healthy on ${DEPLOY_HOST}"
}

main() {
  require_command git
  require_command curl
  require_command jq
  require_command ssh

  require_clean_worktree

  local repo_root current_branch current_sha repo_slug
  repo_root="$(git rev-parse --show-toplevel)"
  cd "$repo_root"

  current_branch="$(git rev-parse --abbrev-ref HEAD)"
  [ "$current_branch" = "$DEPLOY_BRANCH" ] || fail "current branch is ${current_branch}; expected ${DEPLOY_BRANCH}"

  current_sha="$(git rev-parse HEAD)"
  repo_slug="$(resolve_repo_slug)"

  log "Pushing ${DEPLOY_BRANCH} to ${GIT_REMOTE}"
  git push "$GIT_REMOTE" "$DEPLOY_BRANCH"

  log "Waiting for workflow ${WORKFLOW_NAME} for ${current_sha}"
  wait_for_github_build "$repo_slug" "$current_sha"

  log "Deploying ${IMAGE_TAG} to ${DEPLOY_HOST}:${DEPLOY_DIR}"
  deploy_remote

  log "Waiting for ${SERVICE_NAME} health check"
  wait_for_remote_health

  log "Deployment finished"
}

GIT_REMOTE="${GIT_REMOTE:-origin}"
DEPLOY_BRANCH="${DEPLOY_BRANCH:-main}"
WORKFLOW_NAME="${WORKFLOW_NAME:-Build GHCR Image (amd64)}"
DEPLOY_HOST="${DEPLOY_HOST:-root@pandora-prod}"
DEPLOY_DIR="${DEPLOY_DIR:-/root/new-api}"
COMPOSE_BIN="${COMPOSE_BIN:-docker compose}"
SERVICE_NAME="${SERVICE_NAME:-new-api}"
IMAGE_TAG="${IMAGE_TAG:-ghcr.io/xiaomingchen/new-api:main-amd64}"
GITHUB_POLL_INTERVAL="${GITHUB_POLL_INTERVAL:-15}"
GITHUB_POLL_MAX_ATTEMPTS="${GITHUB_POLL_MAX_ATTEMPTS:-40}"
DEPLOY_HEALTH_POLL_INTERVAL="${DEPLOY_HEALTH_POLL_INTERVAL:-5}"
DEPLOY_HEALTH_MAX_ATTEMPTS="${DEPLOY_HEALTH_MAX_ATTEMPTS:-20}"

main "$@"
