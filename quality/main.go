package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/animus-labs/animus-go/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/internal/platform/auth"
	"github.com/animus-labs/animus-go/internal/platform/env"
	"github.com/animus-labs/animus-go/internal/platform/httpserver"
	"github.com/animus-labs/animus-go/internal/platform/objectstore"
	"github.com/animus-labs/animus-go/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := env.String("QUALITY_HTTP_ADDR", ":8082")
	shutdownTimeout, err := env.Duration("QUALITY_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		logger.Error("invalid env", "error", err)
		os.Exit(2)
	}

	dbCfg, err := postgres.ConfigFromEnv()
	if err != nil {
		logger.Error("invalid database config", "error", err)
		os.Exit(2)
	}
	db, err := postgres.Open(ctx, dbCfg)
	if err != nil {
		logger.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	storeCfg, err := objectstore.ConfigFromEnv()
	if err != nil {
		logger.Error("invalid object store config", "error", err)
		os.Exit(2)
	}
	storeClient, err := objectstore.NewMinIOClient(storeCfg)
	if err != nil {
		logger.Error("object store client init failed", "error", err)
		os.Exit(2)
	}
	startupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := objectstore.EnsureBuckets(startupCtx, storeClient, storeCfg); err != nil {
		cancel()
		logger.Error("object store unavailable", "error", err)
		os.Exit(1)
	}
	cancel()

	internalAuthSecret := env.String("ANIMUS_INTERNAL_AUTH_SECRET", "")
	headersAuth, err := auth.NewGatewayHeadersAuthenticator(internalAuthSecret)
	if err != nil {
		logger.Error("invalid internal auth config", "error", err)
		os.Exit(2)
	}

	authorizer := auth.MethodRoleAuthorizer()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", httpserver.Healthz("quality"))
	mux.HandleFunc(
		"/readyz",
		httpserver.ReadyzWithChecks(
			"quality",
			httpserver.ReadinessCheck{
				Name: "postgres",
				Check: func(ctx context.Context) error {
					checkCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
					defer cancel()
					return db.PingContext(checkCtx)
				},
			},
			httpserver.ReadinessCheck{
				Name: "minio",
				Check: func(ctx context.Context) error {
					checkCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
					defer cancel()
					return objectstore.CheckBuckets(checkCtx, storeClient, storeCfg)
				},
			},
		),
	)

	api := newQualityAPI(logger, db, storeClient, storeCfg)
	api.register(mux)

	handler := auth.Middleware{
		Logger:        logger,
		Authenticator: headersAuth,
		Authorize:     authorizer,
		Audit: func(ctx context.Context, event auth.DenyEvent) error {
			auditCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
			defer cancel()
			return auditlog.InsertAuthDeny(auditCtx, db, "quality", event)
		},
		SkipPrefixes: []string{"/healthz", "/readyz"},
	}.Wrap(mux)

	cfg := httpserver.Config{
		Service:         "quality",
		Addr:            addr,
		ShutdownTimeout: shutdownTimeout,
	}

	if err := httpserver.Run(ctx, logger, cfg, httpserver.Wrap(logger, "quality", handler)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
