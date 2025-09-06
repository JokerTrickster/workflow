'use client';

import { useEffect, useState } from 'react';
import { RepositoryCard } from '../presentation/components/RepositoryCard';
import { WorkspacePanel } from '../presentation/components/WorkspacePanel';
import { Button } from '../components/ui/button';
import { Github, Plus } from 'lucide-react';
import { Repository } from '../domain/entities/Repository';

// Mock data for development
const mockRepositories = [
  {
    id: 1,
    name: 'ai-git-workbench',
    full_name: 'captain/ai-git-workbench',
    description: 'AI-powered Git workbench for managing multiple repositories',
    html_url: 'https://github.com/captain/ai-git-workbench',
    clone_url: 'https://github.com/captain/ai-git-workbench.git',
    ssh_url: 'git@github.com:captain/ai-git-workbench.git',
    private: false,
    language: 'TypeScript',
    stargazers_count: 42,
    forks_count: 8,
    created_at: '2024-01-15T10:30:00Z',
    updated_at: '2024-09-06T08:15:00Z',
    pushed_at: '2024-09-06T08:15:00Z',
    default_branch: 'main',
    is_connected: false,
  },
  {
    id: 2,
    name: 'sample-project',
    full_name: 'captain/sample-project',
    description: 'A sample project for testing',
    html_url: 'https://github.com/captain/sample-project',
    clone_url: 'https://github.com/captain/sample-project.git',
    ssh_url: 'git@github.com:captain/sample-project.git',
    private: true,
    language: 'JavaScript',
    stargazers_count: 12,
    forks_count: 3,
    created_at: '2024-02-20T14:20:00Z',
    updated_at: '2024-09-05T16:45:00Z',
    pushed_at: '2024-09-05T16:45:00Z',
    default_branch: 'main',
    is_connected: true,
  },
];

export default function HomePage() {
  const [repositories, setRepositories] = useState(mockRepositories);
  const [isLoggedIn, setIsLoggedIn] = useState(true); // Mock login state
  const [selectedRepository, setSelectedRepository] = useState<Repository | null>(null);

  const handleConnectRepository = (repoId: number) => {
    setRepositories(prev => 
      prev.map(repo => 
        repo.id === repoId 
          ? { ...repo, is_connected: true }
          : repo
      )
    );
  };

  const handleOpenWorkspace = (repository: Repository) => {
    setSelectedRepository(repository);
  };

  const handleCloseWorkspace = () => {
    setSelectedRepository(null);
  };

  const handleLogin = () => {
    // Mock login
    setIsLoggedIn(true);
  };

  if (!isLoggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-6">
          <div className="space-y-2">
            <h1 className="text-4xl font-bold">AI Git Workbench</h1>
            <p className="text-muted-foreground">
              Manage your GitHub repositories with AI assistance
            </p>
          </div>
          <Button onClick={handleLogin} size="lg" className="gap-2">
            <Github className="h-5 w-5" />
            Login with GitHub
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Github className="h-8 w-8" />
              <h1 className="text-2xl font-bold">AI Git Workbench</h1>
            </div>
            <div className="flex items-center gap-4">
              <Button variant="outline" size="sm">
                Settings
              </Button>
              <Button variant="ghost" size="sm">
                Logout
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-3xl font-bold">Your Repositories</h2>
              <p className="text-muted-foreground">
                Connect repositories to start managing them with AI
              </p>
            </div>
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Refresh Repositories
            </Button>
          </div>

          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {repositories.map((repository) => (
              <RepositoryCard
                key={repository.id}
                repository={repository}
                onConnect={handleConnectRepository}
                onOpenWorkspace={handleOpenWorkspace}
              />
            ))}
          </div>
        </div>
      </main>
      
      {/* Workspace Panel */}
      {selectedRepository && (
        <WorkspacePanel
          repository={selectedRepository}
          onClose={handleCloseWorkspace}
        />
      )}
    </div>
  );
}
