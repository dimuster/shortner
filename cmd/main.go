package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortener/internal/handler"
	"shortener/internal/repository"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dsn := getenv("DATABASE_URL", "postgres://app:app@localhost:5431/shortener")

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("failed to create db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Ждём пока postgres поднимется (актуально при docker-compose up)
	if err := waitForDB(pool, logger); err != nil {
		logger.Error("database not ready", "err", err)
		os.Exit(1)
	}

	repo := repository.NewLinkRepository(pool)
	redirectHandler := handler.NewRedirectHandler(repo, logger)
	createLinkHandler := handler.NewCreateLinkHandler(repo, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/{code}", redirectHandler.Redirect)
	r.Post("/links", createLinkHandler.CreateLink)

	addr := getenv("ADDR", "localhost:8080")
	logger.Info("starting server", "addr", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func waitForDB(pool *pgxpool.Pool, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		if err := pool.Ping(ctx); err == nil {
			return nil
		}
		logger.Info("waiting for database...")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
