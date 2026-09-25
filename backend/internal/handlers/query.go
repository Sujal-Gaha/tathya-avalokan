package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"tathya-avalokan/backend/internal/models"
	"tathya-avalokan/backend/internal/proxy"
	"tathya-avalokan/backend/internal/repository"
	"tathya-avalokan/backend/internal/response"

	"github.com/go-chi/chi/v5"
)

type QueryHandler struct {
	iRepo       *repository.InstanceRepository
	proxyEngine *proxy.ProxyEngine
}

func NewQueryHandler(iRepo *repository.InstanceRepository, proxyEngine *proxy.ProxyEngine) *QueryHandler {
	return &QueryHandler{
		iRepo:       iRepo,
		proxyEngine: proxyEngine,
	}
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

	if h.proxyEngine == nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Database proxy engine is uninitialized", nil)
		return
	}

	res, err := h.proxyEngine.TestConnection(r.Context(), inst)
	if err != nil {
		host := ""
		if inst.Host != nil {
			host = *inst.Host
		}
		port := 0
		if inst.Port != nil {
			port = *inst.Port
		}
		status, code, msg, details := proxy.ClassifyConnectionError(err, host, port, string(inst.DriverType))
		response.SendError(w, status, code, msg, details)
		return
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

	if h.proxyEngine == nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Database proxy engine is uninitialized", nil)
		return
	}

	res, err := h.proxyEngine.ExecuteQuery(r.Context(), inst, req)
	if err != nil {
		var roErr *proxy.ReadOnlyViolationError
		if errors.As(err, &roErr) {
			response.SendError(w, http.StatusForbidden, response.ErrCodeReadOnly, roErr.Error(), nil)
			return
		}

		status, code, msg, details := proxy.ClassifyExecutionError(err)
		response.SendError(w, status, code, msg, details)
		return
	}

	response.SendJSON(w, http.StatusOK, res, nil)
}
