#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GATEWAY_PORT="${ANIMUS_GATEWAY_PORT:-8080}"
HEALTH_TIMEOUT_SECONDS="${HEALTH_TIMEOUT_SECONDS:-30}"
HEALTH_POLL_SECONDS="${HEALTH_POLL_SECONDS:-1}"
export ANIMUS_GATEWAY_URL="${ANIMUS_GATEWAY_URL:-http://localhost:${GATEWAY_PORT}}"
export ANIMUS_DEV_SKIP_UI="${ANIMUS_DEV_SKIP_UI:-1}"

COMPOSE_CMD=""
log_file=""

require_tooling() {
  if ! command -v curl >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
    echo "curl (preferred) or python3 is required for health checks" >&2
    exit 1
  fi
}

detect_compose_cmd() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
    return
  fi
  if command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
    return
  fi
  echo "docker compose (preferred) or docker-compose is required" >&2
  exit 1
}

http_get_status() {
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -s -o /dev/null -w "%{http_code}" "${url}" || true
    return 0
  fi
  python3 - <<PY || true
import sys
import urllib.request
try:
    with urllib.request.urlopen("${url}") as resp:
        sys.stdout.write(str(resp.status))
except Exception:
    sys.stdout.write("000")
PY
}

wait_for_health() {
  local url="$1"
  local timeout_s="$2"
  local poll_s="$3"
  local deadline=$((SECONDS + timeout_s))
  local status=""
  while [ "${SECONDS}" -lt "${deadline}" ]; do
    status="$(http_get_status "${url}")"
    if [ "${status}" = "200" ]; then
      return 0
    fi
    sleep "${poll_s}"
  done
  return 1
}

tail_logs() {
  if [ -f "${log_file}" ]; then
    echo "==> last logs" >&2
    tail -n 200 "${log_file}" >&2 || true
  fi
}

cleanup() {
  local code=$?
  if [ -n "${dev_pid:-}" ] && kill -0 "${dev_pid}" >/dev/null 2>&1; then
    kill "${dev_pid}" >/dev/null 2>&1 || true
  fi
  wait "${dev_pid:-}" >/dev/null 2>&1 || true
  if [ -n "${COMPOSE_CMD}" ] && [ "${ANIMUS_DEV_SKIP_INFRA:-0}" != "1" ]; then
    ${COMPOSE_CMD} -f "${ROOT_DIR}/closed/deploy/docker-compose.yml" down >/dev/null 2>&1 || true
  fi
  if [ -f "${log_file}" ]; then
    rm -f "${log_file}" || true
  fi
  exit "${code}"
}
trap cleanup EXIT INT TERM

require_tooling
detect_compose_cmd

export COMPOSE_BIN="${COMPOSE_CMD} -f ${ROOT_DIR}/closed/deploy/docker-compose.yml"
log_file="$(mktemp -t animus-demo.XXXXXX.log)"

echo "==> starting local control plane"
"${ROOT_DIR}/closed/scripts/dev.sh" >"${log_file}" 2>&1 &
dev_pid=$!

echo "==> waiting for gateway ${ANIMUS_GATEWAY_URL}/healthz"
if ! wait_for_health "${ANIMUS_GATEWAY_URL}/healthz" "${HEALTH_TIMEOUT_SECONDS}" "${HEALTH_POLL_SECONDS}"; then
  echo "gateway did not become healthy within ${HEALTH_TIMEOUT_SECONDS}s" >&2
  tail_logs
  exit 1
fi
echo "==> gateway listening on ${ANIMUS_GATEWAY_URL}"

echo "==> running demo"
(
  cd "${ROOT_DIR}"
  go run ./open/cmd/demo -gateway "${ANIMUS_GATEWAY_URL}" -dataset "${ROOT_DIR}/open/demo/data/demo.csv"
)

echo "==> smoke check ok"
if [ "$(http_get_status "${ANIMUS_GATEWAY_URL}/healthz")" != "200" ]; then
  echo "gateway health check failed" >&2
  tail_logs
  exit 1
fi

echo "==> demo complete"
