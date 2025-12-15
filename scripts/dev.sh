#!/usr/bin/env bash
set -euo pipefail

pids=()

COMPOSE_BIN="${COMPOSE_BIN:-docker compose}"

start() {
  local name="$1"
  shift
  echo "==> starting ${name}"
  "$@" &
  pids+=("$!")
}

cleanup() {
  local code=$?
  for pid in "${pids[@]:-}"; do
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
    fi
  done
  wait || true
  exit "$code"
}
trap cleanup EXIT INT TERM

if [ "${ANIMUS_DEV_SKIP_INFRA:-0}" != "1" ]; then
  echo "==> starting infra (postgres, minio)"
  ${COMPOSE_BIN} up -d postgres minio minio-init

  minio_init_id="$(${COMPOSE_BIN} ps -q minio-init 2>/dev/null || true)"
  if [ -n "${minio_init_id}" ]; then
    echo "==> waiting for minio-init to complete"
    exit_code="$(docker wait "${minio_init_id}")"
    if [ "${exit_code}" != "0" ]; then
      echo "minio-init failed (exit code ${exit_code})" >&2
      ${COMPOSE_BIN} logs minio-init || true
      exit 1
    fi
  else
    echo "==> warning: could not resolve minio-init container id; continuing" >&2
  fi
fi

postgres_port="${ANIMUS_POSTGRES_PORT:-5432}"
minio_port="${ANIMUS_MINIO_PORT:-9000}"

export DATABASE_URL="${DATABASE_URL:-postgres://animus:animus@localhost:${postgres_port}/animus?sslmode=disable}"
export ANIMUS_MINIO_ENDPOINT="${ANIMUS_MINIO_ENDPOINT:-localhost:${minio_port}}"
export ANIMUS_MINIO_ACCESS_KEY="${ANIMUS_MINIO_ACCESS_KEY:-animus}"
export ANIMUS_MINIO_SECRET_KEY="${ANIMUS_MINIO_SECRET_KEY:-animusminio}"
export ANIMUS_MINIO_BUCKET_DATASETS="${ANIMUS_MINIO_BUCKET_DATASETS:-datasets}"
export ANIMUS_MINIO_BUCKET_ARTIFACTS="${ANIMUS_MINIO_BUCKET_ARTIFACTS:-artifacts}"

export AUTH_MODE="${AUTH_MODE:-dev}"
export AUTH_SESSION_COOKIE_SECURE="${AUTH_SESSION_COOKIE_SECURE:-false}"

if [ -z "${ANIMUS_INTERNAL_AUTH_SECRET:-}" ]; then
  ANIMUS_INTERNAL_AUTH_SECRET="$(head -c 32 /dev/urandom | base64 | tr -d '\n')"
  export ANIMUS_INTERNAL_AUTH_SECRET
  echo "==> generated ANIMUS_INTERNAL_AUTH_SECRET for this session"
fi

if [ -z "${ANIMUS_CI_WEBHOOK_SECRET:-}" ]; then
  ANIMUS_CI_WEBHOOK_SECRET="$(head -c 32 /dev/urandom | base64 | tr -d '\n')"
  export ANIMUS_CI_WEBHOOK_SECRET
  echo "==> generated ANIMUS_CI_WEBHOOK_SECRET for this session"
fi

if [ -z "${MIGRATE_DOCKER_CONTAINER:-}" ] && docker ps --format '{{.Names}}' | grep -q '^animus-postgres$'; then
  export MIGRATE_DOCKER_CONTAINER="animus-postgres"
fi

echo "==> applying migrations"
./scripts/migrate.sh up

start gateway "${GO:-go}" run ./gateway
start dataset-registry "${GO:-go}" run ./dataset-registry
start quality "${GO:-go}" run ./quality
start experiments "${GO:-go}" run ./experiments
start lineage "${GO:-go}" run ./lineage
start audit "${GO:-go}" run ./audit

echo "==> services are running"
wait
