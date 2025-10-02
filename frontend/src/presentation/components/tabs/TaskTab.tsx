'use client';

import { useState, useEffect, useCallback } from 'react';
import { Repository } from '../../../domain/entities/Repository';
import { Task } from '../../../domain/entities/Task';
import { TaskCreationForm } from '../../../components/TaskCreationForm';
import { TaskFileManager } from '../../../services/TaskFileManager';
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
  Tag,
  MessageCircle,
  GitMerge,
  Eye,
  Link,
  Filter,
  User,
  UserCheck,
  GitCommit,
  ChevronDown,
  X
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
import { apiConfig } from '../../../config/api';
import { GitHubIssue, GitHubPullRequest } from '../../../types/github';

interface TaskTabProps {
  repository: Repository;
  activeTab: string;
}


export function TaskTab({ repository, activeTab }: TaskTabProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoadingTasks, setIsLoadingTasks] = useState(true);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [isCreatingTask, setIsCreatingTask] = useState(false);
  const [selectedGitHubIssue, setSelectedGitHubIssue] = useState<GitHubIssue | undefined>();
  const [selectedGitHubPR, setSelectedGitHubPR] = useState<GitHubPullRequest | undefined>();
  const [selectedTask, setSelectedTask] = useState<Task | undefined>();
  const [showTaskDetailDialog, setShowTaskDetailDialog] = useState(false);
  const [showExecuteDialog, setShowExecuteDialog] = useState(false);
  const [taskToExecute, setTaskToExecute] = useState<Task | undefined>();
  
  const taskFileManager = TaskFileManager.getInstance();

  // Load tasks from epic files on component mount and repository change
  useEffect(() => {
    // Clear any local storage cache if it exists
    try {
      localStorage.removeItem('tasks-cache');
      sessionStorage.removeItem('tasks-cache');
      // Clear all potential cache keys
      Object.keys(localStorage).forEach(key => {
        if (key.includes('tasks') || key.includes('epic') || key.includes('workflow')) {
          localStorage.removeItem(key);
        }
      });    
    } catch {
      // Ignore storage errors
    }
    loadTasks();
  }, [repository.id, repository.name]); // Reload when repository changes

  const loadTasks = useCallback(async () => {
    setIsLoadingTasks(true);
    try {
      // Load tasks from backend API
      const apiResponse = await fetch(apiConfig.endpoints.tasks.list(repository.name));

      if (apiResponse.ok) {
        const apiResult = await apiResponse.json();
        console.log('✅ Tasks loaded from backend:', apiResult.tasks?.length || 0);

        // Convert backend task format to frontend task format
        const convertedTasks: Task[] = (apiResult.tasks || []).map((backendTask: any) => {
          // Parse title and description from tasks field
          const tasksContent = backendTask.tasks || '';
          let title = 'Untitled Task';
          let description = tasksContent;

          // If tasks starts with markdown heading (# Title), extract it
          if (tasksContent.startsWith('# ')) {
            const lines = tasksContent.split('\n');
            title = lines[0].replace(/^#\s*/, '').trim();
            // Description is everything after the first line
            description = lines.slice(1).join('\n').trim();
          } else {
            // Use first line as title
            const firstLine = tasksContent.split('\n')[0]?.trim();
            if (firstLine) {
              title = firstLine.length > 100 ? firstLine.substring(0, 100) + '...' : firstLine;
            }
          }

          return {
            id: backendTask.request_id,
            title,
            description,
            status: backendTask.status,
            repository: backendTask.repository_name,
            epic: 'backend-tasks',
            branch: backendTask.working_dir || undefined,
            created_at: backendTask.created_at,
            updated_at: backendTask.updated_at
          };
        });

        setTasks(convertedTasks);
        console.log(`Loaded ${convertedTasks.length} tasks from backend for repository: ${repository.name}`);
      } else {
        console.warn('⚠️ Failed to load tasks from backend');
        setTasks([]); // Set empty tasks instead of fallback
      }
    } catch (error) {
      console.error('Failed to load tasks from backend:', error);
      setTasks([]); // Set empty tasks on error
    } finally {
      setIsLoadingTasks(false);
    }
  }, [repository.name]);

  // Sub-tab state needs to be declared before useEffect
  const [activeSubTab, setActiveSubTab] = useState('tasks');

  // Load tasks when tab becomes active (on tab change)
  useEffect(() => {
    if (activeTab === 'tasks' || activeSubTab === 'tasks') {
      console.log('🔄 Tab switched to tasks, refreshing task list');
      loadTasks();
    }
  }, [activeTab, activeSubTab, loadTasks]);

  // Handle task detail viewing
  const handleTaskClick = async (task: Task) => {
    try {
      // Fetch full task details from backend API
      const response = await fetch(apiConfig.endpoints.tasks.list(repository.name));

      if (response.ok) {
        const result = await response.json();
        const fullTask = result.tasks?.find((t: any) => t.request_id === task.id);

        if (fullTask) {
          // Parse full tasks content for modal display
          const tasksContent = fullTask.tasks || '';
          let modalDescription = tasksContent;

          // If tasks starts with markdown heading, remove it for display
          if (tasksContent.startsWith('# ')) {
            const lines = tasksContent.split('\n');
            // Skip first line (title) and show rest as description
            modalDescription = lines.slice(1).join('\n').trim();
          }

          setSelectedTask({
            ...task,
            description: modalDescription || task.description
          });
        } else {
          setSelectedTask(task);
        }
      } else {
        setSelectedTask(task);
      }

      setShowTaskDetailDialog(true);
    } catch (error) {
      console.error('Failed to load task details:', error);
      // Still show basic task info if loading fails
      setSelectedTask(task);
      setShowTaskDetailDialog(true);
    }
  };

  const handleCloseTaskDetail = () => {
    setSelectedTask(undefined);
    setShowTaskDetailDialog(false);
  };
  
  // GitHub Issues and PRs state
  const [issuesFilter, setIssuesFilter] = useState<'all' | 'open' | 'closed'>('open');
  const [prsFilter, setPrsFilter] = useState<'all' | 'open' | 'closed'>('open');
  const [searchTerm, setSearchTerm] = useState('');
  
  // Enhanced filtering state
  const [labelFilter, setLabelFilter] = useState<string>('');
  const [assigneeFilter, setAssigneeFilter] = useState<string>('');
  const [linkedTasksFilter, setLinkedTasksFilter] = useState<'all' | 'linked' | 'unlinked'>('all');
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);

  // GitHub API setup
  const repoId = repository.full_name;
  const shouldEnableGitHub = repository.is_connected && !!repository.full_name;

  const {
    data: issuesData,
    isLoading: issuesLoading,
    isError: issuesError,
    error: issuesErrorData,
    refetch: refetchIssues
  } = useGitHubIssues({
    repoId,
    params: { state: issuesFilter, per_page: 20 },
    enabled: shouldEnableGitHub // 항상 활성화하여 탭 변경 시 바로 로드
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
    enabled: shouldEnableGitHub // 항상 활성화하여 탭 변경 시 바로 로드
  });

  // Auto-refetch GitHub data when switching tabs - force refresh every time
  useEffect(() => {
    if (shouldEnableGitHub) {
      try {
        // Add a small delay to ensure hooks are ready
        const timeoutId = setTimeout(() => {
          if (activeTab === 'issues') {
            console.log('🔄 Auto-refetching GitHub issues on tab switch');
            refetchIssues();
          } else if (activeTab === 'prs') {
            console.log('🔄 Auto-refetching GitHub PRs on tab switch');
            refetchPRs();
          }
        }, 100); // 100ms delay

        return () => clearTimeout(timeoutId);
      } catch (error) {
        console.error('Error refetching GitHub data:', error);
      }
    }
  }, [activeTab, shouldEnableGitHub, refetchIssues, refetchPRs]);

  // Task functions (enhanced with domain layer integration)
  const handleCreateTask = async (taskData: Omit<Task, 'id' | 'created_at' | 'updated_at'>) => {
    // Prevent duplicate requests
    if (isCreatingTask) return;

    setIsCreatingTask(true);
    try {
      // Determine epic name from GitHub issue/PR or use default
      let epicName = 'general-tasks';
      if (selectedGitHubIssue) {
        epicName = `github-issue-${selectedGitHubIssue.number}`;
      } else if (selectedGitHubPR) {
        epicName = `github-pr-${selectedGitHubPR.number}`;
      }

      // Create task content
      const taskContent = `# ${taskData.title}

## Description
${taskData.description}

## Repository
${repository.name} (${repository.full_name})

${selectedGitHubIssue ? `## GitHub Issue
- Issue #${selectedGitHubIssue.number}: ${selectedGitHubIssue.title}
- URL: ${selectedGitHubIssue.html_url}
` : ''}${selectedGitHubPR ? `## GitHub Pull Request
- PR #${selectedGitHubPR.number}: ${selectedGitHubPR.title}
- URL: ${selectedGitHubPR.html_url}
` : ''}
## Progress
- [ ] Task created
- [ ] Analysis completed
- [ ] Implementation started
- [ ] Testing completed
- [ ] Task completed

## Notes
Task created on ${new Date().toISOString()}`;

      // Create task via backend API (correct format)
      // Combine title and description in tasks field with markdown format
      const fullTaskDescription = `# ${taskData.title}\n\n${taskData.description}`;

      const taskCreateRequest = {
        tasks: fullTaskDescription, // Title + description in markdown format
        repository_name: repository.name,
        working_dir: `JokerTrickster/${repository.name}`, // Repository path for actual work
        branch_name: taskData.branch_name || '', // User input for branch name
        provider: 'claude',
        cmd: '',
        interactive: false
      };

      const response = await fetch(apiConfig.endpoints.tasks.create(), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(taskCreateRequest)
      });

      if (!response.ok) {
        const errorData = await response.text();
        console.error('Task creation API failed:', errorData);
        throw new Error(`Task creation API failed: ${errorData}`);
      }

      const result = await response.json();
      console.log('Task created:', result);

      // Reload tasks to show the new one
      await loadTasks();
      
      // Log task creation
      const metadata: { githubUrl?: string; issueNumber?: number; prNumber?: number; branchName?: string } = {};
      if (selectedGitHubIssue) {
        metadata.githubUrl = selectedGitHubIssue.html_url;
        metadata.issueNumber = selectedGitHubIssue.number;
      }
      if (selectedGitHubPR) {
        metadata.githubUrl = selectedGitHubPR.html_url;
        metadata.prNumber = selectedGitHubPR.number;
      }
      if (taskData.branch_name) {
        metadata.branchName = taskData.branch_name;
      }
      
      setShowCreateDialog(false);
      setSelectedGitHubIssue(undefined);
      setSelectedGitHubPR(undefined);
      setActiveSubTab('tasks'); // Switch to tasks tab to show the new task

      // Auto-refresh tasks list
      await loadTasks();
    } catch (error) {
      console.error('Failed to create task:', error);
      
      // TODO: Add proper error handling/notification
    } finally {
      setIsCreatingTask(false);
    }
  };

  const handleConfirmExecuteTask = async () => {
    if (!taskToExecute) return;

    setShowExecuteDialog(false);

    try {
      console.log('Executing task:', taskToExecute.id);

      // Use direct API call with explicit POST method
      const response = await fetch(apiConfig.endpoints.tasks.execute(taskToExecute.id), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      console.log('Execute response status:', response.status);

      if (!response.ok) {
        const errorData = await response.text();
        console.error('Execute API failed:', errorData);
        throw new Error(`Execute API failed: ${errorData}`);
      }

      const result = await response.json();
      console.log('Task execution result:', result);

      // Auto-refresh tasks list
      await loadTasks();
    } catch (error) {
      console.error('Failed to execute task:', error);
    } finally {
      setTaskToExecute(undefined);
    }
  };

  const handleCancelExecuteTask = () => {
    setShowExecuteDialog(false);
    setTaskToExecute(undefined);
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
      case 'processing':
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
      case 'processing':
        return 'bg-yellow-100 text-yellow-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getIssueStateBadge = (state: string, stateReason?: string) => {
    if (state === 'open') {
      return (
        <Badge className="bg-green-100 text-green-800 border-green-300">
          <AlertCircle className="h-3 w-3 mr-1" />
          Open
        </Badge>
      );
    }
    
    const isCompleted = stateReason === 'completed';
    return (
      <Badge className={isCompleted ? "bg-purple-100 text-purple-800 border-purple-300" : "bg-gray-100 text-gray-800 border-gray-300"}>
        <CheckCircle className="h-3 w-3 mr-1" />
        {isCompleted ? 'Completed' : 'Closed'}
      </Badge>
    );
  };

  const getPRStateBadge = (state: string, merged: boolean, draft?: boolean) => {
    if (merged) {
      return (
        <Badge className="bg-purple-100 text-purple-800 border-purple-300">
          <GitMerge className="h-3 w-3 mr-1" />
          Merged
        </Badge>
      );
    }
    
    if (draft && state === 'open') {
      return (
        <Badge className="bg-gray-100 text-gray-700 border-gray-300">
          <GitCommit className="h-3 w-3 mr-1" />
          Draft
        </Badge>
      );
    }
    
    return state === 'open' ? (
      <Badge className="bg-green-100 text-green-800 border-green-300">
        <GitPullRequest className="h-3 w-3 mr-1" />
        Open
      </Badge>
    ) : (
      <Badge className="bg-red-100 text-red-800 border-red-300">
        <GitPullRequest className="h-3 w-3 mr-1" />
        Closed
      </Badge>
    );
  };

  // Helper function to check if GitHub item has linked tasks
  const getLinkedTaskCount = (item: GitHubIssue | GitHubPullRequest): number => {
    // For now, check if any local task references this GitHub item
    // In the future, this could be enhanced with a proper database lookup
    return tasks.filter(task => {
      if ('pull_request' in item) {
        // This is actually an issue, but TypeScript guard isn't working properly
        return false;
      }
      const pr = item as GitHubPullRequest;
      const issue = item as GitHubIssue;
      
      // Check if task was created from this GitHub item
      return task.pr_url === (pr.html_url || issue.html_url) ||
        task.description?.includes((pr.html_url || issue.html_url) || '') ||
        task.title.includes(`#${item.number}`);
    }).length;
  };

  // Get unique labels from current data for filter dropdown
  const getAvailableLabels = (type: 'issues' | 'prs') => {
    const items = type === 'issues' ? (issuesData?.issues || []) : (prsData?.pullRequests || []);
    const allLabels = items.flatMap(item => item.labels);
    const uniqueLabels = Array.from(new Set(allLabels.map(label => label.name)))
      .map(name => allLabels.find(label => label.name === name)!)
      .sort((a, b) => a.name.localeCompare(b.name));
    return uniqueLabels;
  };

  // Get unique assignees/reviewers from current data
  const getAvailableAssignees = (type: 'issues' | 'prs') => {
    const items = type === 'issues' ? (issuesData?.issues || []) : (prsData?.pullRequests || []);
    const allUsers = items.flatMap(item => {
      const users = [...item.assignees];
      if (item.assignee) users.push(item.assignee);
      if ('requested_reviewers' in item) {
        users.push(...(item as GitHubPullRequest).requested_reviewers);
      }
      return users;
    });
    const uniqueUsers = Array.from(new Set(allUsers.map(user => user.login)))
      .map(login => allUsers.find(user => user.login === login)!)
      .sort((a, b) => a.login.localeCompare(b.login));
    return uniqueUsers;
  };

  // Reset filters
  const resetFilters = () => {
    setSearchTerm('');
    setLabelFilter('');
    setAssigneeFilter('');
    setLinkedTasksFilter('all');
  };

  // Enhanced filter functions
  const filteredIssues = issuesData?.issues?.filter(issue => {
    // Search filter
    const matchesSearch = searchTerm === '' || 
      issue.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      issue.body?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      issue.user.login.toLowerCase().includes(searchTerm.toLowerCase());
    
    // Label filter
    const matchesLabel = labelFilter === '' || 
      issue.labels.some(label => label.name.toLowerCase().includes(labelFilter.toLowerCase()));
    
    // Assignee filter
    const matchesAssignee = assigneeFilter === '' ||
      issue.assignees.some(assignee => assignee.login.toLowerCase().includes(assigneeFilter.toLowerCase())) ||
      (issue.assignee && issue.assignee.login.toLowerCase().includes(assigneeFilter.toLowerCase()));
    
    // Task linkage filter (for future implementation - currently shows all)
    const matchesLinkage = linkedTasksFilter === 'all' ||
      (linkedTasksFilter === 'linked' && getLinkedTaskCount(issue) > 0) ||
      (linkedTasksFilter === 'unlinked' && getLinkedTaskCount(issue) === 0);
    
    return matchesSearch && matchesLabel && matchesAssignee && matchesLinkage;
  }) || [];

  const filteredPRs = prsData?.pullRequests?.filter(pr => {
    // Search filter
    const matchesSearch = searchTerm === '' || 
      pr.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pr.body?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pr.user.login.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pr.head.ref.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pr.base.ref.toLowerCase().includes(searchTerm.toLowerCase());
    
    // Label filter
    const matchesLabel = labelFilter === '' || 
      pr.labels.some(label => label.name.toLowerCase().includes(labelFilter.toLowerCase()));
    
    // Assignee/Reviewer filter
    const matchesAssignee = assigneeFilter === '' ||
      pr.assignees.some(assignee => assignee.login.toLowerCase().includes(assigneeFilter.toLowerCase())) ||
      pr.requested_reviewers.some(reviewer => reviewer.login.toLowerCase().includes(assigneeFilter.toLowerCase())) ||
      (pr.assignee && pr.assignee.login.toLowerCase().includes(assigneeFilter.toLowerCase()));
    
    // Task linkage filter (for future implementation - currently shows all)
    const matchesLinkage = linkedTasksFilter === 'all' ||
      (linkedTasksFilter === 'linked' && getLinkedTaskCount(pr) > 0) ||
      (linkedTasksFilter === 'unlinked' && getLinkedTaskCount(pr) === 0);
    
    return matchesSearch && matchesLabel && matchesAssignee && matchesLinkage;
  }) || [];

  // Render functions for each tab
  const renderTasks = () => (
    <div className="space-y-4 h-full flex flex-col">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h3 className="text-lg font-semibold">Local Tasks</h3>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={loadTasks} disabled={isLoadingTasks}>
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoadingTasks ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
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
      </div>

      {isLoadingTasks ? (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                <div className="h-3 bg-gray-200 rounded w-1/2 mt-2"></div>
              </CardHeader>
            </Card>
          ))}
        </div>
      ) : tasks.length === 0 ? (
        <div className="text-center py-8">
          <Clock className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-semibold mb-2">No Tasks Found</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Create your first task to get started with AI-powered development
          </p>
        </div>
      ) : (
        <div className="space-y-4 flex-1 overflow-y-auto">
          {tasks.map((task) => (
          <Card 
            key={task.id} 
            className="cursor-pointer hover:shadow-md transition-shadow"
            onClick={() => handleTaskClick(task)}
          >
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
                  {task.github_issue_url && (
                    <Button variant="outline" size="sm" asChild title="View GitHub Issue">
                      <a href={task.github_issue_url} target="_blank" rel="noopener noreferrer">
                        <MessageCircle className="h-4 w-4" />
                      </a>
                    </Button>
                  )}
                  {task.github_pr_url && (
                    <Button variant="outline" size="sm" asChild title="View Pull Request">
                      <a href={task.github_pr_url} target="_blank" rel="noopener noreferrer">
                        <GitPullRequest className="h-4 w-4" />
                      </a>
                    </Button>
                  )}
                  {task.pr_url && (
                    <Button variant="outline" size="sm" asChild title="Legacy PR Link">
                      <a href={task.pr_url} target="_blank" rel="noopener noreferrer">
                        <ExternalLink className="h-4 w-4" />
                      </a>
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
                    {new Date(task.created_at).toLocaleString()}
                  </div>
                  {task.ai_tokens_used && (
                    <div>Tokens: {task.ai_tokens_used.toLocaleString()}</div>
                  )}
                  {task.build_status && (
                    <Badge variant={task.build_status === 'success' ? 'default' : 'secondary'} className="text-xs">
                      Build: {task.build_status}
                    </Badge>
                  )}
                  {task.provider && (
                    <div className="flex items-center gap-1">
                      <User className="h-3 w-3" />
                      {task.provider}
                    </div>
                  )}
                  {task.cleanup_status && (
                    <Badge variant="outline" className="text-xs">
                      Cleanup: {task.cleanup_status}
                    </Badge>
                  )}
                </div>

                {/* GitHub Integration Status */}
                {(task.github_issue_url || task.github_pr_url || task.branch_name) && (
                  <div className="mt-3 pt-3 border-t border-gray-100">
                    <div className="flex flex-wrap items-center gap-3 text-xs">
                      {task.github_issue_url && (
                        <a
                          href={task.github_issue_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1 text-blue-600 hover:text-blue-800"
                        >
                          <MessageCircle className="h-3 w-3" />
                          Issue
                        </a>
                      )}
                      {task.github_pr_url && (
                        <a
                          href={task.github_pr_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1 text-green-600 hover:text-green-800"
                        >
                          <GitPullRequest className="h-3 w-3" />
                          Pull Request
                        </a>
                      )}
                      {task.branch_name && (
                        <div className="flex items-center gap-1 text-gray-600">
                          <GitBranch className="h-3 w-3" />
                          {task.branch_name}
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </CardContent>
            )}
          </Card>
          ))}
        </div>
      )}
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
        <div className="space-y-4">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1">
              <Search className="h-4 w-4 absolute left-3 top-3 text-muted-foreground" />
              <Input
                placeholder={`Search ${type === 'issues' ? 'issues, authors, content' : 'PRs, branches, authors, content'}...`}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={filter} onValueChange={setFilter}>
              <SelectTrigger className="w-full sm:w-[140px]">
                <SelectValue placeholder="State" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All States</SelectItem>
                <SelectItem value="open">Open</SelectItem>
                <SelectItem value="closed">{type === 'prs' ? 'Closed/Merged' : 'Closed'}</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
              className="w-full sm:w-auto"
            >
              <Filter className="h-4 w-4 mr-2" />
              Filters
              <ChevronDown className={`h-4 w-4 ml-2 transition-transform ${showAdvancedFilters ? 'rotate-180' : ''}`} />
            </Button>
          </div>

          {/* Advanced Filters */}
          {showAdvancedFilters && (
            <div className="p-4 border rounded-lg bg-muted/30 space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {/* Label Filter */}
                <div className="space-y-2">
                  <label className="text-sm font-medium flex items-center gap-2">
                    <Tag className="h-3 w-3" />
                    Label
                  </label>
                  <Select value={labelFilter} onValueChange={setLabelFilter}>
                    <SelectTrigger>
                      <SelectValue placeholder="Any label" />
                    </SelectTrigger>
                    <SelectContent className="max-h-48">
                      <SelectItem value="">Any label</SelectItem>
                      {getAvailableLabels(type).map(label => (
                        <SelectItem key={label.id} value={label.name}>
                          <div className="flex items-center gap-2">
                            <div 
                              className="w-3 h-3 rounded-sm" 
                              style={{ backgroundColor: `#${label.color}` }}
                            />
                            {label.name}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Assignee/Reviewer Filter */}
                <div className="space-y-2">
                  <label className="text-sm font-medium flex items-center gap-2">
                    <User className="h-3 w-3" />
                    {type === 'prs' ? 'Assignee/Reviewer' : 'Assignee'}
                  </label>
                  <Select value={assigneeFilter} onValueChange={setAssigneeFilter}>
                    <SelectTrigger>
                      <SelectValue placeholder="Anyone" />
                    </SelectTrigger>
                    <SelectContent className="max-h-48">
                      <SelectItem value="">Anyone</SelectItem>
                      {getAvailableAssignees(type).map(user => (
                        <SelectItem key={user.id} value={user.login}>
                          <div className="flex items-center gap-2">
                            <img 
                              src={user.avatar_url} 
                              alt={user.login}
                              className="w-4 h-4 rounded-full"
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display = 'none';
                              }}
                            />
                            {user.login}
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Task Linkage Filter */}
                <div className="space-y-2">
                  <label className="text-sm font-medium flex items-center gap-2">
                    <Link className="h-3 w-3" />
                    Task Status
                  </label>
                  <Select value={linkedTasksFilter} onValueChange={(value) => setLinkedTasksFilter(value as 'all' | 'linked' | 'unlinked')}>
                    <SelectTrigger>
                      <SelectValue placeholder="All items" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All items</SelectItem>
                      <SelectItem value="linked">
                        <div className="flex items-center gap-2">
                          <Link className="h-3 w-3 text-green-600" />
                          Has linked tasks
                        </div>
                      </SelectItem>
                      <SelectItem value="unlinked">
                        <div className="flex items-center gap-2">
                          <XCircle className="h-3 w-3 text-gray-400" />
                          No linked tasks
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Filter Actions */}
              <div className="flex items-center justify-between pt-2 border-t">
                <div className="text-xs text-muted-foreground">
                  Showing {type === 'issues' ? filteredIssues.length : filteredPRs.length} of {type === 'issues' ? (issuesData?.issues?.length || 0) : (prsData?.pullRequests?.length || 0)} {type}
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={resetFilters}
                  className="text-xs"
                >
                  Clear all filters
                </Button>
              </div>
            </div>
          )}
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
                        <div className="space-y-2">
                          {/* Primary badges */}
                          <div className="flex flex-wrap items-center gap-2">
                            {issue && getIssueStateBadge(issue.state, issue.state_reason || undefined)}
                            {pr && getPRStateBadge(pr.state, pr.merged, pr.draft)}
                            
                            {/* Task linkage indicator */}
                            {getLinkedTaskCount(item) > 0 && (
                              <Badge className="bg-blue-100 text-blue-800 border-blue-300">
                                <Link className="h-3 w-3 mr-1" />
                                {getLinkedTaskCount(item)} task{getLinkedTaskCount(item) > 1 ? 's' : ''}
                              </Badge>
                            )}
                            
                            {/* Branch info for PRs */}
                            {pr && (
                              <Badge variant="outline" className="text-xs">
                                <GitBranch className="h-3 w-3 mr-1" />
                                {pr.head.ref} → {pr.base.ref}
                              </Badge>
                            )}
                          </div>

                          {/* Secondary info */}
                          <div className="flex flex-wrap items-center gap-2">
                            {/* Assignees with avatars */}
                            {item.assignees.length > 0 && (
                              <div className="flex items-center gap-1">
                                <Badge variant="outline" className="text-xs px-2 py-1">
                                  <UserCheck className="h-3 w-3 mr-1" />
                                  <div className="flex items-center gap-1">
                                    <span>{item.assignees.length} assigned</span>
                                    <div className="flex -space-x-1">
                                      {item.assignees.slice(0, 3).map((assignee, idx) => (
                                        <img
                                          key={assignee.id}
                                          src={assignee.avatar_url}
                                          alt={assignee.login}
                                          title={assignee.login}
                                          className="w-4 h-4 rounded-full border border-white"
                                          style={{ zIndex: 10 - idx }}
                                          onError={(e) => {
                                            (e.target as HTMLImageElement).style.display = 'none';
                                          }}
                                        />
                                      ))}
                                      {item.assignees.length > 3 && (
                                        <div className="w-4 h-4 rounded-full bg-gray-100 border border-white flex items-center justify-center text-xs text-gray-600">
                                          +{item.assignees.length - 3}
                                        </div>
                                      )}
                                    </div>
                                  </div>
                                </Badge>
                              </div>
                            )}

                            {/* Reviewers for PRs */}
                            {pr && pr.requested_reviewers.length > 0 && (
                              <div className="flex items-center gap-1">
                                <Badge variant="outline" className="text-xs px-2 py-1">
                                  <Eye className="h-3 w-3 mr-1" />
                                  <div className="flex items-center gap-1">
                                    <span>{pr.requested_reviewers.length} reviewer{pr.requested_reviewers.length > 1 ? 's' : ''}</span>
                                    <div className="flex -space-x-1">
                                      {pr.requested_reviewers.slice(0, 3).map((reviewer, idx) => (
                                        <img
                                          key={reviewer.id}
                                          src={reviewer.avatar_url}
                                          alt={reviewer.login}
                                          title={reviewer.login}
                                          className="w-4 h-4 rounded-full border border-white"
                                          style={{ zIndex: 10 - idx }}
                                          onError={(e) => {
                                            (e.target as HTMLImageElement).style.display = 'none';
                                          }}
                                        />
                                      ))}
                                      {pr.requested_reviewers.length > 3 && (
                                        <div className="w-4 h-4 rounded-full bg-gray-100 border border-white flex items-center justify-center text-xs text-gray-600">
                                          +{pr.requested_reviewers.length - 3}
                                        </div>
                                      )}
                                    </div>
                                  </div>
                                </Badge>
                              </div>
                            )}

                            {/* PR specific indicators */}
                            {pr && pr.draft && (
                              <Badge variant="outline" className="text-xs">
                                <GitCommit className="h-3 w-3 mr-1" />
                                Draft
                              </Badge>
                            )}

                            {/* Comments indicator */}
                            {item.comments > 0 && (
                              <Badge variant="outline" className="text-xs">
                                <MessageCircle className="h-3 w-3 mr-1" />
                                {item.comments}
                              </Badge>
                            )}
                          </div>

                          {/* Labels */}
                          {item.labels.length > 0 && (
                            <div className="flex flex-wrap items-center gap-1">
                              {item.labels.slice(0, 4).map(label => (
                                <Badge 
                                  key={label.id} 
                                  variant="outline" 
                                  className="text-xs px-2 py-0.5"
                                  style={{ 
                                    backgroundColor: `#${label.color}15`,
                                    borderColor: `#${label.color}`,
                                    color: `#${label.color}`
                                  }}
                                >
                                  <Tag className="h-2.5 w-2.5 mr-1" />
                                  {label.name}
                                </Badge>
                              ))}
                              {item.labels.length > 4 && (
                                <Badge variant="outline" className="text-xs px-2 py-0.5">
                                  +{item.labels.length - 4} more
                                </Badge>
                              )}
                            </div>
                          )}
                        </div>
                      </div>
                      <div className="flex gap-2 shrink-0">
                        {/* Create Task Button - Enhanced with context */}
                        {getLinkedTaskCount(item) > 0 ? (
                          <Button 
                            variant="secondary" 
                            size="sm"
                            onClick={() => {
                              if (issue) handleCreateTaskFromIssue(issue);
                              if (pr) handleCreateTaskFromPR(pr);
                            }}
                            className="bg-blue-50 text-blue-700 hover:bg-blue-100 border-blue-200"
                          >
                            <Plus className="h-4 w-4 mr-2" />
                            Add Another Task
                          </Button>
                        ) : (
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
                        )}
                      </div>
                    </div>
                  </CardHeader>
                  {item.body && (
                    <CardContent className="pt-0">
                      <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                        {item.body}
                      </p>
                      <div className="flex flex-wrap items-center justify-between gap-4 text-xs text-muted-foreground">
                        <div className="flex flex-wrap items-center gap-4">
                          <div className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            Created {new Date(item.created_at).toLocaleString()}
                          </div>
                          {item.updated_at !== item.created_at && (
                            <div className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              Updated {new Date(item.updated_at).toLocaleString()}
                            </div>
                          )}
                          <div className="flex items-center gap-1">
                            <img 
                              src={item.user.avatar_url} 
                              alt={item.user.login}
                              className="w-3 h-3 rounded-full"
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display = 'none';
                              }}
                            />
                            by {item.user.login}
                          </div>
                        </div>
                        <div className="flex items-center gap-4">
                          {pr && (
                            <>
                              {pr.additions > 0 && (
                                <div className="text-green-600">+{pr.additions}</div>
                              )}
                              {pr.deletions > 0 && (
                                <div className="text-red-600">-{pr.deletions}</div>
                              )}
                              <div>{pr.commits} commit{pr.commits !== 1 ? 's' : ''}</div>
                              <div>{pr.changed_files} file{pr.changed_files !== 1 ? 's' : ''}</div>
                            </>
                          )}
                          {item.comments > 0 && (
                            <div className="flex items-center gap-1">
                              <MessageCircle className="h-3 w-3" />
                              {item.comments}
                            </div>
                          )}
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

  // Render content based on activeTab from parent
  const renderContent = () => {
    if (activeTab === 'tasks') {
      return renderTasks();
    }

    if (activeTab === 'issues') {
      return renderGitHubContent(
        'issues',
        filteredIssues,
        issuesLoading,
        issuesError,
        issuesErrorData as Error,
        refetchIssues,
        issuesFilter,
        setIssuesFilter
      );
    }

    if (activeTab === 'prs') {
      return renderGitHubContent(
        'prs',
        filteredPRs,
        prsLoading,
        prsError,
        prsErrorData as Error,
        refetchPRs,
        prsFilter,
        setPrsFilter
      );
    }

    // Default to tasks
    return renderTasks();
  };

  return (
    <>
      {renderContent()}

    {/* Task Detail Modal */}
    <Dialog open={showTaskDetailDialog} onOpenChange={setShowTaskDetailDialog}>
      <DialogContent className="sm:max-w-[800px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {selectedTask && getStatusIcon(selectedTask.status)}
            {selectedTask?.title}
          </DialogTitle>
          <DialogDescription>
            Task details and content
          </DialogDescription>
        </DialogHeader>
        
        {selectedTask && (
          <div className="space-y-6">
            {/* Task Metadata */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium text-muted-foreground">Status</label>
                <div className="mt-1">
                  <Badge className={getStatusColor(selectedTask.status)}>
                    {selectedTask.status.replace('_', ' ')}
                  </Badge>
                </div>
              </div>
              
              <div>
                <label className="text-sm font-medium text-muted-foreground">Created</label>
                <div className="mt-1 text-sm">
                  {new Date(selectedTask.created_at).toLocaleString()}
                </div>
              </div>
              
              {selectedTask.branch_name && (
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Branch</label>
                  <div className="mt-1 flex items-center gap-1 text-sm">
                    <GitBranch className="h-3 w-3" />
                    {selectedTask.branch_name}
                  </div>
                </div>
              )}
              
              {selectedTask.pr_url && (
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Pull Request</label>
                  <div className="mt-1">
                    <a 
                      href={selectedTask.pr_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-blue-600 hover:text-blue-800 flex items-center gap-1"
                    >
                      <ExternalLink className="h-3 w-3" />
                      View PR
                    </a>
                  </div>
                </div>
              )}
            </div>

            {/* Task Content */}
            <div>
              <label className="text-sm font-medium text-muted-foreground">Content</label>
              <div className="mt-2 p-4 bg-muted rounded-lg">
                <pre className="whitespace-pre-wrap text-sm font-mono">
                  {selectedTask.description}
                </pre>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={handleCloseTaskDetail}>
                Close
              </Button>
              
              {/* Pending Task Actions */}
              {selectedTask.status === 'pending' && (
                <Button 
                  onClick={async () => {
                    try {
                      // Call backend API to execute the task
                      const response = await fetch(apiConfig.endpoints.tasks.execute(selectedTask.id), {
                        method: 'POST',
                        headers: {
                          'Content-Type': 'application/json',
                        }
                      });

                      if (!response.ok) {
                        const errorData = await response.text();
                        console.error('Start task API failed:', errorData);
                        throw new Error(`Start task API failed: ${errorData}`);
                      }

                      const result = await response.json();
                      console.log('Task started:', result);

                      // Immediately update task status in frontend for responsive UI
                      setTasks(prevTasks =>
                        prevTasks.map(task =>
                          task.id === selectedTask.id
                            ? { ...task, status: 'in_progress' as const }
                            : task
                        )
                      );

                      // Close modal immediately
                      handleCloseTaskDetail();

                      // Refresh tasks list after a delay to get actual backend status
                      setTimeout(async () => {
                        await loadTasks();
                      }, 1000);
                    } catch (error) {
                      console.error('Failed to start task:', error);
                    }
                  }}
                >
                  <Play className="h-4 w-4 mr-2" />
                  Start Task
                </Button>
              )}

              {/* In Progress Task Actions */}
              {selectedTask.status === 'in_progress' && (
                <>
                  <Button 
                    variant="destructive"
                    onClick={async () => {
                      try {
                        // Refresh tasks list
                        await loadTasks();

                        handleCloseTaskDetail();
                      } catch (error) {
                        console.error('Failed to cancel task:', error);
                      }
                    }}
                  >
                    <X className="h-4 w-4 mr-2" />
                    Cancel Task
                  </Button>
                  
                  <Button 
                    onClick={async () => {
                      try {
                        // Close modal and let user continue working
                        handleCloseTaskDetail();
                      } catch (error) {
                        console.error('Failed to resume task:', error);
                      }
                    }}
                  >
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Resume Task
                  </Button>
                </>
              )}

              {/* Completed Task Actions */}
              {selectedTask.status === 'completed' && selectedTask.pr_url && (
                <Button variant="outline" asChild>
                  <a href={selectedTask.pr_url} target="_blank" rel="noopener noreferrer">
                    <ExternalLink className="h-4 w-4 mr-2" />
                    View PR
                  </a>
                </Button>
              )}
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>

    {/* Execute Task Confirmation Dialog */}
    <Dialog open={showExecuteDialog} onOpenChange={setShowExecuteDialog}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Execute Task</DialogTitle>
          <DialogDescription>
            Are you sure you want to execute this task?
          </DialogDescription>
        </DialogHeader>
        {taskToExecute && (
          <div className="space-y-4">
            <div className="p-4 bg-gray-50 rounded-lg">
              <h4 className="font-medium text-sm">{taskToExecute.title}</h4>
              {taskToExecute.description && (
                <p className="text-sm text-gray-600 mt-1">{taskToExecute.description}</p>
              )}
            </div>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" onClick={handleCancelExecuteTask}>
                Cancel
              </Button>
              <Button onClick={handleConfirmExecuteTask}>
                <Play className="h-4 w-4 mr-2" />
                Start Execute
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
    </>
  );
}