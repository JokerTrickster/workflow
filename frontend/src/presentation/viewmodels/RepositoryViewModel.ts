'use client';

import { useState, useCallback } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { GetUserRepositoriesUseCase } from '../../domain/usecases/GetUserRepositoriesUseCase';

export class RepositoryViewModel {
  private _repositories: Repository[] = [];
  private _loading = false;
  private _error: string | null = null;
  private _listeners: (() => void)[] = [];

  constructor(private getUserRepositoriesUseCase: GetUserRepositoriesUseCase) {}

  get repositories(): Repository[] {
    return this._repositories;
  }

  get loading(): boolean {
    return this._loading;
  }

  get error(): string | null {
    return this._error;
  }

  subscribe(listener: () => void) {
    this._listeners.push(listener);
    return () => {
      this._listeners = this._listeners.filter(l => l !== listener);
    };
  }

  private notifyListeners() {
    this._listeners.forEach(listener => listener());
  }

  async loadRepositories() {
    try {
      this._loading = true;
      this._error = null;
      this.notifyListeners();

      this._repositories = await this.getUserRepositoriesUseCase.execute();
    } catch (error) {
      this._error = error instanceof Error ? error.message : 'Unknown error occurred';
    } finally {
      this._loading = false;
      this.notifyListeners();
    }
  }
}

export const useRepositoryViewModel = (viewModel: RepositoryViewModel) => {
  const [, forceUpdate] = useState({});
  
  const triggerUpdate = useCallback(() => {
    forceUpdate({});
  }, []);

  useState(() => {
    const unsubscribe = viewModel.subscribe(triggerUpdate);
    return () => unsubscribe();
  });

  return {
    repositories: viewModel.repositories,
    loading: viewModel.loading,
    error: viewModel.error,
    loadRepositories: () => viewModel.loadRepositories(),
  };
};