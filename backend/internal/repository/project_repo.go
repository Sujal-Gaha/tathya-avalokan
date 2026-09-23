package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// CreateProject inserts a new project record and returns the response DTO.
func (r *ProjectRepository) CreateProject(ctx context.Context, req models.ProjectCreate) (*models.ProjectResponse, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	query := `INSERT INTO projects (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, id, req.Name, req.Description, now, now)
	if err != nil {
		return nil, err
	}

	return &models.ProjectResponse{
		ID:             id,
		Name:           req.Name,
		Description:    req.Description,
		InstancesCount: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// ListProjects retrieves all projects ordered by created_at DESC with live instances_count.
func (r *ProjectRepository) ListProjects(ctx context.Context) ([]models.ProjectResponse, error) {
	query := `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at, COUNT(d.id) AS instances_count
		FROM projects p
		LEFT JOIN database_instances d ON p.id = d.project_id
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.ProjectResponse, 0)
	for rows.Next() {
		var p models.ProjectResponse
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &desc, &p.CreatedAt, &p.UpdatedAt, &p.InstancesCount); err != nil {
			return nil, err
		}
		if desc.Valid {
			p.Description = &desc.String
		}
		items = append(items, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetProjectByID retrieves a single project with nested database instances.
func (r *ProjectRepository) GetProjectByID(ctx context.Context, id string) (*models.ProjectResponse, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM projects WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var p models.ProjectResponse
	var desc sql.NullString
	if err := row.Scan(&p.ID, &p.Name, &desc, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if desc.Valid {
		p.Description = &desc.String
	}

	// Query child instances
	instQuery := `
		SELECT id, project_id, name, driver_type, host, port, database_name, username, encrypted_credentials, ssl_mode, is_read_only, created_at, updated_at
		FROM database_instances
		WHERE project_id = ?
		ORDER BY created_at ASC
	`
	instRows, err := r.db.QueryContext(ctx, instQuery, id)
	if err != nil {
		return nil, err
	}
	defer instRows.Close()

	instances := make([]models.DatabaseInstanceResponse, 0)
	for instRows.Next() {
		var inst models.DatabaseInstance
		var host, username sql.NullString
		var port sql.NullInt64
		if err := instRows.Scan(
			&inst.ID,
			&inst.ProjectID,
			&inst.Name,
			&inst.DriverType,
			&host,
			&port,
			&inst.DatabaseName,
			&username,
			&inst.EncryptedCredentials,
			&inst.SSLMode,
			&inst.IsReadOnly,
			&inst.CreatedAt,
			&inst.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if host.Valid {
			inst.Host = &host.String
		}
		if port.Valid {
			pInt := int(port.Int64)
			inst.Port = &pInt
		}
		if username.Valid {
			inst.Username = &username.String
		}

		isPasswordSet := crypto.HasConfiguredCredentials(inst.EncryptedCredentials)
		instances = append(instances, inst.ToResponse(isPasswordSet))
	}

	if err := instRows.Err(); err != nil {
		return nil, err
	}

	p.InstancesCount = len(instances)
	p.Instances = &instances
	return &p, nil
}

// UpdateProject updates non-nil fields of a project.
func (r *ProjectRepository) UpdateProject(ctx context.Context, id string, req models.ProjectUpdate) (*models.ProjectResponse, error) {
	current, err := r.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Description != nil {
		current.Description = req.Description
	}
	current.UpdatedAt = time.Now().UTC()

	query := `UPDATE projects SET name = ?, description = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.ExecContext(ctx, query, current.Name, current.Description, current.UpdatedAt, id)
	if err != nil {
		return nil, err
	}

	return current, nil
}

// DeleteProject deletes a project by ID (cascades database_instances).
func (r *ProjectRepository) DeleteProject(ctx context.Context, id string) (bool, error) {
	query := `DELETE FROM projects WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
