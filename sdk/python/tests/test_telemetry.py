import queue
import threading
import unittest
from unittest import mock

from animus_sdk.errors import AnimusAPIError
from animus_sdk.telemetry import RunTelemetryLogger


class TestRunTelemetryLogger(unittest.TestCase):
    def test_log_metrics_sends_payload(self) -> None:
        with mock.patch("animus_sdk.telemetry.request_json") as request_json:
            request_json.return_value = None
            logger = RunTelemetryLogger(gateway_url="http://example", run_id="run-1", auth_token="token", timeout_seconds=0.1)
            try:
                ok = logger.log_metrics(step=3, metrics={"loss": 1.25}, metadata={"split": "train"})
                self.assertTrue(ok)
                self.assertTrue(logger.flush(timeout_seconds=2.0))
            finally:
                logger.close(flush=True, timeout_seconds=2.0)

            request_json.assert_called()
            method, url = request_json.call_args.args[:2]
            self.assertEqual(method, "POST")
            self.assertEqual(url, "http://example/api/experiments/experiment-runs/run-1/metrics")
            self.assertEqual(
                request_json.call_args.kwargs.get("json_body"),
                {"step": 3, "metrics": {"loss": 1.25}, "metadata": {"split": "train"}},
            )
            self.assertEqual(request_json.call_args.kwargs.get("auth_token"), "token")

    def test_log_status_and_progress_are_events(self) -> None:
        with mock.patch("animus_sdk.telemetry.request_json") as request_json:
            request_json.return_value = None
            logger = RunTelemetryLogger(gateway_url="http://example", run_id="run-2", auth_token=None, timeout_seconds=0.1)
            try:
                self.assertTrue(logger.log_status(status="running"))
                self.assertTrue(logger.log_progress(step=7, total_steps=10, percent=0.7, message="epoch 1"))
                self.assertTrue(logger.flush(timeout_seconds=2.0))
            finally:
                logger.close(flush=True, timeout_seconds=2.0)

            urls = [c.args[1] for c in request_json.call_args_list]
            self.assertIn("http://example/api/experiments/experiment-runs/run-2/events", urls)
            self.assertEqual(len(urls), 2)

    def test_close_does_not_raise_when_queue_full(self) -> None:
        with mock.patch("animus_sdk.telemetry.request_json") as request_json:
            request_json.return_value = None
            logger = RunTelemetryLogger(gateway_url="http://example", run_id="run-3", timeout_seconds=0.1, max_queue=1)
            try:
                original_put = logger._queue.put_nowait

                def put_nowait(item):  # type: ignore[no-untyped-def]
                    if item is None:
                        raise queue.Full
                    return original_put(item)

                with mock.patch.object(logger._queue, "put_nowait", side_effect=put_nowait):
                    logger.close(flush=False, timeout_seconds=1.0)
            finally:
                if logger._thread.is_alive():
                    logger.close(flush=False, timeout_seconds=1.0)

    def test_worker_survives_queue_full_on_retry_requeue(self) -> None:
        drop_seen = threading.Event()
        put_calls = 0
        put_lock = threading.Lock()

        def request_side_effect(*args, **kwargs):  # type: ignore[no-untyped-def]
            if request_side_effect.calls == 0:
                request_side_effect.calls += 1
                raise AnimusAPIError(500, "server_error", None)
            request_side_effect.calls += 1
            return None

        request_side_effect.calls = 0  # type: ignore[attr-defined]

        with mock.patch("animus_sdk.telemetry.request_json", side_effect=request_side_effect):
            logger = RunTelemetryLogger(gateway_url="http://example", run_id="run-4", timeout_seconds=0.1, max_retries=1)
            try:
                original_put = logger._queue.put_nowait

                def put_nowait(item):  # type: ignore[no-untyped-def]
                    nonlocal put_calls
                    with put_lock:
                        put_calls += 1
                        n = put_calls
                    if n == 3:
                        drop_seen.set()
                        raise queue.Full
                    return original_put(item)

                with mock.patch.object(logger._queue, "put_nowait", side_effect=put_nowait):
                    self.assertTrue(logger.log_metric(step=0, name="loss", value=1.0))
                    self.assertTrue(drop_seen.wait(timeout=2.0))
                    self.assertTrue(logger.log_metric(step=1, name="loss", value=0.5))
                    self.assertTrue(logger.flush(timeout_seconds=2.0))
            finally:
                logger.close(flush=True, timeout_seconds=2.0)

            self.assertGreaterEqual(request_side_effect.calls, 2)  # type: ignore[attr-defined]


if __name__ == "__main__":
    unittest.main()

