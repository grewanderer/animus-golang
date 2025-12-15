import unittest

from animus_sdk.git import git_metadata_from_env, normalize_repo


class TestGitMetadata(unittest.TestCase):
    def test_github_env(self) -> None:
        env = {
            "GITHUB_SHA": "abcdef123456",
            "GITHUB_REPOSITORY": "animus-labs/animus-go",
            "GITHUB_REF": "refs/heads/main",
        }
        meta = git_metadata_from_env(env)
        self.assertIsNotNone(meta)
        assert meta is not None
        self.assertEqual(meta.source, "github")
        self.assertEqual(meta.commit, "abcdef123456")
        self.assertEqual(meta.ref, "refs/heads/main")
        self.assertEqual(meta.repo, "github.com/animus-labs/animus-go")

    def test_gitlab_env(self) -> None:
        env = {
            "CI_COMMIT_SHA": "deadbeef",
            "CI_SERVER_HOST": "gitlab.example.local",
            "CI_PROJECT_PATH": "ml/team-a/project-x",
            "CI_COMMIT_REF_NAME": "main",
        }
        meta = git_metadata_from_env(env)
        self.assertIsNotNone(meta)
        assert meta is not None
        self.assertEqual(meta.source, "gitlab")
        self.assertEqual(meta.commit, "deadbeef")
        self.assertEqual(meta.ref, "main")
        self.assertEqual(meta.repo, "gitlab.example.local/ml/team-a/project-x")

    def test_normalize_repo_scp(self) -> None:
        self.assertEqual(normalize_repo("git@github.com:org/repo.git"), "github.com/org/repo")

    def test_normalize_repo_https(self) -> None:
        self.assertEqual(normalize_repo("https://github.com/org/repo.git"), "github.com/org/repo")


if __name__ == "__main__":
    unittest.main()

