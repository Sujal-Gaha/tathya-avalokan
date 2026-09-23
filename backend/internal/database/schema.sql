PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS ix_projects_name ON projects (name);

CREATE TABLE IF NOT EXISTS database_instances (
    id VARCHAR(36) PRIMARY KEY,
    project_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    driver_type VARCHAR(20) NOT NULL,
    host VARCHAR(255),
    port INTEGER,
    database_name VARCHAR(100) NOT NULL,
    username VARCHAR(100),
    encrypted_credentials TEXT NOT NULL,
    ssl_mode VARCHAR(30) NOT NULL DEFAULT 'prefer',
    is_read_only BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS ix_database_instances_project_id ON database_instances (project_id);
