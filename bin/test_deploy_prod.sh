#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT_PATH="$ROOT_DIR/bin/deploy-prod.sh"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

STATE_DIR="$TMP_DIR/state"
MOCK_BIN_DIR="$TMP_DIR/mockbin"
mkdir -p "$STATE_DIR" "$MOCK_BIN_DIR"
touch "$STATE_DIR/calls"

cat >"$MOCK_BIN_DIR/git" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "git \$*" >> "$STATE_DIR/calls"
case "\$*" in
  "rev-parse --show-toplevel")
    echo "$ROOT_DIR"
    ;;
  "remote get-url origin")
    echo "git@github.com:xiaomingchen/new-api.git"
    ;;
  "rev-parse --abbrev-ref HEAD")
    echo "main"
    ;;
  "rev-parse HEAD")
    echo "0123456789abcdef0123456789abcdef01234567"
    ;;
  "status --short")
    ;;
  "push origin main")
    ;;
  *)
    echo "unexpected git args: \$*" >&2
    exit 1
    ;;
esac
EOF

cat >"$MOCK_BIN_DIR/curl" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "curl \$*" >> "$STATE_DIR/calls"
COUNT_FILE="$STATE_DIR/curl-count"
COUNT=0
if [ -f "\$COUNT_FILE" ]; then
  COUNT="\$(cat "\$COUNT_FILE")"
fi
COUNT="\$((COUNT + 1))"
echo "\$COUNT" > "\$COUNT_FILE"

if [ "\$COUNT" -eq 1 ]; then
  cat <<'JSON'
{"workflow_runs":[{"name":"Build GHCR Image (amd64)","head_sha":"0123456789abcdef0123456789abcdef01234567","status":"in_progress","conclusion":null,"html_url":"https://example.test/run/1"}]}
JSON
else
  cat <<'JSON'
{"workflow_runs":[{"name":"Build GHCR Image (amd64)","head_sha":"0123456789abcdef0123456789abcdef01234567","status":"completed","conclusion":"success","html_url":"https://example.test/run/1"}]}
JSON
fi
EOF

cat >"$MOCK_BIN_DIR/ssh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "ssh \$*" >> "$STATE_DIR/calls"
if [[ "\$*" == *"docker inspect --format"* ]]; then
  echo "healthy"
fi
EOF

cat >"$MOCK_BIN_DIR/sleep" <<EOF
#!/usr/bin/env bash
set -euo pipefail
echo "sleep \$*" >> "$STATE_DIR/calls"
EOF

chmod +x "$MOCK_BIN_DIR/git" "$MOCK_BIN_DIR/curl" "$MOCK_BIN_DIR/ssh" "$MOCK_BIN_DIR/sleep"

PATH="$MOCK_BIN_DIR:$PATH" \
  GITHUB_POLL_INTERVAL=0 \
  GITHUB_POLL_MAX_ATTEMPTS=3 \
  DEPLOY_HEALTH_POLL_INTERVAL=0 \
  DEPLOY_HEALTH_MAX_ATTEMPTS=2 \
  bash "$SCRIPT_PATH"

grep -q 'git push origin main' "$STATE_DIR/calls"
grep -q 'curl ' "$STATE_DIR/calls"
grep -q 'ssh .*docker compose pull && docker compose up -d && docker compose ps' "$STATE_DIR/calls"
grep -q 'ssh .*docker inspect --format' "$STATE_DIR/calls"

echo "deploy-prod test passed"
