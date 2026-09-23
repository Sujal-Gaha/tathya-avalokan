package models

import (
	"time"
)

type DriverType string

const (
	DriverPostgres DriverType = "postgresql"
	DriverMySQL    DriverType = "mysql"
	DriverSQLite   DriverType = "sqlite"
)

type DatabaseInstance struct {
	ID                   string     `json:"id"`
	ProjectID            string     `json:"project_id"`
	Name                 string     `json:"name"`
	DriverType           DriverType `json:"driver_type"`
	Host                 *string    `json:"host"`
	Port                 *int       `json:"port"`
	DatabaseName         string     `json:"database_name"`
	Username             *string    `json:"username"`
	EncryptedCredentials string     `json:"-"`
	SSLMode              string     `json:"ssl_mode"`
	IsReadOnly           bool       `json:"is_read_only"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type DatabaseInstanceCreate struct {
	Name          string     `json:"name"`
	DriverType    DriverType `json:"driver_type"`
	Host          *string    `json:"host"`
	Port          *int       `json:"port"`
	DatabaseName  string     `json:"database_name"`
	Username      *string    `json:"username"`
	Password      *string    `json:"password"`
	ConnectionURL *string    `json:"connection_url"`
	SSLMode       *string    `json:"ssl_mode"`
	IsReadOnly    *bool      `json:"is_read_only"`
}

type DatabaseInstanceUpdate struct {
	Name          *string     `json:"name"`
	DriverType    *DriverType `json:"driver_type"`
	Host          *string     `json:"host"`
	Port          *int        `json:"port"`
	DatabaseName  *string     `json:"database_name"`
	Username      *string     `json:"username"`
	Password      *string     `json:"password"`
	ConnectionURL *string     `json:"connection_url"`
	SSLMode       *string     `json:"ssl_mode"`
	IsReadOnly    *bool       `json:"is_read_only"`
}

type DatabaseInstanceResponse struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	Name          string     `json:"name"`
	DriverType    DriverType `json:"driver_type"`
	Host          *string    `json:"host"`
	Port          *int       `json:"port"`
	DatabaseName  string     `json:"database_name"`
	Username      *string    `json:"username"`
	SSLMode       string     `json:"ssl_mode"`
	IsReadOnly    bool       `json:"is_read_only"`
	IsPasswordSet bool       `json:"is_password_set"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (inst *DatabaseInstance) ToResponse(isPasswordSet bool) DatabaseInstanceResponse {
	return DatabaseInstanceResponse{
		ID:            inst.ID,
		ProjectID:     inst.ProjectID,
		Name:          inst.Name,
		DriverType:    inst.DriverType,
		Host:          inst.Host,
		Port:          inst.Port,
		DatabaseName:  inst.DatabaseName,
		Username:      inst.Username,
		SSLMode:       inst.SSLMode,
		IsReadOnly:    inst.IsReadOnly,
		IsPasswordSet: isPasswordSet,
		CreatedAt:     inst.CreatedAt,
		UpdatedAt:     inst.UpdatedAt,
	}
}
