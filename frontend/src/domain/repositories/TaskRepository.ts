import { Task } from '../entities/Task';

export interface TaskRepository {
  getTasks(repositoryId: number): Promise<Task[]>;
  createTask(task: Omit<Task, 'id' | 'created_at' | 'updated_at'>): Promise<Task>;
  updateTask(id: string, updates: Partial<Task>): Promise<Task>;
  deleteTask(id: string): Promise<void>;
  executeTask(id: string): Promise<void>;
}