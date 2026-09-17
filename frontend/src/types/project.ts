import { DatabaseInstance } from './instance';

export interface Project {
  id: string;
  name: string;
  description: string | null;
  instances_count: number;
  instances?: DatabaseInstance[];
  created_at: string;
  updated_at: string;
}

export interface CreateProjectInput {
  name: string;
  description?: string;
}

export interface UpdateProjectInput {
  name?: string;
  description?: string;
}
