package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/opendash-project/opendash/internal/api"
	"github.com/opendash-project/opendash/internal/auth"
	"github.com/opendash-project/opendash/internal/config"
	"github.com/opendash-project/opendash/internal/runtime"
	"github.com/opendash-project/opendash/internal/runtime/docker"
	"github.com/opendash-project/opendash/internal/runtime/noop"
	"github.com/opendash-project/opendash/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	setupLogger(cfg.LogLevel)

	slog.Info("starting opendash", "addr", cfg.Addr())

	var rt runtime.Runtime = docker.NewWithRoot(cfg.AppsRoot)
	if !cfg.UseDockerRuntime {
		rt = noop.New()
	}

	s, err := store.OpenSQLite(context.Background(), cfg.DatabasePath, cfg.DemoMode)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer s.Close()

	// Recover any operations left running after a previous crash so that they
	// do not block new lifecycle requests forever.
	if manager, ok := any(s).(interface{ RecoverOperations(context.Context) error }); ok {
		if err := manager.RecoverOperations(context.Background()); err != nil {
			slog.Error("failed to recover operations", "error", err)
		}
	}

	h := api.NewHandlers(s, rt)
	h.ConfigureAuth(auth.New(s.DB(), cfg.SessionLifetime), cfg.AuthEnabled, cfg.SecureCookies, cfg.SessionLifetime)
	h.ConfigureCatalog(cfg.CatalogRoot, cfg.AppsRoot)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	h.SetLifecycle(ctx)

	router := api.NewRouter(h, cfg.RequestBodyLimit)

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	h.Wait()
	slog.Info("shutdown complete")
}

func setupLogger(level string) {
	var lv slog.Level
	switch level {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
	slog.SetDefault(logger)
}
