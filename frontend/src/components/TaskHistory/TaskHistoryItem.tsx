import React, { useState } from 'react';
import { WorkflowHistory } from '../../services/claudeService';
import { StatusBadge } from './StatusBadge';

interface TaskHistoryItemProps {
  task: WorkflowHistory;
}

/**
 * Individual task history item component displaying task details and expandable content
 */
export function TaskHistoryItem({ task }: TaskHistoryItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const formatDuration = (ms?: number) => {
    if (!ms) return 'N/A';
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const truncateText = (text: string, maxLength: number = 100) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

  return (
    <div className="bg-white border border-gray-200 rounded-lg shadow-sm hover:shadow-md transition-shadow">
      <div className="p-4">
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-2">
              <StatusBadge status={task.status} />
              <span className="text-xs text-gray-500 font-mono">
                {task.request_id.substring(0, 8)}...
              </span>
            </div>
            <p className="text-sm text-gray-900 leading-relaxed">
              {isExpanded ? task.tasks : truncateText(task.tasks)}
            </p>
          </div>

          {task.tasks.length > 100 && (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="ml-2 text-indigo-600 hover:text-indigo-800 text-sm font-medium flex-shrink-0"
            >
              {isExpanded ? 'Show less' : 'Show more'}
            </button>
          )}
        </div>

        {/* Metadata */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs text-gray-600 mb-3">
          <div>
            <span className="font-medium text-gray-700">Created:</span>
            <br />
            {formatDate(task.created_at)}
          </div>

          {task.completed_at && (
            <div>
              <span className="font-medium text-gray-700">Completed:</span>
              <br />
              {formatDate(task.completed_at)}
            </div>
          )}

          <div>
            <span className="font-medium text-gray-700">Duration:</span>
            <br />
            {formatDuration(task.processing_time_ms)}
          </div>

          <div className="flex items-center gap-2">
            {task.interactive && (
              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-blue-100 text-blue-800">
                Interactive
              </span>
            )}
            {task.continue_task && (
              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-purple-100 text-purple-800">
                Continue
              </span>
            )}
          </div>
        </div>

        {/* Working Directory */}
        {task.working_dir && (
          <div className="mb-3">
            <span className="text-xs font-medium text-gray-700">Working Directory: </span>
            <code className="text-xs bg-gray-100 px-2 py-1 rounded font-mono">
              {task.working_dir}
            </code>
          </div>
        )}

        {/* Error Display */}
        {task.error && (
          <div className="mb-3 p-3 bg-red-50 border border-red-200 rounded-md">
            <h4 className="text-sm font-medium text-red-800 mb-1">Error:</h4>
            <pre className="text-xs text-red-700 whitespace-pre-wrap overflow-auto max-h-32">
              {task.error}
            </pre>
          </div>
        )}

        {/* Result Display */}
        {task.result && (
          <div className="mb-3">
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="text-sm font-medium text-gray-700 hover:text-gray-900 mb-2 flex items-center gap-1"
            >
              <span>Result:</span>
              <svg
                className={`w-4 h-4 transform transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            {isExpanded && (
              <pre className="text-xs bg-gray-50 border p-3 rounded font-mono overflow-auto max-h-60">
                {task.result}
              </pre>
            )}
          </div>
        )}

        {/* Actions */}
        <div className="flex justify-end pt-2 border-t border-gray-100">
          <button className="text-xs text-gray-500 hover:text-gray-700 font-medium">
            View Details
          </button>
        </div>
      </div>
    </div>
  );
}