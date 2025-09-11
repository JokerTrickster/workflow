import { promises as fs } from 'fs';
import path from 'path';
import { WorkLogEntry } from '../../../services/WorkLogManager';

// Utility functions for work logs API
export const getLogsDir = (repository: string) => {
  return path.join(process.cwd(), '.claude', 'logs', repository);
};

export const ensureDirectoryExists = async (dirPath: string) => {
  try {
    await fs.mkdir(dirPath, { recursive: true });
  } catch (error) {
    console.error('Failed to create directory:', error);
    throw new Error(`Cannot create logs directory: ${dirPath}`);
  }
};

export const getLogFilePath = (repository: string, date: string) => {
  const logsDir = getLogsDir(repository);
  return path.join(logsDir, `${date}.md`);
};

export const formatLogEntry = (entry: WorkLogEntry): string => {
  const time = new Date(entry.timestamp).toLocaleTimeString('en-US', { hour12: false });
  let markdown = `### ${time} - ${entry.taskTitle} (${entry.status})\n\n`;
  
  if (entry.progressUpdate) {
    markdown += `**Progress**: ${entry.progressUpdate}\n\n`;
  }
  
  if (entry.issuesDiscovered && entry.issuesDiscovered.length > 0) {
    markdown += `**Issues Discovered**:\n${entry.issuesDiscovered.map(issue => `- ${issue}`).join('\n')}\n\n`;
  }
  
  if (entry.improvementsMade && entry.improvementsMade.length > 0) {
    markdown += `**Improvements Made**:\n${entry.improvementsMade.map(improvement => `- ${improvement}`).join('\n')}\n\n`;
  }
  
  if (entry.metadata) {
    markdown += `**Metadata**:\n`;
    if (entry.metadata.branch) markdown += `- Branch: ${entry.metadata.branch}\n`;
    if (entry.metadata.githubIssue) markdown += `- GitHub Issue: #${entry.metadata.githubIssue}\n`;
    if (entry.metadata.prUrl) markdown += `- PR URL: ${entry.metadata.prUrl}\n`;
    if (entry.metadata.tokensUsed) markdown += `- Tokens Used: ${entry.metadata.tokensUsed}\n`;
    markdown += '\n';
  }
  
  markdown += '---\n\n';
  return markdown;
};

export const getDailyLogHeader = (repository: string, date: string): string => {
  return `# Work Log - ${repository} - ${date}

Generated automatically by Claude Code workflow system.

---

`;
};

export const validateRepository = (repository: string): void => {
  if (!repository || typeof repository !== 'string') {
    throw new Error('Repository parameter is required and must be a string');
  }
  
  // Sanitize repository name to prevent path traversal
  if (repository.includes('..') || repository.includes('/') || repository.includes('\\')) {
    throw new Error('Invalid repository name: contains illegal characters');
  }
};

export const validateWorkLogEntry = (entry: WorkLogEntry): void => {
  if (!entry || typeof entry !== 'object') {
    throw new Error('Entry is required and must be an object');
  }
  
  if (!entry.taskId || !entry.taskTitle || !entry.repository) {
    throw new Error('Entry must have taskId, taskTitle, and repository');
  }
  
  if (!['pending', 'in_progress', 'completed', 'failed'].includes(entry.status)) {
    throw new Error('Entry status must be one of: pending, in_progress, completed, failed');
  }
};

// Simple file lock mechanism to prevent concurrent writes
const fileLocks = new Map<string, Promise<void>>();

export const writeFileWithLock = async (filePath: string, content: string): Promise<void> => {
  // Get or create a lock for this file
  const existingLock = fileLocks.get(filePath);
  
  const writeOperation = async () => {
    try {
      await fs.writeFile(filePath, content, 'utf-8');
    } finally {
      // Clean up the lock when done
      if (fileLocks.get(filePath) === lockPromise) {
        fileLocks.delete(filePath);
      }
    }
  };

  const lockPromise = existingLock 
    ? existingLock.then(writeOperation)
    : writeOperation();
    
  fileLocks.set(filePath, lockPromise);
  
  return lockPromise;
};