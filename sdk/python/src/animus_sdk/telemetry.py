from __future__ import annotations

import os
import queue
import threading
import time
from dataclasses import dataclass

from .errors import AnimusAPIError
from .http_client import request_json


def _normalize_base(url: str) -> str:
    url = (url or "").strip()
    return url.rstrip("/")


@dataclass
class _TelemetryTask:
    kind: str
    url: str
    body: dict[str, object]
    attempts: int = 0
    next_retry_at: float = 0.0


class RunTelemetryLogger:
    def __init__(
        self,
        *,
        gateway_url: str,
        run_id: str,
        auth_token: str | None = None,
        timeout_seconds: float = 5.0,
        max_queue: int = 2048,
        max_retries: int = 6,
    ) -> None:
        self._gateway_url = _normalize_base(gateway_url)
        if not self._gateway_url:
            raise ValueError("gateway_url is required")
        self._run_id = (run_id or "").strip()
        if not self._run_id:
            raise ValueError("run_id is required")

        token = (auth_token or "").strip() or None
        self._auth_token = token
        self._timeout_seconds = float(timeout_seconds)

        if max_queue <= 0:
            raise ValueError("max_queue must be > 0")
        self._queue: queue.Queue[_TelemetryTask | None] = queue.Queue(maxsize=int(max_queue))
        self._max_retries = int(max_retries)
        self._stop = threading.Event()
        self._last_error: AnimusAPIError | None = None
        self._thread = threading.Thread(target=self._run_loop, name="animus-telemetry", daemon=True)
        self._thread.start()

    @classmethod
    def from_env(
        cls,
        *,
        gateway_url: str | None = None,
        run_id: str | None = None,
        auth_token: str | None = None,
        timeout_seconds: float = 5.0,
        max_queue: int = 2048,
        max_retries: int = 6,
    ) -> "RunTelemetryLogger":
        base = (gateway_url or os.environ.get("DATAPILOT_URL") or os.environ.get("ANIMUS_GATEWAY_URL") or "").strip()
        if not base:
            base = "http://localhost:8080"

        run = (run_id or os.environ.get("RUN_ID") or "").strip()
        if not run:
            raise ValueError("run_id is required (or set RUN_ID)")

        token = (auth_token or os.environ.get("TOKEN") or os.environ.get("ANIMUS_AUTH_TOKEN") or "").strip() or None

        return cls(
            gateway_url=base,
            run_id=run,
            auth_token=token,
            timeout_seconds=timeout_seconds,
            max_queue=max_queue,
            max_retries=max_retries,
        )

    @property
    def run_id(self) -> str:
        return self._run_id

    @property
    def last_error(self) -> AnimusAPIError | None:
        return self._last_error

    def log_metric(self, *, step: int, name: str, value: float, metadata: dict[str, object] | None = None) -> bool:
        return self.log_metrics(step=step, metrics={name: value}, metadata=metadata)

    def log_metrics(self, *, step: int, metrics: dict[str, float], metadata: dict[str, object] | None = None) -> bool:
        if step < 0:
            raise ValueError("step must be >= 0")
        if not metrics:
            raise ValueError("metrics must not be empty")
        body: dict[str, object] = {"step": int(step), "metrics": metrics}
        if metadata:
            body["metadata"] = metadata

        url = f"{self._gateway_url}/api/experiments/experiment-runs/{self._run_id}/metrics"
        return self._enqueue(_TelemetryTask(kind="metrics", url=url, body=body))

    def log_status(self, *, status: str, message: str | None = None, metadata: dict[str, object] | None = None) -> bool:
        status_value = (status or "").strip()
        if not status_value:
            raise ValueError("status is required")

        meta: dict[str, object] = {"status": status_value}
        if metadata:
            meta.update(metadata)

        msg = (message or "").strip() or f"status: {status_value}"
        return self._log_event(level="info", message=msg, metadata=meta)

    def log_event(self, *, level: str, message: str, metadata: dict[str, object] | None = None) -> bool:
        return self._log_event(level=level, message=message, metadata=metadata)

    def log_progress(
        self,
        *,
        step: int,
        total_steps: int | None = None,
        percent: float | None = None,
        message: str | None = None,
        metadata: dict[str, object] | None = None,
    ) -> bool:
        if step < 0:
            raise ValueError("step must be >= 0")
        meta: dict[str, object] = {"progress_step": int(step)}
        if total_steps is not None:
            if total_steps <= 0:
                raise ValueError("total_steps must be > 0")
            meta["progress_total_steps"] = int(total_steps)
        if percent is not None:
            meta["progress_percent"] = float(percent)
        if metadata:
            meta.update(metadata)

        msg = (message or "").strip() or "progress"
        return self._log_event(level="info", message=msg, metadata=meta)

    def close(self, *, flush: bool = True, timeout_seconds: float = 5.0) -> None:
        self._stop.set()
        try:
            self._queue.put_nowait(None)
        except queue.Full:
            pass
        if flush:
            self.flush(timeout_seconds=timeout_seconds)
        self._thread.join(timeout=max(0.1, float(timeout_seconds)))

    def flush(self, *, timeout_seconds: float = 5.0) -> bool:
        deadline = time.monotonic() + max(0.0, float(timeout_seconds))
        while time.monotonic() < deadline:
            if getattr(self._queue, "unfinished_tasks", 0) == 0:
                return True
            time.sleep(0.05)
        return getattr(self._queue, "unfinished_tasks", 0) == 0

    def _log_event(self, *, level: str, message: str, metadata: dict[str, object] | None = None) -> bool:
        msg = (message or "").strip()
        if not msg:
            raise ValueError("message is required")
        lvl = (level or "").strip().lower() or "info"
        if lvl not in {"debug", "info", "warn", "error"}:
            raise ValueError("invalid level")

        body: dict[str, object] = {"level": lvl, "message": msg}
        if metadata:
            body["metadata"] = metadata
        url = f"{self._gateway_url}/api/experiments/experiment-runs/{self._run_id}/events"
        return self._enqueue(_TelemetryTask(kind="event", url=url, body=body))

    def _enqueue(self, task: _TelemetryTask) -> bool:
        if self._stop.is_set():
            return False
        try:
            self._queue.put_nowait(task)
            return True
        except queue.Full:
            return False

    def _run_loop(self) -> None:
        while True:
            if self._stop.is_set() and self._queue.empty():
                return

            try:
                task = self._queue.get(timeout=0.2)
            except queue.Empty:
                continue

            if task is None:
                self._queue.task_done()
                continue

            try:
                now = time.monotonic()
                if task.next_retry_at > now:
                    if self._stop.is_set():
                        task.next_retry_at = 0.0
                    else:
                        delay = task.next_retry_at - now
                        try:
                            self._queue.put_nowait(task)
                        except queue.Full:
                            pass
                        else:
                            if delay > 0 and self._queue.qsize() == 1:
                                self._stop.wait(timeout=min(delay, 0.2))
                        continue

                request_json(
                    "POST",
                    task.url,
                    json_body=task.body,
                    auth_token=self._auth_token,
                    timeout_seconds=self._timeout_seconds,
                )
            except AnimusAPIError as e:
                self._last_error = e
                retryable = e.status == 0 or e.status == 429 or e.status >= 500
                task.attempts += 1
                if retryable and task.attempts <= self._max_retries and not self._stop.is_set():
                    backoff = min(10.0, 0.25 * (2 ** (task.attempts - 1)))
                    task.next_retry_at = time.monotonic() + backoff
                    try:
                        self._queue.put_nowait(task)
                    except queue.Full:
                        pass
            except Exception as e:  # noqa: BLE001 - never crash the training process on telemetry errors
                self._last_error = AnimusAPIError(0, "telemetry_logger_error", None, {"detail": str(e), "kind": task.kind})
            finally:
                self._queue.task_done()

    def __enter__(self) -> "RunTelemetryLogger":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:  # type: ignore[override]
        self.close(flush=True)
