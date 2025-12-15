package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/animus-labs/animus-go/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/internal/platform/auth"
	"github.com/animus-labs/animus-go/internal/platform/env"
	"github.com/animus-labs/animus-go/internal/platform/httpserver"
	"github.com/animus-labs/animus-go/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := env.String("GATEWAY_HTTP_ADDR", ":8080")
	shutdownTimeout, err := env.Duration("GATEWAY_SHUTDOWN_TIMEOUT", 10*time.Second)
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

	authCfg, err := auth.ConfigFromEnv()
	if err != nil {
		logger.Error("invalid auth config", "error", err)
		os.Exit(2)
	}

	internalAuthSecret := env.String("ANIMUS_INTERNAL_AUTH_SECRET", "")
	if strings.TrimSpace(internalAuthSecret) == "" {
		logger.Error("missing internal auth secret", "env", "ANIMUS_INTERNAL_AUTH_SECRET")
		os.Exit(2)
	}

	var authenticator auth.Authenticator
	var oidcService *auth.OIDCService
	switch authCfg.Mode {
	case auth.ModeDev:
		authenticator = auth.NewDevAuthenticator(authCfg)
	case auth.ModeOIDC:
		svc, err := auth.NewOIDCService(ctx, authCfg)
		if err != nil {
			logger.Error("oidc init failed", "error", err)
			os.Exit(1)
		}
		oidcService = svc
		authenticator = svc
	case auth.ModeDisabled:
		authenticator = nil
	default:
		logger.Error("unsupported auth mode", "mode", authCfg.Mode)
		os.Exit(2)
	}

	authorizer := auth.MethodRoleAuthorizer()
	auditFn := func(ctx context.Context, event auth.DenyEvent) error {
		auditCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
		defer cancel()
		return auditlog.InsertAuthDeny(auditCtx, db, "gateway", event)
	}
	protected := func(handler http.Handler) http.Handler {
		if authenticator == nil {
			return handler
		}
		return auth.Middleware{
			Logger:        logger,
			Authenticator: authenticator,
			Authorize:     authorizer,
			Audit:         auditFn,
		}.Wrap(handler)
	}
	sessionProtected := func(handler http.Handler) http.Handler {
		if authenticator == nil {
			return handler
		}
		return auth.Middleware{
			Logger:        logger,
			Authenticator: authenticator,
			Audit:         auditFn,
		}.Wrap(handler)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", httpserver.Healthz("gateway"))
	mux.HandleFunc(
		"/readyz",
		httpserver.ReadyzWithChecks(
			"gateway",
			httpserver.ReadinessCheck{
				Name: "postgres",
				Check: func(ctx context.Context) error {
					checkCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
					defer cancel()
					return db.PingContext(checkCtx)
				},
			},
		),
	)

	if oidcService != nil {
		mux.HandleFunc("/auth/logout", oidcService.LogoutHandler())
		mux.Handle("/auth/session", sessionProtected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, _ := auth.IdentityFromContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"subject": identity.Subject,
				"email":   identity.Email,
				"roles":   identity.Roles,
			})
		})))

		if err := authCfg.ValidateForLogin(); err == nil {
			login, err := oidcService.LoginHandler()
			if err != nil {
				logger.Error("oidc login handler init failed", "error", err)
				os.Exit(2)
			}
			callback, err := oidcService.CallbackHandler()
			if err != nil {
				logger.Error("oidc callback handler init failed", "error", err)
				os.Exit(2)
			}
			mux.HandleFunc("/auth/login", login)
			mux.HandleFunc("/auth/callback", callback)
		} else {
			mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotImplemented)
				_, _ = w.Write([]byte("{\"error\":\"login_not_configured\"}\n"))
			})
			mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotImplemented)
				_, _ = w.Write([]byte("{\"error\":\"login_not_configured\"}\n"))
			})
		}
	} else if authCfg.Mode == auth.ModeDisabled {
		mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
		})
		mux.HandleFunc("/auth/session", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{\"mode\":\"disabled\"}\n"))
		})
	} else {
		mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
		})
		mux.Handle("/auth/session", sessionProtected(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, _ := auth.IdentityFromContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"subject": identity.Subject,
				"email":   identity.Email,
				"roles":   identity.Roles,
			})
		})))
	}

	datasetRegistryProxy, err := newReverseProxy(logger, internalAuthSecret, env.String("DATASET_REGISTRY_BASE_URL", "http://localhost:8081"))
	if err != nil {
		logger.Error("proxy init failed", "service", "dataset-registry", "error", err)
		os.Exit(2)
	}
	qualityProxy, err := newReverseProxy(logger, internalAuthSecret, env.String("QUALITY_BASE_URL", "http://localhost:8082"))
	if err != nil {
		logger.Error("proxy init failed", "service", "quality", "error", err)
		os.Exit(2)
	}
	experimentsProxy, err := newReverseProxy(logger, internalAuthSecret, env.String("EXPERIMENTS_BASE_URL", "http://localhost:8083"))
	if err != nil {
		logger.Error("proxy init failed", "service", "experiments", "error", err)
		os.Exit(2)
	}
	lineageProxy, err := newReverseProxy(logger, internalAuthSecret, env.String("LINEAGE_BASE_URL", "http://localhost:8084"))
	if err != nil {
		logger.Error("proxy init failed", "service", "lineage", "error", err)
		os.Exit(2)
	}
	auditProxy, err := newReverseProxy(logger, internalAuthSecret, env.String("AUDIT_BASE_URL", "http://localhost:8085"))
	if err != nil {
		logger.Error("proxy init failed", "service", "audit", "error", err)
		os.Exit(2)
	}

	mux.Handle("/api/dataset-registry/", protected(http.StripPrefix("/api/dataset-registry", datasetRegistryProxy)))
	mux.Handle("/api/quality/", protected(http.StripPrefix("/api/quality", qualityProxy)))
	mux.Handle("/api/experiments/", protected(http.StripPrefix("/api/experiments", experimentsProxy)))
	mux.Handle("/api/lineage/", protected(http.StripPrefix("/api/lineage", lineageProxy)))
	mux.Handle("/api/audit/", protected(http.StripPrefix("/api/audit", auditProxy)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpserver.Healthz("gateway")(w, r)
	})

	cfg := httpserver.Config{
		Service:         "gateway",
		Addr:            addr,
		ShutdownTimeout: shutdownTimeout,
	}

	if err := httpserver.Run(ctx, logger, cfg, httpserver.Wrap(logger, "gateway", mux)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func newReverseProxy(logger *slog.Logger, internalAuthSecret string, target string) (http.Handler, error) {
	upstream, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	if upstream.Scheme == "" || upstream.Host == "" {
		return nil, fmt.Errorf("invalid upstream url: %q", target)
	}

	proxy := httputil.NewSingleHostReverseProxy(upstream)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		director(r)
		r.Header.Del(auth.HeaderSubject)
		r.Header.Del(auth.HeaderEmail)
		r.Header.Del(auth.HeaderRoles)
		r.Header.Del(auth.HeaderInternalAuthTimestamp)
		r.Header.Del(auth.HeaderInternalAuthSignature)
		if identity, ok := auth.IdentityFromContext(r.Context()); ok {
			r.Header.Set(auth.HeaderSubject, identity.Subject)
			if identity.Email != "" {
				r.Header.Set(auth.HeaderEmail, identity.Email)
			}
			roles := strings.Join(identity.Roles, ",")
			if roles != "" {
				r.Header.Set(auth.HeaderRoles, roles)
			}
			ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
			sig, err := auth.ComputeInternalAuthSignature(
				internalAuthSecret,
				ts,
				r.Method,
				r.URL.Path,
				r.Header.Get("X-Request-Id"),
				identity.Subject,
				identity.Email,
				roles,
			)
			if err == nil {
				r.Header.Set(auth.HeaderInternalAuthTimestamp, ts)
				r.Header.Set(auth.HeaderInternalAuthSignature, sig)
			}
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("proxy error", "request_id", r.Header.Get("X-Request-Id"), "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("{\"error\":\"bad_gateway\"}\n"))
	}
	return proxy, nil
}
