#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GATEWAY_PORT="${ANIMUS_GATEWAY_PORT:-8080}"
export ANIMUS_GATEWAY_URL="${ANIMUS_GATEWAY_URL:-http://localhost:${GATEWAY_PORT}}"
export ANIMUS_DEV_SKIP_UI="${ANIMUS_DEV_SKIP_UI:-1}"

http_ok() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -sf "${url}" >/dev/null
    return $?
  fi
  if command -v python3 >/dev/null 2>&1; then
    python3 - <<PY >/dev/null 2>&1
import sys, urllib.request
try:
    with urllib.request.urlopen("${url}") as resp:
        sys.exit(0 if 200 <= resp.status < 300 else 1)
except Exception:
    sys.exit(1)
PY
    return $?
  fi
  echo "curl or python3 is required for health checks" >&2
  return 1
}

wait_for_health() {
  local url="$1"
  local attempts="${2:-60}"
  local sleep_s="${3:-1}"
  local i
  for i in $(seq 1 "${attempts}"); do
    if http_ok "${url}"; then
      return 0
    fi
    sleep "${sleep_s}"
  done
  return 1
}

cleanup() {
  if [ -n "${dev_pid:-}" ] && kill -0 "${dev_pid}" >/dev/null 2>&1; then
    kill "${dev_pid}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT INT TERM

echo "==> starting local control plane"
"${ROOT_DIR}/closed/scripts/dev.sh" &
dev_pid=$!

echo "==> waiting for gateway ${ANIMUS_GATEWAY_URL}/healthz"
if ! wait_for_health "${ANIMUS_GATEWAY_URL}/healthz" 90 1; then
  echo "gateway did not become healthy" >&2
  exit 1
fi
echo "==> gateway listening on ${ANIMUS_GATEWAY_URL}"

echo "==> running demo"
(
  cd "${ROOT_DIR}"
  go run ./open/cmd/demo -gateway "${ANIMUS_GATEWAY_URL}" -dataset "${ROOT_DIR}/open/demo/data/demo.csv"
)

echo "==> smoke check ok"
if ! http_ok "${ANIMUS_GATEWAY_URL}/healthz"; then
  echo "gateway health check failed" >&2
  exit 1
fi

echo "==> demo complete"
