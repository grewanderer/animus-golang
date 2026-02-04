#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=/dev/null
source "${ROOT_DIR}/scripts/go_env.sh"

if [[ "$#" -eq 0 ]]; then
  set -- ./closed/...
fi

echo "==> go test $*"
exec go test "$@"
