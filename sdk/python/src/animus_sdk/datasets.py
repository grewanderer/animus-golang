from __future__ import annotations

import os

from .http_client import download_file, request_json


class DatasetRegistryClient:
    def __init__(
        self,
        *,
        gateway_url: str | None = None,
        auth_token: str | None = None,
        timeout_seconds: float = 30.0,
    ) -> None:
        base = (gateway_url or os.environ.get("ANIMUS_GATEWAY_URL") or "http://localhost:8080").strip()
        self._gateway_url = base.rstrip("/")
        self._auth_token = (auth_token or os.environ.get("ANIMUS_AUTH_TOKEN") or "").strip() or None
        self._timeout_seconds = float(timeout_seconds)

    def get_dataset_version(self, *, dataset_version_id: str) -> dict:
        version_id = (dataset_version_id or "").strip()
        if not version_id:
            raise ValueError("dataset_version_id is required")
        url = f"{self._gateway_url}/api/dataset-registry/dataset-versions/{version_id}"
        out = request_json("GET", url, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def download_dataset_version(self, *, dataset_version_id: str, dest_path: str) -> dict[str, object]:
        version_id = (dataset_version_id or "").strip()
        if not version_id:
            raise ValueError("dataset_version_id is required")
        if not dest_path:
            raise ValueError("dest_path is required")

        url = f"{self._gateway_url}/api/dataset-registry/dataset-versions/{version_id}/download"
        return download_file(
            "GET",
            url,
            dest_path=dest_path,
            auth_token=self._auth_token,
            timeout_seconds=self._timeout_seconds,
        )

