import { request } from './api-client';
import { CreateProjectInput, Project, UpdateProjectInput } from '../types/project';

export const projectService = {
  async list(): Promise<Project[]> {
    return request<Project[]>('/projects');
  },

  async get(id: string): Promise<Project> {
    return request<Project>(`/projects/${id}`);
  },

  async create(input: CreateProjectInput): Promise<Project> {
    return request<Project>('/projects', {
      method: 'POST',
      body: JSON.stringify(input),
    });
  },

  async update(id: string, input: UpdateProjectInput): Promise<Project> {
    return request<Project>(`/projects/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    });
  },

  async delete(id: string): Promise<{ deleted: boolean; id: string }> {
    return request<{ deleted: boolean; id: string }>(`/projects/${id}`, {
      method: 'DELETE',
    });
  },
};
