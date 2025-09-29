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
  getTasks(repositoryName: string): Promise<Task[]>;
  createTask(taskData: {
    tasks: string;
    repository_name: string;
    working_dir?: string;
    interactive?: boolean;
    cmd?: string;
    provider: 'claude' | 'codex' | 'cursor';
  }): Promise<Task>;
  updateTask(id: string, updates: Partial<Task>): Promise<Task>;
  deleteTask(id: string): Promise<void>;
  executeTask(id: string): Promise<void>;
}