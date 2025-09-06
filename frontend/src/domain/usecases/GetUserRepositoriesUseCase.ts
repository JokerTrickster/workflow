import { Repository } from '../entities/Repository';
import { RepositoryRepository } from '../repositories/RepositoryRepository';

export class GetUserRepositoriesUseCase {
  constructor(private repositoryRepository: RepositoryRepository) {}

  async execute(): Promise<Repository[]> {
    return await this.repositoryRepository.getUserRepositories();
  }
}