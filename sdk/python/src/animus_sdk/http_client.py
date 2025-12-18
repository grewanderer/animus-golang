from __future__ import annotations

import hashlib
import http.client
import json
import mimetypes
import os
import uuid
from pathlib import Path
from typing import BinaryIO
from urllib.parse import urlsplit
import urllib.error
import urllib.request

from .errors import AnimusAPIError


def _build_headers(*, headers: dict[str, str] | None, auth_token: str | None) -> dict[str, str]:
    req_headers = {"Accept": "application/json"}
    if headers:
        req_headers.update(headers)

    token = auth_token or os.environ.get("ANIMUS_AUTH_TOKEN", "").strip()
    if token:
        req_headers.setdefault("Authorization", f"Bearer {token}")

    req_headers.setdefault("X-Request-Id", uuid.uuid4().hex)
    return req_headers


def _parse_error_body(status: int, raw: bytes, fallback_request_id: str | None) -> AnimusAPIError:
    parsed: object | None = None
    if raw:
        try:
            parsed = json.loads(raw.decode("utf-8"))
        except Exception:
            parsed = None

    code = "request_failed"
    request_id = fallback_request_id
    if isinstance(parsed, dict):
        code = str(parsed.get("error") or code)
        request_id = str(parsed.get("request_id") or "") or request_id
    return AnimusAPIError(status, code, request_id, parsed)


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

    req_headers = _build_headers(headers=headers, auth_token=auth_token)

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
                raise AnimusAPIError(int(status or 0), "invalid_json_response", req_headers.get("X-Request-Id")) from e

            if int(status or 0) >= 400:
                raise _parse_error_body(int(status or 0), raw, req_headers.get("X-Request-Id"))

            return parsed
    except urllib.error.HTTPError as e:
        status = int(getattr(e, "code", 0) or 0)
        raw = b""
        try:
            raw = e.read()
        except Exception:
            raw = b""

        raise _parse_error_body(status, raw, req_headers.get("X-Request-Id")) from None
    except urllib.error.URLError as e:
        raise AnimusAPIError(0, "network_error", req_headers.get("X-Request-Id"), {"detail": str(e)}) from None


def download_file(
    method: str,
    url: str,
    *,
    dest_path: str,
    headers: dict[str, str] | None = None,
    auth_token: str | None = None,
    timeout_seconds: float = 30.0,
    chunk_size: int = 1024 * 1024,
    max_bytes: int | None = None,
) -> dict[str, object]:
    """
    Download raw bytes to a file with an atomic rename. Returns basic metadata.
    """
    m = method.strip().upper()
    if m not in {"GET", "HEAD"}:
        raise ValueError("download_file supports only GET/HEAD")
    if chunk_size <= 0:
        raise ValueError("chunk_size must be > 0")

    req_headers = _build_headers(headers=headers, auth_token=auth_token)
    request_id = req_headers.get("X-Request-Id")

    dst = Path(dest_path)
    dst.parent.mkdir(parents=True, exist_ok=True)
    tmp = dst.with_suffix(dst.suffix + ".tmp")

    req = urllib.request.Request(url, headers=req_headers, method=m)
    try:
        with urllib.request.urlopen(req, timeout=timeout_seconds) as resp:
            status = int(resp.getcode() or 0)
            if status >= 400:
                raw = resp.read()
                raise _parse_error_body(status, raw, request_id)

            content_type = (resp.headers.get("Content-Type") or "").strip()
            if m == "HEAD":
                return {"content_type": content_type, "size_bytes": 0, "sha256": ""}

            sha256 = hashlib.sha256()
            size_bytes = 0
            with tmp.open("wb") as f:
                while True:
                    chunk = resp.read(chunk_size)
                    if not chunk:
                        break
                    f.write(chunk)
                    sha256.update(chunk)
                    size_bytes += len(chunk)
                    if max_bytes is not None and size_bytes > max_bytes:
                        raise AnimusAPIError(0, "download_too_large", request_id, {"max_bytes": max_bytes, "size_bytes": size_bytes})

            tmp.replace(dst)
            return {"content_type": content_type, "size_bytes": size_bytes, "sha256": sha256.hexdigest()}
    except urllib.error.HTTPError as e:
        status = int(getattr(e, "code", 0) or 0)
        raw = b""
        try:
            raw = e.read()
        except Exception:
            raw = b""
        raise _parse_error_body(status, raw, request_id) from None
    except urllib.error.URLError as e:
        raise AnimusAPIError(0, "network_error", request_id, {"detail": str(e)}) from None
    finally:
        try:
            if tmp.exists():
                tmp.unlink()
        except Exception:
            pass


def _guess_content_type(path: str) -> str:
    guessed, _ = mimetypes.guess_type(path)
    return guessed or "application/octet-stream"


def upload_multipart_file_json(
    method: str,
    url: str,
    *,
    fields: dict[str, str] | None,
    file_field_name: str,
    file_path: str,
    filename: str | None = None,
    content_type: str | None = None,
    headers: dict[str, str] | None = None,
    auth_token: str | None = None,
    timeout_seconds: float = 30.0,
    chunk_size: int = 1024 * 1024,
) -> object | None:
    """
    Streaming multipart/form-data upload (sends file without loading into memory).
    Expects a JSON response body.
    """
    m = method.strip().upper()
    if m not in {"POST", "PUT"}:
        raise ValueError("upload_multipart_file_json supports only POST/PUT")
    if chunk_size <= 0:
        raise ValueError("chunk_size must be > 0")

    target = urlsplit(url)
    if target.scheme not in {"http", "https"}:
        raise ValueError("only http/https URLs are supported")
    if not target.hostname:
        raise ValueError("invalid url (missing host)")

    boundary = f"----animus-{uuid.uuid4().hex}"
    request_headers = _build_headers(headers=headers, auth_token=auth_token)
    request_headers["Content-Type"] = f"multipart/form-data; boundary={boundary}"

    file_name = (filename or Path(file_path).name or "file.bin").strip() or "file.bin"
    file_ct = (content_type or _guess_content_type(file_path)).strip() or "application/octet-stream"

    pre_parts: list[bytes] = []
    for key, value in (fields or {}).items():
        k = str(key).strip()
        if not k:
            continue
        v = str(value)
        pre_parts.append(
            (
                f"--{boundary}\r\n"
                f'Content-Disposition: form-data; name="{k}"\r\n'
                "\r\n"
                f"{v}\r\n"
            ).encode("utf-8")
        )

    file_header = (
        f"--{boundary}\r\n"
        f'Content-Disposition: form-data; name="{file_field_name}"; filename="{file_name}"\r\n'
        f"Content-Type: {file_ct}\r\n"
        "\r\n"
    ).encode("utf-8")
    file_footer = b"\r\n"
    closing = f"--{boundary}--\r\n".encode("utf-8")

    file_size = Path(file_path).stat().st_size
    content_length = sum(len(p) for p in pre_parts) + len(file_header) + file_size + len(file_footer) + len(closing)
    request_headers["Content-Length"] = str(int(content_length))

    path = target.path or "/"
    if target.query:
        path += "?" + target.query

    port = target.port
    if port is None:
        port = 443 if target.scheme == "https" else 80

    conn_cls = http.client.HTTPSConnection if target.scheme == "https" else http.client.HTTPConnection
    conn: http.client.HTTPConnection | None = None

    try:
        conn = conn_cls(target.hostname, port, timeout=timeout_seconds)
        conn.putrequest(m, path)
        for k, v in request_headers.items():
            conn.putheader(k, v)
        conn.endheaders()

        for p in pre_parts:
            conn.send(p)

        conn.send(file_header)
        with Path(file_path).open("rb") as f:
            _stream_copy(conn, f, chunk_size=chunk_size)
        conn.send(file_footer)
        conn.send(closing)

        resp = conn.getresponse()
        status = int(resp.status or 0)
        raw = resp.read()
        if status >= 400:
            raise _parse_error_body(status, raw, request_headers.get("X-Request-Id"))
        if not raw:
            return None
        try:
            return json.loads(raw.decode("utf-8"))
        except Exception as e:  # noqa: BLE001
            raise AnimusAPIError(status, "invalid_json_response", request_headers.get("X-Request-Id")) from e
    except AnimusAPIError:
        raise
    except Exception as e:  # noqa: BLE001 - normalize network errors
        raise AnimusAPIError(0, "network_error", request_headers.get("X-Request-Id"), {"detail": str(e)}) from None
    finally:
        try:
            if conn is not None:
                conn.close()
        except Exception:
            pass


def _stream_copy(conn: http.client.HTTPConnection, f: BinaryIO, *, chunk_size: int) -> None:
    while True:
        chunk = f.read(chunk_size)
        if not chunk:
            return
        conn.send(chunk)
