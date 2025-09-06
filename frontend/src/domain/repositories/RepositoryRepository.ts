import { Repository } from '../entities/Repository';

export interface RepositoryRepository {
  getUserRepositories(): Promise<Repository[]>;
  cloneRepository(repoId: number): Promise<void>;
  getRepositoryStatus(repoId: number): Promise<{ connected: boolean; localPath?: string }>;
}