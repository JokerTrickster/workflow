'use client';

import { useState } from 'react';
import { Repository } from '../../../domain/entities/Repository';
import { Task } from '../../../domain/entities/Task';
import { TaskCreationForm } from '../../../components/TaskCreationForm';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../../components/ui/tabs';
import { Button } from '../../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card';
import { Badge } from '../../../components/ui/badge';
import { Input } from '../../../components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../../components/ui/select';
import { 
  Plus, 
  Play, 
  ExternalLink, 
  Calendar, 
  GitBranch, 
  Clock, 
  CheckCircle, 
  XCircle, 
  Search,
  GitPullRequest,
  AlertCircle,
  Loader2,
  RefreshCw,
  Users
} from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../../../components/ui/dialog';
import { useGitHubIssues } from '../../../hooks/useGitHubIssues';
import { useGitHubPullRequests } from '../../../hooks/useGitHubPullRequests';
import { GitHubIssue, GitHubPullRequest } from '../../../types/github';

interface TaskTabProps {
  repository: Repository;
}

// Mock tasks data (keeping the existing task functionality)
const mockTasks: Task[] = [
  {
    id: '1',
    title: 'Add dark mode support',
    description: 'Implement dark mode toggle in settings page with theme persistence',
    status: 'completed',
    repository_id: 1,
    created_at: '2024-09-05T10:00:00Z',
    updated_at: '2024-09-05T15:30:00Z',
    completed_at: '2024-09-05T15:30:00Z',
    branch_name: 'feature/dark-mode',
    pr_url: 'https://github.com/captain/ai-git-workbench/pull/1',
    build_status: 'success',
    lint_status: 'success',
    ai_tokens_used: 1250,
  },
  {
    id: '2',
    title: 'Fix authentication bug',
    description: 'Resolve token refresh issue causing unexpected logouts',
    status: 'in_progress',
    repository_id: 1,
    created_at: '2024-09-06T08:00:00Z',
    updated_at: '2024-09-06T10:15:00Z',
    started_at: '2024-09-06T09:00:00Z',
    branch_name: 'fix/auth-token-refresh',
    build_status: 'pending',
    ai_tokens_used: 800,
  },
  {
    id: '3',
    title: 'Add unit tests',
    description: 'Write comprehensive unit tests for core authentication and repository management functions',
    status: 'pending',
    repository_id: 1,
    created_at: '2024-09-06T09:00:00Z',
    updated_at: '2024-09-06T09:00:00Z',
  },
];

export function TaskTab({ repository }: TaskTabProps) {
  const [tasks, setTasks] = useState<Task[]>(mockTasks);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [isCreatingTask, setIsCreatingTask] = useState(false);
  const [selectedGitHubIssue, setSelectedGitHubIssue] = useState<GitHubIssue | undefined>();
  const [selectedGitHubPR, setSelectedGitHubPR] = useState<GitHubPullRequest | undefined>();
  
  // GitHub Issues and PRs state
  const [issuesFilter, setIssuesFilter] = useState<'all' | 'open' | 'closed'>('open');
  const [prsFilter, setPrsFilter] = useState<'all' | 'open' | 'closed'>('open');
  const [searchTerm, setSearchTerm] = useState('');
  const [activeSubTab, setActiveSubTab] = useState('tasks');

  // GitHub API calls
  const repoId = repository.full_name;
  
  const {
    data: issuesData,
    isLoading: issuesLoading,
    isError: issuesError,
    error: issuesErrorData,
    refetch: refetchIssues
  } = useGitHubIssues({
    repoId,
    params: { state: issuesFilter, per_page: 20 },
    enabled: repository.is_connected && activeSubTab === 'issues'
  });

  const {
    data: prsData,
    isLoading: prsLoading,
    isError: prsError,
    error: prsErrorData,
    refetch: refetchPRs
  } = useGitHubPullRequests({
    repoId,
    params: { state: prsFilter, per_page: 20 },
    enabled: repository.is_connected && activeSubTab === 'prs'
  });

  // Task functions (enhanced with domain layer integration)
  const handleCreateTask = async (taskData: Omit<Task, 'id' | 'created_at' | 'updated_at'>) => {
    setIsCreatingTask(true);
    try {
      // For now, simulate the domain layer until we have a proper repository implementation
      const newTask: Task = {
        id: Date.now().toString(),
        ...taskData,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      setTasks(prev => [newTask, ...prev]);
      setShowCreateDialog(false);
      setSelectedGitHubIssue(undefined);
      setSelectedGitHubPR(undefined);
      setActiveSubTab('tasks'); // Switch to tasks tab to show the new task
    } catch (error) {
      console.error('Failed to create task:', error);
      // TODO: Add proper error handling/notification
    } finally {
      setIsCreatingTask(false);
    }
  };

  const handleExecuteTask = (taskId: string) => {
    setTasks(prev =>
      prev.map(task =>
        task.id === taskId 
          ? { ...task, status: 'in_progress', started_at: new Date().toISOString() }
          : task
      )
    );
  };

  const handleCreateTaskFromIssue = (issue: GitHubIssue) => {
    setSelectedGitHubIssue(issue);
    setSelectedGitHubPR(undefined);
    setShowCreateDialog(true);
  };

  const handleCreateTaskFromPR = (pr: GitHubPullRequest) => {
    setSelectedGitHubPR(pr);
    setSelectedGitHubIssue(undefined);
    setShowCreateDialog(true);
  };

  const handleCancelTaskCreation = () => {
    setShowCreateDialog(false);
    setSelectedGitHubIssue(undefined);
    setSelectedGitHubPR(undefined);
  };

  // Helper functions
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'in_progress':
        return <Clock className="h-4 w-4 text-yellow-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-400" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'in_progress':
        return 'bg-yellow-100 text-yellow-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getIssueStateBadge = (state: string) => {
    return state === 'open' ? (
      <Badge className="bg-green-100 text-green-800">
        <AlertCircle className="h-3 w-3 mr-1" />
        Open
      </Badge>
    ) : (
      <Badge className="bg-purple-100 text-purple-800">
        <CheckCircle className="h-3 w-3 mr-1" />
        Closed
      </Badge>
    );
  };

  const getPRStateBadge = (state: string, merged: boolean) => {
    if (merged) {
      return (
        <Badge className="bg-purple-100 text-purple-800">
          <GitPullRequest className="h-3 w-3 mr-1" />
          Merged
        </Badge>
      );
    }
    
    return state === 'open' ? (
      <Badge className="bg-green-100 text-green-800">
        <GitPullRequest className="h-3 w-3 mr-1" />
        Open
      </Badge>
    ) : (
      <Badge className="bg-red-100 text-red-800">
        <GitPullRequest className="h-3 w-3 mr-1" />
        Closed
      </Badge>
    );
  };

  // Filter functions
  const filteredIssues = issuesData?.issues?.filter(issue =>
    searchTerm === '' || 
    issue.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    issue.body?.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  const filteredPRs = prsData?.pullRequests?.filter(pr =>
    searchTerm === '' || 
    pr.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    pr.body?.toLowerCase().includes(searchTerm.toLowerCase())
  ) || [];

  // Render functions for each tab
  const renderTasks = () => (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h3 className="text-lg font-semibold">Local Tasks</h3>
        <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-2" />
              New Task
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[600px]">
            <DialogHeader>
              <DialogTitle>
                {selectedGitHubIssue || selectedGitHubPR ? 'Create Task from GitHub' : 'Create New Task'}
              </DialogTitle>
              <DialogDescription>
                {selectedGitHubIssue || selectedGitHubPR 
                  ? 'Create a local task based on the selected GitHub item.'
                  : 'Create a new task for the AI agent to work on.'
                }
              </DialogDescription>
            </DialogHeader>
            <div className="pt-4">
              <TaskCreationForm
                repositoryId={repository.id}
                onSubmit={handleCreateTask}
                onCancel={handleCancelTaskCreation}
                githubIssue={selectedGitHubIssue}
                githubPullRequest={selectedGitHubPR}
                isSubmitting={isCreatingTask}
              />
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <div className="space-y-4">
        {tasks.map((task) => (
          <Card key={task.id}>
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between gap-4">
                <div className="space-y-1 flex-1">
                  <CardTitle className="text-base flex items-center gap-2">
                    {getStatusIcon(task.status)}
                    {task.title}
                  </CardTitle>
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge className={getStatusColor(task.status)}>
                      {task.status.replace('_', ' ')}
                    </Badge>
                    {task.branch_name && (
                      <Badge variant="outline" className="text-xs">
                        <GitBranch className="h-3 w-3 mr-1" />
                        {task.branch_name}
                      </Badge>
                    )}
                  </div>
                </div>
                <div className="flex gap-2 shrink-0">
                  {task.pr_url && (
                    <Button variant="outline" size="sm" asChild>
                      <a href={task.pr_url} target="_blank" rel="noopener noreferrer">
                        <ExternalLink className="h-4 w-4" />
                      </a>
                    </Button>
                  )}
                  {task.status === 'pending' && (
                    <Button size="sm" onClick={() => handleExecuteTask(task.id)}>
                      <Play className="h-4 w-4 mr-2" />
                      Execute
                    </Button>
                  )}
                </div>
              </div>
            </CardHeader>
            {task.description && (
              <CardContent className="pt-0">
                <p className="text-sm text-muted-foreground mb-4">{task.description}</p>
                <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Calendar className="h-3 w-3" />
                    {new Date(task.created_at).toLocaleDateString()}
                  </div>
                  {task.ai_tokens_used && (
                    <div>Tokens: {task.ai_tokens_used.toLocaleString()}</div>
                  )}
                  {task.build_status && (
                    <Badge variant={task.build_status === 'success' ? 'default' : 'secondary'} className="text-xs">
                      Build: {task.build_status}
                    </Badge>
                  )}
                </div>
              </CardContent>
            )}
          </Card>
        ))}
      </div>
    </div>
  );

  const renderGitHubContent = (
    type: 'issues' | 'prs',
    items: GitHubIssue[] | GitHubPullRequest[],
    isLoading: boolean,
    isError: boolean,
    error: Error | null,
    refetch: () => void,
    filter: 'all' | 'open' | 'closed',
    setFilter: (value: 'all' | 'open' | 'closed') => void
  ) => {
    if (!repository.is_connected) {
      return (
        <div className="text-center py-8">
          <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-semibold mb-2">Repository Not Connected</h3>
          <p className="text-sm text-muted-foreground">
            Connect this repository to view {type === 'issues' ? 'issues' : 'pull requests'}
          </p>
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {/* Header with controls */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <h3 className="text-lg font-semibold">
            GitHub {type === 'issues' ? 'Issues' : 'Pull Requests'}
          </h3>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={refetch} disabled={isLoading}>
              <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </div>

        {/* Filters */}
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="h-4 w-4 absolute left-3 top-3 text-muted-foreground" />
            <Input
              placeholder={`Search ${type}...`}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          <Select value={filter} onValueChange={setFilter}>
            <SelectTrigger className="w-full sm:w-[180px]">
              <SelectValue placeholder="Filter by state" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="open">Open</SelectItem>
              <SelectItem value="closed">{type === 'prs' ? 'Closed/Merged' : 'Closed'}</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Content */}
        {isLoading && (
          <div className="text-center py-8">
            <Loader2 className="h-8 w-8 mx-auto mb-4 animate-spin" />
            <p className="text-sm text-muted-foreground">Loading {type}...</p>
          </div>
        )}

        {isError && (
          <div className="text-center py-8">
            <AlertCircle className="h-12 w-12 mx-auto mb-4 text-red-500" />
            <h3 className="text-lg font-semibold mb-2">Failed to load {type}</h3>
            <p className="text-sm text-muted-foreground mb-4">
              {error?.message || 'An error occurred while fetching data'}
            </p>
            <Button variant="outline" onClick={refetch}>
              Try Again
            </Button>
          </div>
        )}

        {!isLoading && !isError && items.length === 0 && (
          <div className="text-center py-8">
            <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <h3 className="text-lg font-semibold mb-2">
              No {type === 'issues' ? 'issues' : 'pull requests'} found
            </h3>
            <p className="text-sm text-muted-foreground">
              {searchTerm 
                ? `No ${type} match your search criteria` 
                : `This repository has no ${filter} ${type}`}
            </p>
          </div>
        )}

        {!isLoading && !isError && items.length > 0 && (
          <div className="space-y-4">
            {items.map((item) => {
              const isIssue = 'pull_request' in item ? false : true;
              const issue = isIssue ? (item as GitHubIssue) : null;
              const pr = !isIssue ? (item as GitHubPullRequest) : null;

              return (
                <Card key={item.id}>
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between gap-4">
                      <div className="space-y-2 flex-1">
                        <CardTitle className="text-base">
                          <a 
                            href={item.html_url} 
                            target="_blank" 
                            rel="noopener noreferrer"
                            className="hover:underline flex items-center gap-2"
                          >
                            #{item.number}: {item.title}
                            <ExternalLink className="h-3 w-3 opacity-50" />
                          </a>
                        </CardTitle>
                        <div className="flex flex-wrap items-center gap-2">
                          {issue && getIssueStateBadge(issue.state)}
                          {pr && getPRStateBadge(pr.state, pr.merged)}
                          {pr && (
                            <Badge variant="outline" className="text-xs">
                              <GitBranch className="h-3 w-3 mr-1" />
                              {pr.head.ref} → {pr.base.ref}
                            </Badge>
                          )}
                          {item.assignees.length > 0 && (
                            <Badge variant="outline" className="text-xs">
                              <Users className="h-3 w-3 mr-1" />
                              {item.assignees.length} assigned
                            </Badge>
                          )}
                          {item.labels.map(label => (
                            <Badge 
                              key={label.id} 
                              variant="outline" 
                              className="text-xs"
                              style={{ 
                                backgroundColor: `#${label.color}20`,
                                borderColor: `#${label.color}`,
                                color: `#${label.color}`
                              }}
                            >
                              {label.name}
                            </Badge>
                          ))}
                        </div>
                      </div>
                      <div className="flex gap-2 shrink-0">
                        <Button 
                          variant="outline" 
                          size="sm"
                          onClick={() => {
                            if (issue) handleCreateTaskFromIssue(issue);
                            if (pr) handleCreateTaskFromPR(pr);
                          }}
                        >
                          <Plus className="h-4 w-4 mr-2" />
                          Create Task
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  {item.body && (
                    <CardContent className="pt-0">
                      <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                        {item.body}
                      </p>
                      <div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <Calendar className="h-3 w-3" />
                          {new Date(item.created_at).toLocaleDateString()}
                        </div>
                        <div>
                          by {item.user.login}
                        </div>
                        <div>
                          {item.comments} comments
                        </div>
                      </div>
                    </CardContent>
                  )}
                </Card>
              );
            })}
          </div>
        )}
      </div>
    );
  };

  return (
    <Tabs value={activeSubTab} onValueChange={setActiveSubTab} className="w-full">
      <TabsList className="grid w-full grid-cols-3 mb-6">
        <TabsTrigger value="tasks" className="flex items-center gap-2">
          <CheckCircle className="h-4 w-4" />
          <span className="hidden sm:inline">Tasks</span>
        </TabsTrigger>
        <TabsTrigger value="issues" className="flex items-center gap-2">
          <AlertCircle className="h-4 w-4" />
          <span className="hidden sm:inline">Issues</span>
        </TabsTrigger>
        <TabsTrigger value="prs" className="flex items-center gap-2">
          <GitPullRequest className="h-4 w-4" />
          <span className="hidden sm:inline">PRs</span>
        </TabsTrigger>
      </TabsList>

      <TabsContent value="tasks" className="space-y-4">
        {renderTasks()}
      </TabsContent>

      <TabsContent value="issues" className="space-y-4">
        {renderGitHubContent(
          'issues',
          filteredIssues,
          issuesLoading,
          issuesError,
          issuesErrorData as Error,
          refetchIssues,
          issuesFilter,
          setIssuesFilter
        )}
      </TabsContent>

      <TabsContent value="prs" className="space-y-4">
        {renderGitHubContent(
          'prs',
          filteredPRs,
          prsLoading,
          prsError,
          prsErrorData as Error,
          refetchPRs,
          prsFilter,
          setPrsFilter
        )}
      </TabsContent>
    </Tabs>
  );
}