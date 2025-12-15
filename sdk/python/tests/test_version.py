import unittest

from animus_sdk import __version__


class TestVersion(unittest.TestCase):
    def test_version_nonempty(self) -> None:
        self.assertTrue(isinstance(__version__, str))
        self.assertTrue(len(__version__) > 0)


if __name__ == "__main__":
    unittest.main()

