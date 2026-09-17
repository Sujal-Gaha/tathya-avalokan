export type DriverType = 'postgresql' | 'mysql' | 'sqlite';

export interface DatabaseInstance {
  id: string;
  project_id: string;
  name: string;
  driver_type: DriverType;
  host: string | null;
  port: number | null;
  database_name: string;
  username: string | null;
  ssl_mode: string;
  is_read_only: boolean;
  is_password_set: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateInstanceInput {
  name: string;
  driver_type: DriverType;
  host?: string;
  port?: number;
  database_name: string;
  username?: string;
  password?: string;
  connection_url?: string;
  ssl_mode?: string;
  is_read_only?: boolean;
}

export interface UpdateInstanceInput {
  name?: string;
  host?: string;
  port?: number;
  database_name?: string;
  username?: string;
  password?: string;
  connection_url?: string;
  ssl_mode?: string;
  is_read_only?: boolean;
}
