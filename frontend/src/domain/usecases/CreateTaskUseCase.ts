import { Task } from '../entities/Task';
import { TaskRepository } from '../repositories/TaskRepository';

export class CreateTaskUseCase {
  constructor(private taskRepository: TaskRepository) {}

  async execute(taskData: Omit<Task, 'id' | 'created_at' | 'updated_at'>): Promise<Task> {
    return await this.taskRepository.createTask(taskData);
  }
}