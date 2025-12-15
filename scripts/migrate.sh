#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="${ROOT_DIR}/migrations"

DATABASE_URL="${DATABASE_URL:-postgres://animus:animus@localhost:5432/animus?sslmode=disable}"
MIGRATE_DOCKER_CONTAINER="${MIGRATE_DOCKER_CONTAINER:-}"
MIGRATE_DOCKER_PGUSER="${MIGRATE_DOCKER_PGUSER:-animus}"
MIGRATE_DOCKER_PGDATABASE="${MIGRATE_DOCKER_PGDATABASE:-animus}"

if [ -z "${MIGRATE_DOCKER_CONTAINER}" ] && command -v docker >/dev/null 2>&1; then
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q '^animus-postgres$'; then
    MIGRATE_DOCKER_CONTAINER="animus-postgres"
  fi
fi

psql_cmd() {
  if [ -n "${MIGRATE_DOCKER_CONTAINER}" ]; then
    docker exec -i "${MIGRATE_DOCKER_CONTAINER}" psql -U "${MIGRATE_DOCKER_PGUSER}" -d "${MIGRATE_DOCKER_PGDATABASE}" -v ON_ERROR_STOP=1 "$@"
    return
  fi
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 "$@"
}

ensure_schema_migrations() {
  psql_cmd <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL
}

list_up_migrations() {
  find "${MIGRATIONS_DIR}" -maxdepth 1 -type f -name '*.up.sql' -print0 \
    | sort -z \
    | xargs -0 -n1 echo
}

apply_up() {
  ensure_schema_migrations

  local file base ver_raw ver name
  mapfile -t files < <(list_up_migrations)
  for file in "${files[@]}"; do
    base="$(basename "${file}")"
    ver_raw="${base%%_*}"
    if [[ ! "${ver_raw}" =~ ^[0-9]+$ ]]; then
      echo "invalid migration filename (missing numeric version prefix): ${base}" >&2
      exit 2
    fi
    ver=$((10#${ver_raw}))
    name="${base%.up.sql}"
    if [[ ! "${name}" =~ ^[0-9]+_[A-Za-z0-9_]+$ ]]; then
      echo "invalid migration name (only letters/digits/underscore allowed): ${base}" >&2
      exit 2
    fi

    if psql_cmd -Atqc "SELECT 1 FROM schema_migrations WHERE version = ${ver} LIMIT 1" | grep -q 1; then
      continue
    fi

    echo "==> applying ${base}"
    psql_cmd < "${file}"
    psql_cmd -c "INSERT INTO schema_migrations(version, name) VALUES (${ver}, '${name}')"
  done
}

apply_down() {
  ensure_schema_migrations

  local ver
  ver="$(psql_cmd -Atqc 'SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1')"
  if [ -z "${ver}" ]; then
    echo "no migrations applied"
    return 0
  fi

  local ver_prefix
  ver_prefix="$(printf '%06d' "${ver}")"

  local down_file
  down_file="$(find "${MIGRATIONS_DIR}" -maxdepth 1 -type f -name "${ver_prefix}_*.down.sql" -print -quit)"
  if [ -z "${down_file}" ]; then
    echo "missing down migration for version ${ver_prefix}" >&2
    exit 2
  fi

  echo "==> reverting $(basename "${down_file}")"
  psql_cmd < "${down_file}"
  psql_cmd -c "DELETE FROM schema_migrations WHERE version = ${ver}"
}

cmd="${1:-}"
case "${cmd}" in
  up)
    apply_up
    ;;
  down)
    apply_down
    ;;
  *)
    echo "usage: $0 {up|down}" >&2
    exit 2
    ;;
esac
