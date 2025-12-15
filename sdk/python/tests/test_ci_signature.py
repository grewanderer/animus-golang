import unittest

from animus_sdk.experiments import compute_ci_webhook_signature


class TestCISignature(unittest.TestCase):
    def test_compute_ci_webhook_signature_deterministic(self) -> None:
        secret = "test-secret"
        ts = "1734200000"
        body = b'{"context":{"sha":"abc"},"provider":"github_actions","run_id":"run-123"}'
        sig = compute_ci_webhook_signature(secret, ts, "POST", body)
        self.assertEqual(sig, "W8DZwxZW8jKJeHNF-6yDAZ1JgEH1JNiSrsipBOSvfJY")


if __name__ == "__main__":
    unittest.main()

