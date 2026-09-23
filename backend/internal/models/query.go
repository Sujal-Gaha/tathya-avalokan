package models

type QueryRequest struct {
	SQL            string `json:"sql"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

type ColumnMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type QueryResponse struct {
	Columns         []ColumnMeta     `json:"columns"`
	Rows            []map[string]any `json:"rows"`
	RowsAffected    int              `json:"rows_affected"`
	ExecutionTimeMs float64          `json:"execution_time_ms"`
	HasMore         bool             `json:"has_more"`
}

type ConnectionTestResponse struct {
	Connected     bool    `json:"connected"`
	LatencyMs     float64 `json:"latency_ms"`
	ServerVersion string  `json:"server_version"`
	Message       string  `json:"message"`
}
