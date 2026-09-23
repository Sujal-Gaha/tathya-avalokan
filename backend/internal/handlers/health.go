package handlers

import (
	"net/http"
	"time"

	"tathya-avalokan/backend/internal/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"status":            "healthy",
		"version":           "0.1.0",
		"metadata_database": "connected",
		"metadata_storage":  "sqlite+modernc",
		"supported_drivers": []string{"postgresql", "mysql", "sqlite"},
	}

	metadata := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	response.SendJSON(w, http.StatusOK, data, metadata)
}
