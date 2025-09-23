import React, { useState } from 'react';
import { useClaude, UseClaudeResult } from '../hooks/useClaude';
import { ReqRunTasksClaude } from '../services/claudeService';

interface ClaudeTaskRunnerProps {
  defaultRepository?: string;
  defaultWorkingDir?: string;
}

/**
 * React component for running Claude tasks
 */
export function ClaudeTaskRunner({ 
  defaultRepository = '', 
  defaultWorkingDir = '' 
}: ClaudeTaskRunnerProps) {
  const claude = useClaude();
  const [formData, setFormData] = useState<ReqRunTasksClaude>({
    tasks: '',
    repository_name: defaultRepository,
    provider: 'claude',
    working_dir: defaultWorkingDir,
    interactive: false,
    cmd: '',
    continue_task: false,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.tasks.trim() || !formData.repository_name.trim()) {
      alert('Please provide both tasks and repository name');
      return;
    }

    try {
      await claude.runTasksAndWait(formData);
    } catch (error) {
      console.error('Task execution failed:', error);
    }
  };

  const handleCancel = async () => {
    if (claude.currentTask) {
      try {
        await claude.cancelTask(claude.currentTask.request_id);
      } catch (error) {
        console.error('Failed to cancel task:', error);
      }
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-600';
      case 'failed': return 'text-red-600';
      case 'cancelled': return 'text-yellow-600';
      case 'processing': return 'text-blue-600';
      default: return 'text-gray-600';
    }
  };

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-6">Claude Task Runner</h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="tasks" className="block text-sm font-medium text-gray-700">
            Tasks (Required)
          </label>
          <textarea
            id="tasks"
            value={formData.tasks}
            onChange={(e) => setFormData({ ...formData, tasks: e.target.value })}
            placeholder="Describe the tasks you want Claude to execute..."
            rows={4}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            required
          />
        </div>

        <div>
          <label htmlFor="repository_name" className="block text-sm font-medium text-gray-700">
            Repository Name (Required)
          </label>
          <input
            type="text"
            id="repository_name"
            value={formData.repository_name}
            onChange={(e) => setFormData({ ...formData, repository_name: e.target.value })}
            placeholder="e.g., my-project"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            required
          />
        </div>

        <div>
          <label htmlFor="working_dir" className="block text-sm font-medium text-gray-700">
            Working Directory (Optional)
          </label>
          <input
            type="text"
            id="working_dir"
            value={formData.working_dir}
            onChange={(e) => setFormData({ ...formData, working_dir: e.target.value })}
            placeholder="e.g., /path/to/project"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>

        <div>
          <label htmlFor="provider" className="block text-sm font-medium text-gray-700">
            AI Provider
          </label>
          <select
            id="provider"
            value={formData.provider}
            onChange={(e) => setFormData({ ...formData, provider: e.target.value as 'claude' | 'codex' | 'cursor' })}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            <option value="claude">Claude</option>
            <option value="codex">Codex</option>
            <option value="cursor">Cursor</option>
          </select>
        </div>

        <div>
          <label htmlFor="cmd" className="block text-sm font-medium text-gray-700">
            Command (Optional)
          </label>
          <input
            type="text"
            id="cmd"
            value={formData.cmd}
            onChange={(e) => setFormData({ ...formData, cmd: e.target.value })}
            placeholder="e.g., echo test"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>

        <div className="flex space-x-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={formData.interactive}
              onChange={(e) => setFormData({ ...formData, interactive: e.target.checked })}
              className="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            />
            <span className="ml-2 text-sm text-gray-700">Interactive Mode</span>
          </label>

          <label className="flex items-center">
            <input
              type="checkbox"
              checked={formData.continue_task}
              onChange={(e) => setFormData({ ...formData, continue_task: e.target.checked })}
              className="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
            />
            <span className="ml-2 text-sm text-gray-700">Continue Previous Task</span>
          </label>
        </div>

        <div className="flex space-x-3">
          <button
            type="submit"
            disabled={claude.isLoading || claude.isPolling}
            className="flex-1 bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {claude.isLoading || claude.isPolling ? 'Processing...' : 'Run Tasks'}
          </button>

          {(claude.isLoading || claude.isPolling) && claude.currentTask && (
            <button
              type="button"
              onClick={handleCancel}
              className="bg-red-600 text-white py-2 px-4 rounded-md hover:bg-red-700"
            >
              Cancel
            </button>
          )}

          <button
            type="button"
            onClick={claude.reset}
            className="bg-gray-600 text-white py-2 px-4 rounded-md hover:bg-gray-700"
          >
            Reset
          </button>
        </div>
      </form>

      {/* Error Display */}
      {claude.error && (
        <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-md">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <div className="mt-2 text-sm text-red-700">
                {claude.error}
              </div>
              <div className="mt-2">
                <button
                  onClick={claude.clearError}
                  className="text-sm text-red-800 hover:text-red-900 underline"
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Task Status Display */}
      {claude.taskStatus && (
        <div className="mt-6 p-4 bg-gray-50 border border-gray-200 rounded-md">
          <h3 className="text-lg font-semibold mb-3">Task Status</h3>
          
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="font-medium">Request ID:</span>
              <span className="ml-2 text-gray-600">{claude.taskStatus.request_id}</span>
            </div>
            
            <div>
              <span className="font-medium">Status:</span>
              <span className={`ml-2 font-medium ${getStatusColor(claude.taskStatus.status)}`}>
                {claude.taskStatus.status.toUpperCase()}
              </span>
            </div>
            
            <div>
              <span className="font-medium">Created:</span>
              <span className="ml-2 text-gray-600">
                {new Date(claude.taskStatus.created_at).toLocaleString()}
              </span>
            </div>
            
            {claude.taskStatus.completed_at && (
              <div>
                <span className="font-medium">Completed:</span>
                <span className="ml-2 text-gray-600">
                  {new Date(claude.taskStatus.completed_at).toLocaleString()}
                </span>
              </div>
            )}
            
            {claude.taskStatus.processing_time_ms && (
              <div>
                <span className="font-medium">Processing Time:</span>
                <span className="ml-2 text-gray-600">
                  {(claude.taskStatus.processing_time_ms / 1000).toFixed(2)}s
                </span>
              </div>
            )}
          </div>

          {/* Task Result */}
          {claude.taskStatus.result && (
            <div className="mt-4">
              <h4 className="font-medium text-gray-700">Result:</h4>
              <pre className="mt-2 p-3 bg-white border rounded text-sm overflow-auto max-h-60">
                {JSON.stringify(claude.taskStatus.result, null, 2)}
              </pre>
            </div>
          )}

          {/* Task Error */}
          {claude.taskStatus.error && (
            <div className="mt-4">
              <h4 className="font-medium text-red-700">Error:</h4>
              <div className="mt-2 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                {claude.taskStatus.error}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}