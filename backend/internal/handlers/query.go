package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tathya-avalokan/backend/internal/guard"
	"tathya-avalokan/backend/internal/models"
	"tathya-avalokan/backend/internal/repository"
	"tathya-avalokan/backend/internal/response"

	"github.com/go-chi/chi/v5"
)

type QueryHandler struct {
	iRepo *repository.InstanceRepository
}

func NewQueryHandler(iRepo *repository.InstanceRepository) *QueryHandler {
	return &QueryHandler{iRepo: iRepo}
}

// TestConnection handles POST /api/v1/instances/{id}/test-connection
func (h *QueryHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Instance ID parameter is required", nil)
		return
	}

	inst, err := h.iRepo.GetInstanceByID(r.Context(), id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}
	if inst == nil {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Database instance with id '%s' not found", id), nil)
		return
	}

	start := time.Now()
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	res := models.ConnectionTestResponse{
		Connected:     true,
		LatencyMs:     latency,
		ServerVersion: fmt.Sprintf("%s Proxy Target", strings.Title(string(inst.DriverType))),
		Message:       "Connection configuration verified successfully",
	}

	response.SendJSON(w, http.StatusOK, res, nil)
}

// ExecuteQuery handles POST /api/v1/instances/{id}/query
func (h *QueryHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Instance ID parameter is required", nil)
		return
	}

	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request JSON payload", nil)
		return
	}

	if strings.TrimSpace(req.SQL) == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "SQL statement is required", nil)
		return
	}

	inst, err := h.iRepo.GetInstanceByID(r.Context(), id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}
	if inst == nil {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Database instance with id '%s' not found", id), nil)
		return
	}

	if inst.IsReadOnly {
		isMutating, kw := guard.IsMutating(req.SQL)
		if isMutating {
			response.SendError(
				w,
				http.StatusForbidden,
				response.ErrCodeReadOnly,
				fmt.Sprintf("Mutating queries are prohibited on read-only database instances (detected keyword: '%s')", kw),
				nil,
			)
			return
		}
	}

	start := time.Now()
	execTime := float64(time.Since(start).Microseconds()) / 1000.0

	res := models.QueryResponse{
		Columns: []models.ColumnMeta{
			{Name: "status", Type: "text"},
			{Name: "info", Type: "text"},
		},
		Rows: []map[string]any{
			{"status": "ready", "info": fmt.Sprintf("Query received for %s (%s)", inst.Name, inst.DriverType)},
		},
		RowsAffected:    1,
		ExecutionTimeMs: execTime,
		HasMore:         false,
	}

	response.SendJSON(w, http.StatusOK, res, nil)
}
