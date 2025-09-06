import { UserDto } from '../dto/UserDto';
import { Repository } from '../../domain/entities/Repository';
import { Task } from '../../domain/entities/Task';

export interface ApiDataSource {
  // Auth
  login(): Promise<string>;
  logout(): Promise<void>;
  getCurrentUser(): Promise<UserDto | null>;
  
  // Repositories
  getUserRepositories(): Promise<Repository[]>;
  cloneRepository(repoId: number): Promise<void>;
  getRepositoryStatus(repoId: number): Promise<{ connected: boolean; localPath?: string }>;
  
  // Tasks
  getTasks(repositoryId: number): Promise<Task[]>;
  createTask(task: Omit<Task, 'id' | 'created_at' | 'updated_at'>): Promise<Task>;
  updateTask(id: string, updates: Partial<Task>): Promise<Task>;
  deleteTask(id: string): Promise<void>;
  executeTask(id: string): Promise<void>;
}