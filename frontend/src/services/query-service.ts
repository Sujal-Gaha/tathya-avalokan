import { request } from './api-client';
import { ConnectionTestResult, QueryRequest, QueryResponse } from '../types/query';

export const queryService = {
  async testConnection(instanceId: string): Promise<ConnectionTestResult> {
    return request<ConnectionTestResult>(`/instances/${instanceId}/test-connection`, {
      method: 'POST',
    });
  },

  async executeQuery(instanceId: string, payload: QueryRequest): Promise<QueryResponse> {
    return request<QueryResponse>(`/instances/${instanceId}/query`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },
};
