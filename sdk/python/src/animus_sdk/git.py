from __future__ import annotations

from dataclasses import dataclass
import os
import subprocess
from typing import Mapping
from urllib.parse import urlparse


@dataclass(frozen=True)
class GitMetadata:
    repo: str
    commit: str
    ref: str = ""
    source: str = ""


def normalize_repo(value: str) -> str:
    value = value.strip()
    if not value:
        return ""

    if value.startswith("git@") and ":" in value:
        user_host, path = value.split(":", 1)
        host = user_host.split("@", 1)[1]
        path = path.lstrip("/")
        if path.endswith(".git"):
            path = path[:-4]
        return f"{host}/{path}"

    if "://" in value:
        parsed = urlparse(value)
        host = (parsed.hostname or "").strip()
        path = (parsed.path or "").lstrip("/")
        if path.endswith(".git"):
            path = path[:-4]
        if host and path:
            return f"{host}/{path}"
        return value.rstrip("/")

    value = value.rstrip("/")
    if value.endswith(".git"):
        value = value[:-4]
    return value


def git_metadata_from_env(env: Mapping[str, str] | None = None) -> GitMetadata | None:
    if env is None:
        env = os.environ

    github_sha = env.get("GITHUB_SHA", "").strip()
    github_repo = env.get("GITHUB_REPOSITORY", "").strip()
    if github_sha and github_repo:
        ref = (env.get("GITHUB_REF", "") or env.get("GITHUB_HEAD_REF", "")).strip()
        repo = normalize_repo(f"github.com/{github_repo}")
        return GitMetadata(repo=repo, commit=github_sha, ref=ref, source="github")

    gitlab_sha = env.get("CI_COMMIT_SHA", "").strip()
    gitlab_host = env.get("CI_SERVER_HOST", "").strip()
    gitlab_path = env.get("CI_PROJECT_PATH", "").strip()
    if gitlab_sha and gitlab_host and gitlab_path:
        ref = (env.get("CI_COMMIT_REF_NAME", "") or env.get("CI_COMMIT_BRANCH", "")).strip()
        repo = normalize_repo(f"{gitlab_host}/{gitlab_path}")
        return GitMetadata(repo=repo, commit=gitlab_sha, ref=ref, source="gitlab")

    jenkins_sha = env.get("GIT_COMMIT", "").strip()
    jenkins_url = env.get("GIT_URL", "").strip()
    if jenkins_sha and jenkins_url:
        ref = env.get("GIT_BRANCH", "").strip()
        repo = normalize_repo(jenkins_url)
        return GitMetadata(repo=repo, commit=jenkins_sha, ref=ref, source="jenkins")

    return None


def git_metadata_from_git_cli(cwd: str | None = None) -> GitMetadata | None:
    def run(args: list[str]) -> str:
        out = subprocess.check_output(args, cwd=cwd, stderr=subprocess.DEVNULL)
        return out.decode("utf-8").strip()

    try:
        commit = run(["git", "rev-parse", "HEAD"])
    except Exception:
        return None

    repo = ""
    ref = ""
    try:
        repo = normalize_repo(run(["git", "remote", "get-url", "origin"]))
    except Exception:
        repo = ""
    try:
        ref = run(["git", "rev-parse", "--abbrev-ref", "HEAD"])
        if ref == "HEAD":
            ref = ""
    except Exception:
        ref = ""

    return GitMetadata(repo=repo, commit=commit, ref=ref, source="git")


def get_git_metadata(env: Mapping[str, str] | None = None, cwd: str | None = None) -> GitMetadata | None:
    meta = git_metadata_from_env(env)
    if meta is not None:
        return meta
    return git_metadata_from_git_cli(cwd=cwd)

