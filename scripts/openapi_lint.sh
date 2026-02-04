#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=/dev/null
source "${ROOT_DIR}/scripts/go_env.sh"

if [[ -n "${GOFLAGS:-}" ]]; then
  export GOFLAGS="${GOFLAGS} -mod=vendor"
else
  export GOFLAGS="-mod=vendor"
fi

SPECS=(
  "${ROOT_DIR}/open/api/openapi/experiments.yaml"
  "${ROOT_DIR}/open/api/openapi/dataplane_internal.yaml"
  "${ROOT_DIR}/open/api/openapi/dataset-registry.yaml"
  "${ROOT_DIR}/open/api/openapi/quality.yaml"
  "${ROOT_DIR}/open/api/openapi/lineage.yaml"
  "${ROOT_DIR}/open/api/openapi/audit.yaml"
  "${ROOT_DIR}/open/api/openapi/gateway.yaml"
)

for spec in "${SPECS[@]}"; do
  if [[ ! -f "${spec}" ]]; then
    echo "missing spec: ${spec}"
    exit 1
  fi
done

FILES_CSV="$(IFS=,; echo "${SPECS[*]}")"
echo "==> openapi lint ${FILES_CSV}"
go run "${ROOT_DIR}/cmd/openapi-lint" --files "${FILES_CSV}"
