import { request } from './api-client';
import { CreateInstanceInput, DatabaseInstance, UpdateInstanceInput } from '../types/instance';

export const instanceService = {
  async create(projectId: string, input: CreateInstanceInput): Promise<DatabaseInstance> {
    return request<DatabaseInstance>(`/projects/${projectId}/instances`, {
      method: 'POST',
      body: JSON.stringify(input),
    });
  },

  async get(id: string): Promise<DatabaseInstance> {
    return request<DatabaseInstance>(`/instances/${id}`);
  },

  async update(id: string, input: UpdateInstanceInput): Promise<DatabaseInstance> {
    return request<DatabaseInstance>(`/instances/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    });
  },

  async delete(id: string): Promise<{ deleted: boolean; id: string }> {
    return request<{ deleted: boolean; id: string }>(`/instances/${id}`, {
      method: 'DELETE',
    });
  },
};
