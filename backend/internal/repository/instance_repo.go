package repository

import (
	"context"
	"database/sql"
	"time"

	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/models"

	"github.com/google/uuid"
)

type InstanceRepository struct {
	db *sql.DB
}

func NewInstanceRepository(db *sql.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// CreateInstance registers a new database instance under a project.
func (r *InstanceRepository) CreateInstance(
	ctx context.Context,
	projectID string,
	req models.DatabaseInstanceCreate,
	encryptedCreds string,
) (*models.DatabaseInstanceResponse, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	driverType := req.DriverType
	if driverType == "" {
		driverType = models.DriverPostgres
	}

	sslMode := "prefer"
	if req.SSLMode != nil && *req.SSLMode != "" {
		sslMode = *req.SSLMode
	}

	isReadOnly := false
	if req.IsReadOnly != nil {
		isReadOnly = *req.IsReadOnly
	}

	query := `
		INSERT INTO database_instances (
			id, project_id, name, driver_type, host, port, database_name,
			username, encrypted_credentials, ssl_mode, is_read_only, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		id,
		projectID,
		req.Name,
		string(driverType),
		req.Host,
		req.Port,
		req.DatabaseName,
		req.Username,
		encryptedCreds,
		sslMode,
		isReadOnly,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}

	isPasswordSet := crypto.HasConfiguredCredentials(encryptedCreds)

	return &models.DatabaseInstanceResponse{
		ID:            id,
		ProjectID:     projectID,
		Name:          req.Name,
		DriverType:    driverType,
		Host:          req.Host,
		Port:          req.Port,
		DatabaseName:  req.DatabaseName,
		Username:      req.Username,
		SSLMode:       sslMode,
		IsReadOnly:    isReadOnly,
		IsPasswordSet: isPasswordSet,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// GetInstanceByID retrieves a single instance domain model including raw encrypted credentials.
func (r *InstanceRepository) GetInstanceByID(ctx context.Context, id string) (*models.DatabaseInstance, error) {
	query := `
		SELECT id, project_id, name, driver_type, host, port, database_name,
		       username, encrypted_credentials, ssl_mode, is_read_only, created_at, updated_at
		FROM database_instances
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var inst models.DatabaseInstance
	var host, username sql.NullString
	var port sql.NullInt64
	if err := row.Scan(
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
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

	return &inst, nil
}

// UpdateInstance partially updates an instance, modifying fields and encrypted credentials as needed.
func (r *InstanceRepository) UpdateInstance(
	ctx context.Context,
	id string,
	req models.DatabaseInstanceUpdate,
	newEncryptedCreds *string,
) (*models.DatabaseInstanceResponse, error) {
	current, err := r.GetInstanceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.DriverType != nil {
		current.DriverType = *req.DriverType
	}
	if req.Host != nil {
		current.Host = req.Host
	}
	if req.Port != nil {
		current.Port = req.Port
	}
	if req.DatabaseName != nil {
		current.DatabaseName = *req.DatabaseName
	}
	if req.Username != nil {
		current.Username = req.Username
	}
	if req.SSLMode != nil {
		current.SSLMode = *req.SSLMode
	}
	if req.IsReadOnly != nil {
		current.IsReadOnly = *req.IsReadOnly
	}
	if newEncryptedCreds != nil {
		current.EncryptedCredentials = *newEncryptedCreds
	}
	current.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE database_instances SET
			name = ?, driver_type = ?, host = ?, port = ?, database_name = ?,
			username = ?, encrypted_credentials = ?, ssl_mode = ?, is_read_only = ?, updated_at = ?
		WHERE id = ?
	`
	_, err = r.db.ExecContext(
		ctx,
		query,
		current.Name,
		string(current.DriverType),
		current.Host,
		current.Port,
		current.DatabaseName,
		current.Username,
		current.EncryptedCredentials,
		current.SSLMode,
		current.IsReadOnly,
		current.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, err
	}

	isPasswordSet := crypto.HasConfiguredCredentials(current.EncryptedCredentials)
	res := current.ToResponse(isPasswordSet)
	return &res, nil
}

// DeleteInstance removes an instance by ID.
func (r *InstanceRepository) DeleteInstance(ctx context.Context, id string) (bool, error) {
	query := `DELETE FROM database_instances WHERE id = ?`
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
