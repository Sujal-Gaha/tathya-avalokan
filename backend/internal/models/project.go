package models

import (
	"time"
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type ProjectUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ProjectResponse struct {
	ID             string                      `json:"id"`
	Name           string                      `json:"name"`
	Description    *string                     `json:"description"`
	InstancesCount int                         `json:"instances_count"`
	Instances      *[]DatabaseInstanceResponse `json:"instances,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}
