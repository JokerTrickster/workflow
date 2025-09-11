'use client';

import { useState, useCallback, useMemo } from 'react';
import { Task } from '../domain/entities/Task';
import { GitHubIssue, GitHubPullRequest } from '../types/github';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Textarea } from './ui/textarea';
import { Badge } from './ui/badge';
import { 
  ExternalLink, 
  GitBranch,
  AlertCircle,
  GitPullRequest,
  Calendar,
  Users,
  Loader2
} from 'lucide-react';

interface TaskCreationFormProps {
  repositoryId: number;
  onSubmit: (taskData: Omit<Task, 'id' | 'created_at' | 'updated_at'>) => Promise<void>;
  onCancel: () => void;
  githubIssue?: GitHubIssue;
  githubPullRequest?: GitHubPullRequest;
  isSubmitting?: boolean;
}

export function TaskCreationForm({
  repositoryId,
  onSubmit,
  onCancel,
  githubIssue,
  githubPullRequest,
  isSubmitting = false
}: TaskCreationFormProps) {
  const githubItem = githubIssue || githubPullRequest;
  const isFromGitHub = Boolean(githubItem);
  
  // Initialize form with GitHub data if available
  const [title, setTitle] = useState(() => {
    if (githubIssue) return `Issue #${githubIssue.number}: ${githubIssue.title}`;
    if (githubPullRequest) return `PR #${githubPullRequest.number}: ${githubPullRequest.title}`;
    return '';
  });

  const [description, setDescription] = useState(() => {
    if (githubIssue) {
      return `GitHub Issue: ${githubIssue.html_url}\n\n${githubIssue.body || 'No description provided'}`;
    }
    if (githubPullRequest) {
      return `GitHub PR: ${githubPullRequest.html_url}\nBranch: ${githubPullRequest.head.ref} → ${githubPullRequest.base.ref}\n\n${githubPullRequest.body || 'No description provided'}`;
    }
    return '';
  });

  const [branchName, setBranchName] = useState(() => {
    if (githubPullRequest) return githubPullRequest.head.ref;
    return '';
  });

  // Validation errors state
  const [errors, setErrors] = useState<{
    title?: string;
    description?: string;
    branchName?: string;
  }>({});

  // Clear specific error when user starts typing (memoized)
  const clearError = useCallback((field: keyof typeof errors) => {
    setErrors(prev => {
      if (!prev[field]) return prev; // No change needed
      return { ...prev, [field]: undefined };
    });
  }, []);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate mandatory fields
    const newErrors: typeof errors = {};
    
    if (!title.trim()) {
      newErrors.title = 'Task title is required';
    }
    
    if (!description.trim()) {
      newErrors.description = 'Task description is required';
    }
    
    if (!branchName.trim()) {
      newErrors.branchName = 'Branch name is required';
    }
    
    setErrors(newErrors);
    
    // Stop if there are validation errors
    if (Object.keys(newErrors).length > 0) return;

    const taskData: Omit<Task, 'id' | 'created_at' | 'updated_at'> = {
      title: title.trim(),
      description: description.trim(),
      status: 'pending',
      repository_id: repositoryId,
      branch_name: branchName.trim() || undefined,
      pr_url: githubPullRequest?.html_url || undefined,
    };

    await onSubmit(taskData);
  }, [title, description, branchName, repositoryId, githubPullRequest, onSubmit]);

  const getGitHubMetadata = () => {
    if (!githubItem) return null;

    return (
      <div className="space-y-3 p-4 bg-muted/50 rounded-lg border">
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2 min-w-0">
            {githubIssue && <AlertCircle className="h-4 w-4 text-green-600 shrink-0" />}
            {githubPullRequest && <GitPullRequest className="h-4 w-4 text-blue-600 shrink-0" />}
            <span className="text-sm font-medium truncate">
              {githubIssue ? 'GitHub Issue' : 'GitHub Pull Request'} #{githubItem.number}
            </span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            asChild
            className="shrink-0"
          >
            <a
              href={githubItem.html_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1"
            >
              <ExternalLink className="h-3 w-3" />
            </a>
          </Button>
        </div>

        <div className="flex flex-wrap gap-2">
          <Badge 
            className={
              githubItem.state === 'open' 
                ? 'bg-green-100 text-green-800' 
                : githubPullRequest?.merged 
                ? 'bg-purple-100 text-purple-800'
                : 'bg-red-100 text-red-800'
            }
          >
            {githubPullRequest?.merged ? 'Merged' : githubItem.state}
          </Badge>

          {githubPullRequest && (
            <Badge variant="outline" className="text-xs">
              <GitBranch className="h-3 w-3 mr-1" />
              {githubPullRequest.head.ref} → {githubPullRequest.base.ref}
            </Badge>
          )}

          {githubItem.assignees.length > 0 && (
            <Badge variant="outline" className="text-xs">
              <Users className="h-3 w-3 mr-1" />
              {githubItem.assignees.length} assigned
            </Badge>
          )}
        </div>

        {githubItem.labels.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {githubItem.labels.slice(0, 3).map(label => (
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
            {githubItem.labels.length > 3 && (
              <Badge variant="outline" className="text-xs">
                +{githubItem.labels.length - 3} more
              </Badge>
            )}
          </div>
        )}

        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <div className="flex items-center gap-1">
            <Calendar className="h-3 w-3" />
            {new Date(githubItem.created_at).toLocaleDateString()}
          </div>
          <div>by {githubItem.user.login}</div>
          {githubItem.comments > 0 && <div>{githubItem.comments} comments</div>}
        </div>
      </div>
    );
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {isFromGitHub && getGitHubMetadata()}

      <div className="space-y-2">
        <label htmlFor="title" className="text-sm font-medium">
          Task Title <span className="text-red-500">*</span>
        </label>
        <Input
          id="title"
          placeholder="Enter task title"
          value={title}
          onChange={(e) => {
            setTitle(e.target.value);
            clearError('title');
          }}
          required
          disabled={isSubmitting}
          className={errors.title ? 'border-red-500 focus:border-red-500' : ''}
        />
        {errors.title && (
          <p className="text-sm text-red-600 mt-1">{errors.title}</p>
        )}
      </div>

      <div className="space-y-2">
        <label htmlFor="description" className="text-sm font-medium">
          Description <span className="text-red-500">*</span>
        </label>
        <Textarea
          id="description"
          placeholder="Enter task description"
          value={description}
          onChange={(e) => {
            setDescription(e.target.value);
            clearError('description');
          }}
          rows={4}
          disabled={isSubmitting}
          required
          className={errors.description ? 'border-red-500 focus:border-red-500' : ''}
        />
        {errors.description && (
          <p className="text-sm text-red-600 mt-1">{errors.description}</p>
        )}
      </div>

      <div className="space-y-2">
        <label htmlFor="branchName" className="text-sm font-medium">
          Branch Name <span className="text-red-500">*</span>
        </label>
        <Input
          id="branchName"
          placeholder="feature/branch-name"
          value={branchName}
          onChange={(e) => {
            setBranchName(e.target.value);
            clearError('branchName');
          }}
          disabled={isSubmitting || Boolean(githubPullRequest)}
          required
          className={errors.branchName ? 'border-red-500 focus:border-red-500' : ''}
        />
        {errors.branchName && (
          <p className="text-sm text-red-600 mt-1">{errors.branchName}</p>
        )}
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button 
          type="button" 
          variant="outline" 
          onClick={onCancel}
          disabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button 
          type="submit" 
          disabled={!title.trim() || !description.trim() || !branchName.trim() || isSubmitting}
        >
          {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
          Create Task
        </Button>
      </div>
    </form>
  );
}