export interface ColumnMeta {
  name: string;
  type: string;
}

export interface QueryRequest {
  sql: string;
  limit?: number;
  offset?: number;
  timeout_seconds?: number;
}

export interface QueryResponse {
  columns: ColumnMeta[];
  rows: Record<string, unknown>[];
  rows_affected: number;
  execution_time_ms: number;
  has_more: boolean;
}

export interface ConnectionTestResult {
  connected: boolean;
  latency_ms: number;
  server_version: string | null;
  message: string;
}

export interface EditorTab {
  id: string;
  instanceId: string;
  title: string;
  sql: string;
  isExecuting?: boolean;
  result?: QueryResponse | null;
  error?: string | null;
}
