import { User } from '../../domain/entities/User';
import { UserRepository } from '../../domain/repositories/UserRepository';
import { ApiDataSource } from '../datasources/ApiDataSource';

export class UserRepositoryImpl implements UserRepository {
  constructor(private apiDataSource: ApiDataSource) {}

  async getCurrentUser(): Promise<User | null> {
    const userDto = await this.apiDataSource.getCurrentUser();
    if (!userDto) return null;
    
    return {
      id: userDto.id,
      login: userDto.login,
      name: userDto.name,
      email: userDto.email,
      avatar_url: userDto.avatar_url,
      html_url: userDto.html_url,
      created_at: userDto.created_at,
      updated_at: userDto.updated_at,
    };
  }

  async login(): Promise<string> {
    return await this.apiDataSource.login();
  }

  async logout(): Promise<void> {
    await this.apiDataSource.logout();
  }
}