import { ApiResponseEnvelope } from '../types/api';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export class ApiError extends Error {
  code: string;
  details?: Record<string, unknown> | null;

  constructor(message: string, code: string, details?: Record<string, unknown> | null) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.details = details;
  }
}

export async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const response = await fetch(url, { ...options, headers });
  const envelope: ApiResponseEnvelope<T> = await response.json();

  if (!response.ok || envelope.error) {
    const error = envelope.error || {
      code: 'HTTP_ERROR',
      message: response.statusText || 'An unexpected error occurred',
      details: null,
    };
    throw new ApiError(error.message, error.code, error.details);
  }

  if (envelope.data === null) {
    throw new ApiError('Received empty payload from server', 'EMPTY_PAYLOAD');
  }

  return envelope.data;
}
