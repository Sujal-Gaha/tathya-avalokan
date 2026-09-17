export interface ApiErrorDetail {
  code: string;
  message: string;
  details?: Record<string, unknown> | null;
}

export interface ApiResponseEnvelope<T> {
  data: T | null;
  error: ApiErrorDetail | null;
  metadata: Record<string, unknown>;
}
