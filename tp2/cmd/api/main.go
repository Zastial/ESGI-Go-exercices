package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "mira/internal/http"
	"mira/internal/store"

	_ "mira/cmd/api/docs"
)

// @title        Mira API
// @version      1.0
// @description  API de gestion de notes (CRUD + recherche texte simple) du projet fil rouge Mira. Stockage en mémoire.

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	router := apihttp.NewRouter(store.NewMemoryStore(), logger)

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	srv := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
