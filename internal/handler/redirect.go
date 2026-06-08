package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"shortener/internal/repository"
)

// LinkResolver — интерфейс, чтобы потом легко подменить на версию с Redis
type LinkResolver interface {
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type RedirectHandler struct {
	repo   LinkResolver
	logger *slog.Logger
}

func NewRedirectHandler(repo LinkResolver, logger *slog.Logger) *RedirectHandler {
	return &RedirectHandler{repo: repo, logger: logger}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "code")

	originalURL, err := h.repo.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("failed to get original url", "short_code", shortCode, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}
