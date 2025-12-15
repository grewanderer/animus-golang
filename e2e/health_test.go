//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestServices_Healthz(t *testing.T) {
	infra := ensureInfra(t)
	services := []struct {
		name      string
		path      string
		addrEnv   string
		healthURL string
		readyURL  string
	}{
		{name: "gateway", path: "./gateway", addrEnv: "GATEWAY_HTTP_ADDR"},
		{name: "dataset-registry", path: "./dataset-registry", addrEnv: "DATASET_REGISTRY_HTTP_ADDR"},
		{name: "quality", path: "./quality", addrEnv: "QUALITY_HTTP_ADDR"},
		{name: "experiments", path: "./experiments", addrEnv: "EXPERIMENTS_HTTP_ADDR"},
		{name: "lineage", path: "./lineage", addrEnv: "LINEAGE_HTTP_ADDR"},
		{name: "audit", path: "./audit", addrEnv: "AUDIT_HTTP_ADDR"},
	}

	repoRoot := repoRoot(t)
	tmpDir := t.TempDir()

	for _, svc := range services {
		svc := svc
		t.Run(svc.name, func(t *testing.T) {
			addr := freeAddr(t)
			svc.healthURL = fmt.Sprintf("http://%s/healthz", addr)
			svc.readyURL = fmt.Sprintf("http://%s/readyz", addr)

			bin := filepath.Join(tmpDir, fmt.Sprintf("%s.bin", svc.name))
			build := exec.Command("go", "build", "-o", bin, svc.path)
			build.Dir = repoRoot
			buildOut, err := build.CombinedOutput()
			if err != nil {
				t.Fatalf("go build %s: %v\n%s", svc.path, err, string(buildOut))
			}

			var out bytes.Buffer
			cmd := exec.Command(bin)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("%s=%s", svc.addrEnv, addr),
				"DATABASE_URL="+infra.databaseURL,
				"ANIMUS_INTERNAL_AUTH_SECRET="+infra.internalAuthSecret,
				"ANIMUS_CI_WEBHOOK_SECRET="+infra.ciWebhookSecret,
				"ANIMUS_MINIO_ENDPOINT="+infra.minioEndpoint,
				"ANIMUS_MINIO_ACCESS_KEY="+infra.minioAccessKey,
				"ANIMUS_MINIO_SECRET_KEY="+infra.minioSecretKey,
				"ANIMUS_MINIO_USE_SSL=false",
				"ANIMUS_MINIO_BUCKET_DATASETS="+infra.minioBucketDatasets,
				"ANIMUS_MINIO_BUCKET_ARTIFACTS="+infra.minioBucketArtifacts,
				"AUTH_MODE=dev",
				"AUTH_SESSION_COOKIE_SECURE=false",
			)
			cmd.Stdout = &out
			cmd.Stderr = &out

			if err := cmd.Start(); err != nil {
				t.Fatalf("start %s: %v", svc.name, err)
			}
			t.Cleanup(func() { stopProcess(t, cmd, &out) })

			waitHTTP200(t, svc.readyURL)

			resp, err := http.Get(svc.healthURL)
			if err != nil {
				t.Fatalf("GET %s: %v\n%s", svc.healthURL, err, out.String())
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET %s status=%d, want 200\n%s", svc.healthURL, resp.StatusCode, out.String())
			}
		})
	}
}

type infraConfig struct {
	databaseURL          string
	minioEndpoint        string
	minioAccessKey       string
	minioSecretKey       string
	minioBucketDatasets  string
	minioBucketArtifacts string
	internalAuthSecret   string
	ciWebhookSecret      string
}

func ensureInfra(t *testing.T) infraConfig {
	t.Helper()

	if v := strings.TrimSpace(os.Getenv("ANIMUS_E2E_DATABASE_URL")); v != "" {
		minioEndpoint := strings.TrimSpace(os.Getenv("ANIMUS_E2E_MINIO_ENDPOINT"))
		if minioEndpoint == "" {
			t.Fatalf("ANIMUS_E2E_MINIO_ENDPOINT is required when ANIMUS_E2E_DATABASE_URL is set")
		}
		minioAccessKey := strings.TrimSpace(os.Getenv("ANIMUS_E2E_MINIO_ACCESS_KEY"))
		minioSecretKey := strings.TrimSpace(os.Getenv("ANIMUS_E2E_MINIO_SECRET_KEY"))
		if minioAccessKey == "" || minioSecretKey == "" {
			t.Fatalf("ANIMUS_E2E_MINIO_ACCESS_KEY and ANIMUS_E2E_MINIO_SECRET_KEY are required when using external minio")
		}

		bucketDatasets := strings.TrimSpace(os.Getenv("ANIMUS_E2E_MINIO_BUCKET_DATASETS"))
		if bucketDatasets == "" {
			bucketDatasets = "datasets"
		}
		bucketArtifacts := strings.TrimSpace(os.Getenv("ANIMUS_E2E_MINIO_BUCKET_ARTIFACTS"))
		if bucketArtifacts == "" {
			bucketArtifacts = "artifacts"
		}

		internalSecret := strings.TrimSpace(os.Getenv("ANIMUS_E2E_INTERNAL_AUTH_SECRET"))
		if internalSecret == "" {
			internalSecret = randomSecret(t, 32)
		}
		ciSecret := strings.TrimSpace(os.Getenv("ANIMUS_E2E_CI_WEBHOOK_SECRET"))
		if ciSecret == "" {
			ciSecret = randomSecret(t, 32)
		}

		return infraConfig{
			databaseURL:          v,
			minioEndpoint:        minioEndpoint,
			minioAccessKey:       minioAccessKey,
			minioSecretKey:       minioSecretKey,
			minioBucketDatasets:  bucketDatasets,
			minioBucketArtifacts: bucketArtifacts,
			internalAuthSecret:   internalSecret,
			ciWebhookSecret:      ciSecret,
		}
	}

	if strings.TrimSpace(os.Getenv("ANIMUS_E2E_SKIP_DOCKER")) == "1" {
		t.Skip("docker infra is disabled (ANIMUS_E2E_SKIP_DOCKER=1); set ANIMUS_E2E_DATABASE_URL + ANIMUS_E2E_MINIO_* to run")
	}

	if !commandExists("docker") {
		t.Skip("docker not found; set ANIMUS_E2E_DATABASE_URL + ANIMUS_E2E_MINIO_* to run without docker")
	}

	dbContainer := fmt.Sprintf("animus-e2e-postgres-%d", time.Now().UnixNano())
	minioContainer := fmt.Sprintf("animus-e2e-minio-%d", time.Now().UnixNano())

	dbURL := startPostgres(t, dbContainer)
	minioEndpoint := startMinIO(t, minioContainer)

	const (
		minioRootUser     = "animus-root"
		minioRootPassword = "animus-root-password"
	)
	const (
		bucketDatasets  = "datasets"
		bucketArtifacts = "artifacts"
	)

	waitMinIOReady(t, minioEndpoint, 20*time.Second)
	ensureMinIOBuckets(t, minioEndpoint, minioRootUser, minioRootPassword, bucketDatasets, bucketArtifacts)

	waitPostgresReady(t, dbURL, 20*time.Second)

	return infraConfig{
		databaseURL:          dbURL,
		minioEndpoint:        minioEndpoint,
		minioAccessKey:       minioRootUser,
		minioSecretKey:       minioRootPassword,
		minioBucketDatasets:  bucketDatasets,
		minioBucketArtifacts: bucketArtifacts,
		internalAuthSecret:   randomSecret(t, 32),
		ciWebhookSecret:      randomSecret(t, 32),
	}
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func randomSecret(t *testing.T, n int) string {
	t.Helper()
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func startPostgres(t *testing.T, name string) string {
	t.Helper()

	image := strings.TrimSpace(os.Getenv("ANIMUS_E2E_POSTGRES_IMAGE"))
	if image == "" {
		image = "postgres:14-alpine"
	}

	run := exec.Command("docker", "run",
		"-d",
		"--rm",
		"--name", name,
		"-e", "POSTGRES_USER=animus",
		"-e", "POSTGRES_PASSWORD=animus",
		"-e", "POSTGRES_DB=animus",
		"-p", "127.0.0.1:0:5432",
		image,
	)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run postgres: %v\n%s", err, string(out))
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })

	port := dockerHostPort(t, name, "5432/tcp")
	return fmt.Sprintf("postgres://animus:animus@127.0.0.1:%d/animus?sslmode=disable", port)
}

func startMinIO(t *testing.T, name string) string {
	t.Helper()

	image := strings.TrimSpace(os.Getenv("ANIMUS_E2E_MINIO_IMAGE"))
	if image == "" {
		image = "minio/minio@sha256:14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e"
	}

	run := exec.Command("docker", "run",
		"-d",
		"--rm",
		"--name", name,
		"-e", "MINIO_ROOT_USER=animus-root",
		"-e", "MINIO_ROOT_PASSWORD=animus-root-password",
		"-p", "127.0.0.1:0:9000",
		image,
		"server", "/data", "--console-address", ":9001",
	)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run minio: %v\n%s", err, string(out))
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })

	port := dockerHostPort(t, name, "9000/tcp")
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func dockerHostPort(t *testing.T, containerName string, portProto string) int {
	t.Helper()

	cmd := exec.Command("docker", "inspect", "-f", fmt.Sprintf("{{(index (index .NetworkSettings.Ports %q) 0).HostPort}}", portProto), containerName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker inspect %s: %v\n%s", containerName, err, string(out))
	}
	portRaw := strings.TrimSpace(string(out))
	port, err := strconv.Atoi(portRaw)
	if err != nil || port <= 0 {
		t.Fatalf("invalid port mapping for %s (%s): %q", containerName, portProto, portRaw)
	}
	return port
}

func waitPostgresReady(t *testing.T, databaseURL string, timeout time.Duration) {
	t.Helper()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		pingCtx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
		err := db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return
		}

		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for postgres: %v", err)
		case <-ticker.C:
		}
	}
}

func waitMinIOReady(t *testing.T, endpoint string, timeout time.Duration) {
	t.Helper()

	url := fmt.Sprintf("http://%s/minio/health/ready", endpoint)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(timeout)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for minio %s", url)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func ensureMinIOBuckets(t *testing.T, endpoint, accessKey, secretKey, datasetsBucket, artifactsBucket string) {
	t.Helper()

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
		Region: "us-east-1",
	})
	if err != nil {
		t.Fatalf("minio client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ensure := func(bucket string) {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			t.Fatalf("bucket exists %s: %v", bucket, err)
		}
		if exists {
			return
		}
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			t.Fatalf("make bucket %s: %v", bucket, err)
		}
	}

	ensure(datasetsBucket)
	ensure(artifactsBucket)
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

func freeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func waitHTTP200(t *testing.T, url string) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(8 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %s", url)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func stopProcess(t *testing.T, cmd *exec.Cmd, out *bytes.Buffer) {
	t.Helper()

	if cmd.Process == nil {
		return
	}

	_ = cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	case err := <-done:
		if err != nil {
			body := out.String()
			if len(body) > 8000 {
				body = body[len(body)-8000:]
			}
			t.Fatalf("process exit: %v\n%s", err, body)
		}
	}
}
