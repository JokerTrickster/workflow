import { ApiDataSource } from './ApiDataSource';
import { UserDto } from '../dto/UserDto';
import { Repository } from '../../domain/entities/Repository';
import { Task } from '../../domain/entities/Task';
import { apiClient } from '../../infrastructure/api/ApiClient';

export class ApiDataSourceImpl implements ApiDataSource {
  async login(): Promise<string> {
    window.location.href = '/api/auth/github';
    return '';
  }

  async logout(): Promise<void> {
    await apiClient.post('/auth/logout');
    localStorage.removeItem('auth_token');
  }

  async getCurrentUser(): Promise<UserDto | null> {
    try {
      return await apiClient.get<UserDto>('/auth/me');
    } catch {
      return null;
    }
  }

  async getUserRepositories(): Promise<Repository[]> {
    return await apiClient.get<Repository[]>('/repos');
  }

  async cloneRepository(repoId: number): Promise<void> {
    await apiClient.post(`/repos/clone`, { repo_id: repoId });
  }

  async getRepositoryStatus(repoId: number): Promise<{ connected: boolean; localPath?: string }> {
    return await apiClient.get(`/repos/${repoId}/status`);
  }

  async getTasks(repositoryName: string): Promise<Task[]> {
    return await apiClient.get<Task[]>(`/api/v1/tasks?repository_name=${encodeURIComponent(repositoryName)}`);
  }

  async createTask(taskData: {
    tasks: string;
    repository_name: string;
    working_dir?: string;
    interactive?: boolean;
    cmd?: string;
    provider: 'claude' | 'codex' | 'cursor';
  }): Promise<Task> {
    return await apiClient.post<Task>('/api/v1/tasks', taskData);
  }

  async updateTask(id: string, updates: Partial<Task>): Promise<Task> {
    return await apiClient.put<Task>(`/api/v1/tasks/${id}`, updates);
  }

  async deleteTask(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/tasks/${id}`);
  }

  async executeTask(id: string): Promise<void> {
    await apiClient.post(`/api/v1/tasks/${id}/execute`);
  }
}