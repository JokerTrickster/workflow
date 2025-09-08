'use client';

import { useState } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { Task } from '../../domain/entities/Task';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../components/ui/tabs';
import { Button } from '../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card';
import { Badge } from '../../components/ui/badge';
import { Input } from '../../components/ui/input';
import { Textarea } from '../../components/ui/textarea';
import { Plus, Play, ExternalLink, Calendar, GitBranch, Clock, CheckCircle, XCircle } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../../components/ui/dialog';

interface WorkspacePanelProps {
  repository: Repository;
  onClose: () => void;
}

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

export function WorkspacePanel({ repository, onClose }: WorkspacePanelProps) {
  const [tasks, setTasks] = useState<Task[]>(mockTasks);
  const [newTaskTitle, setNewTaskTitle] = useState('');
  const [newTaskDescription, setNewTaskDescription] = useState('');
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const handleCreateTask = () => {
    if (!newTaskTitle.trim()) return;

    const newTask: Task = {
      id: Date.now().toString(),
      title: newTaskTitle,
      description: newTaskDescription,
      status: 'pending',
      repository_id: repository.id,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    setTasks(prev => [newTask, ...prev]);
    setNewTaskTitle('');
    setNewTaskDescription('');
    setShowCreateDialog(false);
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

  return (
    <div className="fixed inset-0 bg-background z-50 flex flex-col safe-area-inset">
      {/* Header */}
      <div className="border-b bg-card">
        <div className="container mx-auto max-w-7xl px-4 py-4 px-safe-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <h1 className="text-xl md:text-2xl font-bold">{repository.name}</h1>
                <Badge variant={repository.private ? "secondary" : "outline"}>
                  {repository.private ? 'Private' : 'Public'}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">{repository.description}</p>
            </div>
            <Button variant="outline" onClick={onClose} className="shrink-0">
              ← Back
            </Button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 container mx-auto max-w-7xl px-4 py-6 px-safe-4">
        <Tabs defaultValue="tasks" className="w-full">
          <TabsList className="grid w-full grid-cols-3 mb-6">
            <TabsTrigger value="repo">Repository</TabsTrigger>
            <TabsTrigger value="tasks">Tasks</TabsTrigger>
            <TabsTrigger value="logs">Logs</TabsTrigger>
          </TabsList>

          <TabsContent value="repo" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <GitBranch className="h-5 w-5" />
                  Repository Details
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm font-medium">Language</p>
                    <p className="text-sm text-muted-foreground">{repository.language || 'N/A'}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium">Stars</p>
                    <p className="text-sm text-muted-foreground">{repository.stargazers_count}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium">Forks</p>
                    <p className="text-sm text-muted-foreground">{repository.forks_count}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium">Default Branch</p>
                    <p className="text-sm text-muted-foreground">{repository.default_branch}</p>
                  </div>
                </div>
                <div className="flex gap-2 pt-4">
                  <Button variant="outline" size="sm" asChild>
                    <a href={repository.html_url} target="_blank" rel="noopener noreferrer">
                      <ExternalLink className="h-4 w-4 mr-2" />
                      View on GitHub
                    </a>
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="tasks" className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <h2 className="text-lg font-semibold">Tasks</h2>
              <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogTrigger asChild>
                  <Button size="sm">
                    <Plus className="h-4 w-4 mr-2" />
                    New Task
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <DialogHeader>
                    <DialogTitle>Create New Task</DialogTitle>
                    <DialogDescription>
                      Create a new task for the AI agent to work on.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="space-y-4 pt-4">
                    <div>
                      <Input
                        placeholder="Task title"
                        value={newTaskTitle}
                        onChange={(e) => setNewTaskTitle(e.target.value)}
                      />
                    </div>
                    <div>
                      <Textarea
                        placeholder="Task description (optional)"
                        value={newTaskDescription}
                        onChange={(e) => setNewTaskDescription(e.target.value)}
                        rows={4}
                      />
                    </div>
                    <div className="flex justify-end gap-2">
                      <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                        Cancel
                      </Button>
                      <Button onClick={handleCreateTask}>
                        Create Task
                      </Button>
                    </div>
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
          </TabsContent>

          <TabsContent value="logs" className="space-y-4">
            <div className="text-center py-12">
              <p className="text-muted-foreground">No logs available yet</p>
              <p className="text-sm text-muted-foreground mt-2">
                Activity logs will appear here when tasks are executed
              </p>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}