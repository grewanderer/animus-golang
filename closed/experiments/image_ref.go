package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/runtimeexec"
)

func parseImageDigestFromRef(ref string) (string, bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", false
	}
	if isSHA256Digest(ref) {
		return strings.ToLower(strings.TrimSpace(ref)), true
	}
	at := strings.LastIndex(ref, "@")
	if at <= 0 || at == len(ref)-1 {
		return "", false
	}
	if strings.TrimSpace(ref[:at]) == "" {
		return "", false
	}
	digest := strings.ToLower(strings.TrimSpace(ref[at+1:]))
	if !isSHA256Digest(digest) {
		return "", false
	}
	return digest, true
}

func isSHA256Digest(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(value, "sha256:") {
		return false
	}
	hexPart := strings.TrimPrefix(value, "sha256:")
	if len(hexPart) != 64 {
		return false
	}
	_, err := hex.DecodeString(hexPart)
	return err == nil
}

type dockerImageIDResolver = runtimeexec.DockerImageIDResolver

func resolveImageForExecutor(ctx context.Context, executorKind string, executor any, imageRef string) (executionRef string, imageDigest string, err error) {
	imageRef = strings.TrimSpace(imageRef)
	if imageRef == "" {
		return "", "", errors.New("image ref is required")
	}

	switch strings.TrimSpace(executorKind) {
	case "docker":
		if parsed, ok := parseImageDigestFromRef(imageRef); ok {
			if strings.Contains(imageRef, "@") || isSHA256Digest(imageRef) {
				return imageRef, parsed, nil
			}
		}
		resolver, ok := executor.(dockerImageIDResolver)
		if !ok {
			return "", "", errors.New("docker executor does not support image resolution")
		}
		id, err := resolver.ResolveImageID(ctx, imageRef)
		if err != nil {
			if errors.Is(err, runtimeexec.ErrImageRefNotFound) {
				return "", "", err
			}
			return "", "", fmt.Errorf("resolve docker image id: %w", err)
		}
		if !isSHA256Digest(id) {
			return "", "", fmt.Errorf("unexpected docker image id: %q", id)
		}
		return id, strings.ToLower(strings.TrimSpace(id)), nil
	case "kubernetes_job":
		digest, ok := parseImageDigestFromRef(imageRef)
		if !ok || !strings.Contains(imageRef, "@") {
			return "", "", runtimeexec.ErrImageRefDigestRequired
		}
		return imageRef, digest, nil
	default:
		return "", "", fmt.Errorf("unsupported executor kind: %q", strings.TrimSpace(executorKind))
	}
}
