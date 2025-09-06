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

  async getTasks(repositoryId: number): Promise<Task[]> {
    return await apiClient.get<Task[]>(`/tasks?repository_id=${repositoryId}`);
  }

  async createTask(task: Omit<Task, 'id' | 'created_at' | 'updated_at'>): Promise<Task> {
    return await apiClient.post<Task>('/tasks', task);
  }

  async updateTask(id: string, updates: Partial<Task>): Promise<Task> {
    return await apiClient.put<Task>(`/tasks/${id}`, updates);
  }

  async deleteTask(id: string): Promise<void> {
    await apiClient.delete(`/tasks/${id}`);
  }

  async executeTask(id: string): Promise<void> {
    await apiClient.post(`/tasks/${id}/execute`);
  }
}