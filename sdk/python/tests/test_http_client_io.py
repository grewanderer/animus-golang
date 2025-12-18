import hashlib
import json
import unittest
from pathlib import Path
from tempfile import TemporaryDirectory
from unittest import mock

from animus_sdk.datasets import DatasetRegistryClient
from animus_sdk.experiments import ExperimentsClient


class _FakeHeaders:
    def __init__(self, headers: dict[str, str]):
        self._headers = headers

    def get(self, key: str, default=None):
        return self._headers.get(key, default)


class _FakeHTTPResponse:
    def __init__(self, *, status: int, body: bytes, headers: dict[str, str]):
        self._status = int(status)
        self._body = body
        self._offset = 0
        self.headers = _FakeHeaders(headers)

    def getcode(self):
        return self._status

    def read(self, n: int = -1) -> bytes:
        if n is None or n < 0:
            out = self._body[self._offset :]
            self._offset = len(self._body)
            return out
        if self._offset >= len(self._body):
            return b""
        out = self._body[self._offset : self._offset + n]
        self._offset += len(out)
        return out

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, tb):
        return False


class _FakeConnResponse:
    def __init__(self, *, status: int, body: bytes):
        self.status = int(status)
        self._body = body

    def read(self) -> bytes:
        return self._body


class _FakeHTTPConnection:
    def __init__(self, host, port=None, timeout=None):  # noqa: D401 - mimic http.client signature
        self.host = host
        self.port = port
        self.timeout = timeout
        self.method = None
        self.path = None
        self.headers: dict[str, str] = {}
        self.body_parts: list[bytes] = []

    def putrequest(self, method, url):
        self.method = method
        self.path = url

    def putheader(self, header, value):
        self.headers[str(header)] = str(value)

    def endheaders(self):
        return

    def send(self, data: bytes):
        self.body_parts.append(data)

    def getresponse(self):
        body = json.dumps({"artifact": {"artifact_id": "a1"}}).encode("utf-8")
        return _FakeConnResponse(status=201, body=body)

    def close(self):
        return


class TestHTTPClientIO(unittest.TestCase):
    def test_dataset_download_writes_file_and_sha(self):
        captured: dict[str, object] = {}

        def fake_urlopen(req, timeout=None):
            captured["req"] = req
            return _FakeHTTPResponse(status=200, body=b"hello", headers={"Content-Type": "application/zip"})

        with TemporaryDirectory() as td, mock.patch("animus_sdk.http_client.urllib.request.urlopen", side_effect=fake_urlopen):
            dest = Path(td) / "dataset.zip"
            client = DatasetRegistryClient(gateway_url="http://example", auth_token="token")
            meta = client.download_dataset_version(dataset_version_id="v1", dest_path=str(dest))

            self.assertEqual(dest.read_bytes(), b"hello")
            self.assertEqual(meta.get("content_type"), "application/zip")
            self.assertEqual(meta.get("size_bytes"), 5)
            self.assertEqual(meta.get("sha256"), hashlib.sha256(b"hello").hexdigest())

            req = captured.get("req")
            self.assertIsNotNone(req)
            self.assertEqual(req.get_method(), "GET")
            self.assertEqual(req.get_header("Authorization"), "Bearer token")
            headers = {k.lower(): v for (k, v) in req.header_items()}
            self.assertTrue((headers.get("x-request-id") or "").strip())

    def test_upload_artifact_builds_multipart_body(self):
        fake_conn = _FakeHTTPConnection("example")

        def fake_conn_factory(host, port=None, timeout=None):
            self.assertEqual(host, "example")
            return fake_conn

        with TemporaryDirectory() as td, mock.patch("animus_sdk.http_client.http.client.HTTPConnection", side_effect=fake_conn_factory):
            p = Path(td) / "model.json"
            p.write_bytes(b"abc")

            client = ExperimentsClient(gateway_url="http://example", auth_token="token")
            resp = client.upload_run_artifact(
                run_id="r1",
                kind="model",
                file_path=str(p),
                name="m",
                metadata={"k": "v"},
                filename="model.bin",
                content_type="application/octet-stream",
            )

            self.assertEqual(resp.get("artifact", {}).get("artifact_id"), "a1")

            body = b"".join(fake_conn.body_parts)
            self.assertEqual(int(fake_conn.headers.get("Content-Length", "0")), len(body))
            self.assertEqual(fake_conn.headers.get("Authorization"), "Bearer token")
            self.assertTrue((fake_conn.headers.get("X-Request-Id") or "").strip())

            ct = fake_conn.headers.get("Content-Type", "")
            self.assertIn("multipart/form-data", ct)
            self.assertIn("boundary=", ct)
            boundary = ct.split("boundary=", 1)[1].strip()
            self.assertTrue(boundary)

            self.assertIn(b'Content-Disposition: form-data; name="kind"', body)
            self.assertIn(b"\r\nmodel\r\n", body)
            self.assertIn(b'Content-Disposition: form-data; name="name"', body)
            self.assertIn(b"\r\nm\r\n", body)
            self.assertIn(b'Content-Disposition: form-data; name="metadata"', body)
            self.assertIn(b'{"k":"v"}', body)
            self.assertIn(b'Content-Disposition: form-data; name="file"; filename="model.bin"', body)
            self.assertIn(b"Content-Type: application/octet-stream", body)
            self.assertIn(b"abc", body)
            self.assertTrue(body.rstrip().endswith(f"--{boundary}--".encode("utf-8")))
