package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"mira/internal/config"
	apihttp "mira/internal/http"
	"mira/internal/migrate"
	"mira/internal/service"
	"mira/internal/store"
)

func main() {
	config.LoadEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	pool := mustConnect(ctx, logger)
	defer pool.Close()

	if err := migrate.Apply(ctx, pool); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	repo := store.NewPostgresStore(pool)
	queue := service.NewQueue(100)
	processor := service.NewProcessor(repo)
	workers := service.NewWorkerPool(queue, processor, logger, 3, 4*time.Second)

	workerCtx, stopWorkers := context.WithCancel(context.Background())
	workers.Start(workerCtx)
	defer func() {
		queue.Close()
		stopWorkers()
		workers.Wait()
	}()

	server := &http.Server{
		Addr:              serverAddr(),
		Handler:           apihttp.NewRouter(repo, queue, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	waitSignal()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func mustConnect(ctx context.Context, logger *slog.Logger) *pgxpool.Pool {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	return pool
}

func serverAddr() string {
	if v := os.Getenv("PORT"); v != "" {
		return ":" + v
	}
	return ":8080"
}

func waitSignal() {
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()
}
