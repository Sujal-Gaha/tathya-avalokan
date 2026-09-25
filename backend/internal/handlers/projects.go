package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"tathya-avalokan/backend/internal/models"
	"tathya-avalokan/backend/internal/proxy"
	"tathya-avalokan/backend/internal/repository"
	"tathya-avalokan/backend/internal/response"

	"github.com/go-chi/chi/v5"
)

type ProjectsHandler struct {
	repo        *repository.ProjectRepository
	proxyEngine *proxy.ProxyEngine
}

func NewProjectsHandler(repo *repository.ProjectRepository, proxyEngine *proxy.ProxyEngine) *ProjectsHandler {
	return &ProjectsHandler{
		repo:        repo,
		proxyEngine: proxyEngine,
	}
}

// CreateProject handles POST /api/v1/projects
func (h *ProjectsHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req models.ProjectCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request JSON payload", nil)
		return
	}

	if req.Name == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Project name is required", nil)
		return
	}

	proj, err := h.repo.CreateProject(r.Context(), req)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	response.SendJSON(w, http.StatusCreated, proj, nil)
}

// ListProjects handles GET /api/v1/projects
func (h *ProjectsHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.repo.ListProjects(r.Context())
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	response.SendJSON(w, http.StatusOK, projects, map[string]any{
		"total_count": len(projects),
	})
}

// GetProject handles GET /api/v1/projects/{id}
func (h *ProjectsHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Project id parameter is required", nil)
		return
	}

	project, err := h.repo.GetProjectByID(r.Context(), id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	if project == nil {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Project with id '%s' was not found", id), nil)
		return
	}

	response.SendJSON(w, http.StatusOK, project, nil)
}

// UpdateProject handles PATCH /api/v1/projects/{id}
func (h *ProjectsHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Project id parameter is required", nil)
		return
	}

	var rawMap map[string]any
	if err := json.NewDecoder(r.Body).Decode(&rawMap); err != nil || len(rawMap) == 0 {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "No update fields provided in request body", nil)
		return
	}

	var req models.ProjectUpdate
	rawBytes, _ := json.Marshal(rawMap)
	_ = json.Unmarshal(rawBytes, &req)

	updated, err := h.repo.UpdateProject(r.Context(), id, req)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	if updated == nil {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Project with id '%s' was not found", id), nil)
		return
	}

	response.SendJSON(w, http.StatusOK, updated, nil)
}

// DeleteProject handles DELETE /api/v1/projects/{id}
func (h *ProjectsHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Project id parameter is required", nil)
		return
	}

	var childIDs []string
	if h.proxyEngine != nil {
		var err error
		childIDs, err = h.repo.GetInstanceIDsByProjectID(r.Context(), id)
		if err != nil {
			log.Printf("[WARN] Failed to fetch child instance IDs for eviction: %v", err)
		}
	}

	deleted, err := h.repo.DeleteProject(r.Context(), id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	if !deleted {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Project with id '%s' was not found", id), nil)
		return
	}

	if h.proxyEngine != nil && len(childIDs) > 0 {
		h.proxyEngine.EvictMultiple(childIDs)
	}

	response.SendJSON(w, http.StatusOK, map[string]any{
		"deleted": true,
		"id":      id,
	}, nil)
}
