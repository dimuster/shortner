package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"shortener/internal/service"
)

type LinkCreator interface {
	CreateLink(ctx context.Context, shortCode, originalURL string) error
}

type CreateLinkHandler struct {
	repo   LinkCreator
	logger *slog.Logger
}

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url"`
}

type CreateLinkResponse struct {
	ShortСode string `json:"short_code"`
}

func NewCreateLinkHandler(repo LinkCreator, logger *slog.Logger) *CreateLinkHandler {
	return &CreateLinkHandler{repo: repo, logger: logger}
}

func (c *CreateLinkHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Error("invalid json in request", "err", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.OriginalURL == "" {
		c.logger.Error("missing original_url")
		http.Error(w, "original_url is required", http.StatusBadRequest)
		return
	}

	raw := req.OriginalURL
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		c.logger.Error("invalid url", "url", req.OriginalURL)
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	code, err := c.createWithRetry(r.Context(), raw)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := &CreateLinkResponse{code}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (c *CreateLinkHandler) createWithRetry(ctx context.Context, originalURL string) (string, error) {
	for range 5 {
		code, err := service.GenerateCode(6)
		if err != nil {
			return "", err
		}
		if err := c.repo.CreateLink(ctx, code, originalURL); err == nil {
			return code, nil
		}
	}
	return "", errors.New("failed after 5 attempts")
}
