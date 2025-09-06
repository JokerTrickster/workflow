import { User } from '../entities/User';

export interface UserRepository {
  getCurrentUser(): Promise<User | null>;
  login(): Promise<string>;
  logout(): Promise<void>;
}