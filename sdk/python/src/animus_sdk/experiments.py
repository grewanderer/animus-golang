from __future__ import annotations

import base64
import hashlib
import hmac
import json
import os
from datetime import datetime, timezone

from .git import GitMetadata, get_git_metadata
from .http_client import download_file, request_json, upload_multipart_file_json


def _format_dt(value: datetime | None) -> str | None:
    if value is None:
        return None
    if value.tzinfo is None:
        value = value.replace(tzinfo=timezone.utc)
    return value.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def compute_ci_webhook_signature(secret: str, ts: str, method: str, body: bytes) -> str:
    body_hash = hashlib.sha256(body).hexdigest()
    msg = "\n".join([ts.strip(), method.strip().upper(), body_hash])
    mac = hmac.new(secret.encode("utf-8"), msg.encode("utf-8"), hashlib.sha256).digest()
    return base64.urlsafe_b64encode(mac).decode("utf-8").rstrip("=")


class ExperimentsClient:
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
        self._timeout_seconds = timeout_seconds

    def create_experiment(self, *, name: str, description: str = "", metadata: dict[str, object] | None = None) -> dict:
        body = {
            "name": name,
            "description": description,
            "metadata": metadata or {},
        }
        url = f"{self._gateway_url}/api/experiments/experiments"
        out = request_json("POST", url, json_body=body, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def list_experiments(self, *, limit: int = 100, name: str | None = None) -> dict:
        query = f"limit={int(limit)}"
        if name:
            query += f"&name={name}"
        url = f"{self._gateway_url}/api/experiments/experiments?{query}"
        out = request_json("GET", url, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def create_run(
        self,
        *,
        experiment_id: str,
        dataset_version_id: str | None = None,
        status: str,
        started_at: datetime | None = None,
        ended_at: datetime | None = None,
        git: GitMetadata | None = None,
        params: dict[str, object] | None = None,
        metrics: dict[str, object] | None = None,
        artifacts_prefix: str | None = None,
    ) -> dict:
        git_meta = git or get_git_metadata()
        body: dict[str, object] = {
            "dataset_version_id": dataset_version_id or "",
            "status": status,
            "started_at": _format_dt(started_at) or None,
            "ended_at": _format_dt(ended_at) or None,
            "git_repo": git_meta.repo if git_meta else "",
            "git_commit": git_meta.commit if git_meta else "",
            "git_ref": git_meta.ref if git_meta else "",
            "params": params or {},
            "metrics": metrics or {},
            "artifacts_prefix": artifacts_prefix or "",
        }
        body = {k: v for (k, v) in body.items() if v not in (None, "")}

        url = f"{self._gateway_url}/api/experiments/experiments/{experiment_id}/runs"
        out = request_json("POST", url, json_body=body, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def create_run_with_git(
        self,
        *,
        experiment_id: str,
        dataset_version_id: str | None = None,
        status: str,
        started_at: datetime | None = None,
        ended_at: datetime | None = None,
        params: dict[str, object] | None = None,
        metrics: dict[str, object] | None = None,
        artifacts_prefix: str | None = None,
    ) -> dict:
        git_meta = get_git_metadata() or GitMetadata(repo="", commit="", ref="", source="")
        return self.create_run(
            experiment_id=experiment_id,
            dataset_version_id=dataset_version_id,
            status=status,
            started_at=started_at,
            ended_at=ended_at,
            git=git_meta if git_meta.commit else None,
            params=params,
            metrics=metrics,
            artifacts_prefix=artifacts_prefix,
        )

    def post_ci_webhook(
        self,
        *,
        payload: dict[str, object],
        ci_secret: str | None = None,
        ts: str | None = None,
    ) -> dict:
        secret = (ci_secret or os.environ.get("ANIMUS_CI_WEBHOOK_SECRET") or "").strip()
        if not secret:
            raise ValueError("ci_secret is required (or set ANIMUS_CI_WEBHOOK_SECRET)")

        ts_value = (ts or str(int(datetime.now(tz=timezone.utc).timestamp()))).strip()
        body = json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")

        headers = {
            "X-Animus-CI-Ts": ts_value,
            "X-Animus-CI-Sig": compute_ci_webhook_signature(secret, ts_value, "POST", body),
        }
        url = f"{self._gateway_url}/api/experiments/ci/webhook"
        out = request_json(
            "POST",
            url,
            data=body,
            headers=headers,
            auth_token=self._auth_token,
            timeout_seconds=self._timeout_seconds,
        )
        assert isinstance(out, dict)
        return out

    def post_ci_report(
        self,
        *,
        payload: dict[str, object],
        ci_secret: str | None = None,
        ts: str | None = None,
    ) -> dict:
        secret = (ci_secret or os.environ.get("ANIMUS_CI_WEBHOOK_SECRET") or "").strip()
        if not secret:
            raise ValueError("ci_secret is required (or set ANIMUS_CI_WEBHOOK_SECRET)")

        ts_value = (ts or str(int(datetime.now(tz=timezone.utc).timestamp()))).strip()
        body = json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")

        headers = {
            "X-Animus-CI-Ts": ts_value,
            "X-Animus-CI-Sig": compute_ci_webhook_signature(secret, ts_value, "POST", body),
        }
        url = f"{self._gateway_url}/api/experiments/ci/report"
        out = request_json(
            "POST",
            url,
            data=body,
            headers=headers,
            auth_token=self._auth_token,
            timeout_seconds=self._timeout_seconds,
        )
        assert isinstance(out, dict)
        return out

    def execute_run(
        self,
        *,
        experiment_id: str,
        dataset_version_id: str,
        image_ref: str,
        git_repo: str = "",
        git_commit: str = "",
        git_ref: str = "",
        params: dict[str, object] | None = None,
        resources: dict[str, object] | None = None,
    ) -> dict:
        body: dict[str, object] = {
            "experiment_id": experiment_id,
            "dataset_version_id": dataset_version_id,
            "image_ref": image_ref,
            "git_repo": git_repo,
            "git_commit": git_commit,
            "git_ref": git_ref,
            "params": params or {},
            "resources": resources or {},
        }
        body = {k: v for (k, v) in body.items() if v not in ("", None)}
        url = f"{self._gateway_url}/api/experiments/experiments/runs:execute"
        out = request_json("POST", url, json_body=body, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def get_run(self, *, run_id: str) -> dict:
        url = f"{self._gateway_url}/api/experiments/experiment-runs/{run_id}"
        out = request_json("GET", url, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def list_run_artifacts(self, *, run_id: str, kind: str | None = None, limit: int = 200) -> dict:
        query = f"limit={int(limit)}"
        if kind:
            query += f"&kind={kind}"
        url = f"{self._gateway_url}/api/experiments/experiment-runs/{run_id}/artifacts?{query}"
        out = request_json("GET", url, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def get_run_artifact(self, *, run_id: str, artifact_id: str) -> dict:
        run = (run_id or "").strip()
        if not run:
            raise ValueError("run_id is required")
        artifact = (artifact_id or "").strip()
        if not artifact:
            raise ValueError("artifact_id is required")
        url = f"{self._gateway_url}/api/experiments/experiment-runs/{run}/artifacts/{artifact}"
        out = request_json("GET", url, auth_token=self._auth_token, timeout_seconds=self._timeout_seconds)
        assert isinstance(out, dict)
        return out

    def download_run_artifact(self, *, run_id: str, artifact_id: str, dest_path: str) -> dict[str, object]:
        run = (run_id or "").strip()
        if not run:
            raise ValueError("run_id is required")
        artifact = (artifact_id or "").strip()
        if not artifact:
            raise ValueError("artifact_id is required")
        if not dest_path:
            raise ValueError("dest_path is required")

        url = f"{self._gateway_url}/api/experiments/experiment-runs/{run}/artifacts/{artifact}/download"
        return download_file(
            "GET",
            url,
            dest_path=dest_path,
            auth_token=self._auth_token,
            timeout_seconds=self._timeout_seconds,
        )

    def upload_run_artifact(
        self,
        *,
        run_id: str,
        kind: str,
        file_path: str,
        name: str | None = None,
        metadata: dict[str, object] | None = None,
        filename: str | None = None,
        content_type: str | None = None,
    ) -> dict:
        run = (run_id or "").strip()
        if not run:
            raise ValueError("run_id is required")
        k = (kind or "").strip()
        if not k:
            raise ValueError("kind is required")
        if not file_path:
            raise ValueError("file_path is required")

        fields: dict[str, str] = {"kind": k}
        if name:
            fields["name"] = name
        if metadata is not None:
            fields["metadata"] = json.dumps(metadata, separators=(",", ":"), sort_keys=True)

        url = f"{self._gateway_url}/api/experiments/experiment-runs/{run}/artifacts"
        out = upload_multipart_file_json(
            "POST",
            url,
            fields=fields,
            file_field_name="file",
            file_path=file_path,
            filename=filename,
            content_type=content_type,
            auth_token=self._auth_token,
            timeout_seconds=self._timeout_seconds,
        )
        assert isinstance(out, dict)
        return out
