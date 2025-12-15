#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF' >&2
usage: ./scripts/airgap-bundle.sh --output <dir> [--image-repo <repo>] [--tag <tag>] [--include-infra <0|1>]

Creates an air-gapped install bundle containing:
  - Helm chart package (animus-datapilot-*.tgz)
  - Docker image tarball (images.tar)
  - values.airgap.yaml + SHA256SUMS

The script does NOT pull images from the network. All required images must already exist locally.
EOF
  exit 2
}

OUTPUT_DIR=""
IMAGE_REPO="animus"
TAG="latest"
INCLUDE_INFRA="1"

while [ $# -gt 0 ]; do
  case "$1" in
    --output)
      OUTPUT_DIR="${2:-}"
      shift 2
      ;;
    --image-repo)
      IMAGE_REPO="${2:-}"
      shift 2
      ;;
    --tag)
      TAG="${2:-}"
      shift 2
      ;;
    --include-infra)
      INCLUDE_INFRA="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "unknown arg: $1" >&2
      usage
      ;;
  esac
done

if [ -z "${OUTPUT_DIR}" ]; then
  usage
fi

if ! command -v helm >/dev/null 2>&1; then
  echo "helm not found" >&2
  exit 2
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found (required to produce images.tar)" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHART_DIR="${ROOT_DIR}/deploy/helm/animus-datapilot"

mkdir -p "${OUTPUT_DIR}"

echo "==> packaging Helm chart"
helm package "${CHART_DIR}" --destination "${OUTPUT_DIR}" >/dev/null

chart_pkg="$(ls -1t "${OUTPUT_DIR}"/animus-datapilot-*.tgz | head -n 1)"
if [ -z "${chart_pkg}" ]; then
  echo "helm package did not produce a chart archive" >&2
  exit 2
fi

POSTGRES_IMAGE="${POSTGRES_IMAGE:-postgres:14-alpine}"
MINIO_IMAGE="${MINIO_IMAGE:-minio/minio@sha256:14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e}"
MINIO_MC_IMAGE="${MINIO_MC_IMAGE:-minio/mc@sha256:a7fe349ef4bd8521fb8497f55c6042871b2ae640607cf99d9bede5e9bdf11727}"

images=(
  "${IMAGE_REPO}/gateway:${TAG}"
  "${IMAGE_REPO}/dataset-registry:${TAG}"
  "${IMAGE_REPO}/quality:${TAG}"
  "${IMAGE_REPO}/experiments:${TAG}"
  "${IMAGE_REPO}/lineage:${TAG}"
  "${IMAGE_REPO}/audit:${TAG}"
  "${IMAGE_REPO}/ui:${TAG}"
)

if [ "${INCLUDE_INFRA}" = "1" ]; then
  images+=("${POSTGRES_IMAGE}" "${MINIO_IMAGE}" "${MINIO_MC_IMAGE}")
fi

echo "==> verifying images exist locally"
missing=()
for img in "${images[@]}"; do
  if ! docker image inspect "${img}" >/dev/null 2>&1; then
    missing+=("${img}")
  fi
done
if [ "${#missing[@]}" -gt 0 ]; then
  echo "missing local images:" >&2
  for img in "${missing[@]}"; do
    echo "  - ${img}" >&2
  done
  echo "build or load them first, then rerun the bundler." >&2
  exit 2
fi

echo "==> writing images.tar"
docker save -o "${OUTPUT_DIR}/images.tar" "${images[@]}"

cat > "${OUTPUT_DIR}/values.airgap.yaml" <<EOF
image:
  repository: ${IMAGE_REPO}
  tag: ${TAG}
  pullPolicy: IfNotPresent

postgres:
  image: ${POSTGRES_IMAGE}

minio:
  image: ${MINIO_IMAGE}
  mcImage: ${MINIO_MC_IMAGE}
EOF

echo "==> generating SHA256SUMS"
(cd "${OUTPUT_DIR}" && sha256sum "$(basename "${chart_pkg}")" images.tar values.airgap.yaml > SHA256SUMS)

cat > "${OUTPUT_DIR}/README.txt" <<EOF
Animus DataPilot — Air-gapped bundle

Contents:
  - $(basename "${chart_pkg}") (Helm chart)
  - images.tar (container images)
  - values.airgap.yaml (baseline values)
  - SHA256SUMS

Next steps:
  1) Verify integrity:
       sha256sum -c SHA256SUMS
  2) Load images on your build host:
       docker load -i images.tar
  3) Push images to a registry reachable by your Kubernetes cluster,
     or load them into the cluster (kind/minikube/etc).
  4) Install the chart:
       helm upgrade --install animus-datapilot $(basename "${chart_pkg}") -f values.airgap.yaml
EOF

echo "==> bundle created at ${OUTPUT_DIR}"
