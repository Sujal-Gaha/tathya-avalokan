package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/models"
	"tathya-avalokan/backend/internal/repository"
	"tathya-avalokan/backend/internal/response"
)

type InstancesHandler struct {
	pRepo *repository.ProjectRepository
	iRepo *repository.InstanceRepository
}

func NewInstancesHandler(pRepo *repository.ProjectRepository, iRepo *repository.InstanceRepository) *InstancesHandler {
	return &InstancesHandler{
		pRepo: pRepo,
		iRepo: iRepo,
	}
}

// CreateInstance handles POST /api/v1/projects/{project_id}/instances
func (h *InstancesHandler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	if projectID == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Project ID parameter is required", nil)
		return
	}

	// Verify parent project exists
	project, err := h.pRepo.GetProjectByID(r.Context(), projectID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}
	if project == nil {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Parent project with id '%s' does not exist", projectID), nil)
		return
	}

	var req models.DatabaseInstanceCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request JSON payload", nil)
		return
	}

	if req.Name == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Instance name is required", nil)
		return
	}
	if req.DatabaseName == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Database name is required", nil)
		return
	}

	var encryptedBlob string
	hasPassword := req.Password != nil && strings.TrimSpace(*req.Password) != ""
	hasURL := req.ConnectionURL != nil && strings.TrimSpace(*req.ConnectionURL) != ""

	if hasPassword || hasURL {
		creds := map[string]string{
			"password":       "",
			"connection_url": "",
		}
		if req.Password != nil {
			creds["password"] = *req.Password
		}
		if req.ConnectionURL != nil {
			creds["connection_url"] = *req.ConnectionURL
		}

		blob, err := crypto.EncryptCredentials(creds)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Failed to encrypt credentials", nil)
			return
		}
		encryptedBlob = blob
	}

	instance, err := h.iRepo.CreateInstance(r.Context(), projectID, req, encryptedBlob)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	response.SendJSON(w, http.StatusCreated, instance, nil)
}

// GetInstance handles GET /api/v1/instances/{id}
func (h *InstancesHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
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

	isPasswordSet := crypto.HasConfiguredCredentials(inst.EncryptedCredentials)
	res := inst.ToResponse(isPasswordSet)
	response.SendJSON(w, http.StatusOK, res, nil)
}

// UpdateInstance handles PATCH /api/v1/instances/{id}
func (h *InstancesHandler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Instance ID parameter is required", nil)
		return
	}

	var rawMap map[string]any
	if err := json.NewDecoder(r.Body).Decode(&rawMap); err != nil || len(rawMap) == 0 {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "No update fields provided in request body", nil)
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

	var newEncryptedBlob *string
	hasPassword := false
	hasURL := false
	var newPassword, newURL string

	if val, ok := rawMap["password"]; ok {
		hasPassword = true
		if val != nil {
			newPassword = fmt.Sprintf("%v", val)
		}
	}
	if val, ok := rawMap["connection_url"]; ok {
		hasURL = true
		if val != nil {
			newURL = fmt.Sprintf("%v", val)
		}
	}

	if hasPassword || hasURL {
		existingCreds, _ := crypto.DecryptCredentials(inst.EncryptedCredentials)
		if existingCreds == nil {
			existingCreds = make(map[string]string)
		}
		if hasPassword {
			existingCreds["password"] = newPassword
		}
		if hasURL {
			existingCreds["connection_url"] = newURL
		}
		blob, err := crypto.EncryptCredentials(existingCreds)
		if err != nil {
			response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Failed to re-encrypt credentials", nil)
			return
		}
		newEncryptedBlob = &blob
	}

	var req models.DatabaseInstanceUpdate
	rawBytes, _ := json.Marshal(rawMap)
	_ = json.Unmarshal(rawBytes, &req)

	updated, err := h.iRepo.UpdateInstance(r.Context(), id, req, newEncryptedBlob)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}

	response.SendJSON(w, http.StatusOK, updated, nil)
}

// DeleteInstance handles DELETE /api/v1/instances/{id}
func (h *InstancesHandler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendError(w, http.StatusBadRequest, response.ErrCodeValidation, "Instance ID parameter is required", nil)
		return
	}

	deleted, err := h.iRepo.DeleteInstance(r.Context(), id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, response.ErrCodeInternal, err.Error(), nil)
		return
	}
	if !deleted {
		response.SendError(w, http.StatusNotFound, response.ErrCodeNotFound, fmt.Sprintf("Database instance with id '%s' not found", id), nil)
		return
	}

	response.SendJSON(w, http.StatusOK, map[string]any{
		"deleted": true,
		"id":      id,
	}, nil)
}
