import json
import os
import uuid
import urllib.error
import urllib.request

from .errors import AnimusAPIError


def request_json(
    method: str,
    url: str,
    *,
    json_body: object | None = None,
    data: bytes | None = None,
    headers: dict[str, str] | None = None,
    auth_token: str | None = None,
    timeout_seconds: float = 30.0,
) -> object | None:
    if json_body is not None and data is not None:
        raise ValueError("provide only one of json_body or data")

    req_headers = {"Accept": "application/json"}
    if headers:
        req_headers.update(headers)

    token = auth_token or os.environ.get("ANIMUS_AUTH_TOKEN", "").strip()
    if token:
        req_headers.setdefault("Authorization", f"Bearer {token}")

    req_headers.setdefault("X-Request-Id", uuid.uuid4().hex)

    body_bytes: bytes | None
    if json_body is not None:
        body_bytes = json.dumps(json_body, separators=(",", ":"), sort_keys=True).encode("utf-8")
        req_headers.setdefault("Content-Type", "application/json")
    else:
        body_bytes = data
        if body_bytes is not None:
            req_headers.setdefault("Content-Type", "application/json")

    req = urllib.request.Request(url, data=body_bytes, headers=req_headers, method=method.upper())
    try:
        with urllib.request.urlopen(req, timeout=timeout_seconds) as resp:
            status = resp.getcode()
            raw = resp.read()
            if not raw:
                return None
            try:
                parsed = json.loads(raw.decode("utf-8"))
            except Exception as e:  # noqa: BLE001 - surface parse errors cleanly
                raise AnimusAPIError(status, "invalid_json_response", req_headers.get("X-Request-Id")) from e

            if status >= 400:
                code = "request_failed"
                request_id = None
                if isinstance(parsed, dict):
                    code = str(parsed.get("error") or code)
                    request_id = str(parsed.get("request_id") or "") or None
                raise AnimusAPIError(status, code, request_id, parsed)

            return parsed
    except urllib.error.HTTPError as e:
        status = int(getattr(e, "code", 0) or 0)
        raw = b""
        try:
            raw = e.read()
        except Exception:
            raw = b""

        parsed: object | None = None
        if raw:
            try:
                parsed = json.loads(raw.decode("utf-8"))
            except Exception:
                parsed = None

        code = "request_failed"
        request_id = None
        if isinstance(parsed, dict):
            code = str(parsed.get("error") or code)
            request_id = str(parsed.get("request_id") or "") or None
        raise AnimusAPIError(status, code, request_id, parsed) from None
    except urllib.error.URLError as e:
        raise AnimusAPIError(0, "network_error", None, {"detail": str(e)}) from None

