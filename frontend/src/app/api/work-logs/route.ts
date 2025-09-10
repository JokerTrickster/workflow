import { NextRequest, NextResponse } from 'next/server';
import { promises as fs } from 'fs';
import path from 'path';
import { WorkLogEntry, DailyWorkLog } from '../../../services/WorkLogManager';

// Ensure .claude/logs directory exists
const getLogsDir = (repository: string) => {
  return path.join(process.cwd(), '.claude', 'logs', repository);
};

const ensureDirectoryExists = async (dirPath: string) => {
  try {
    await fs.mkdir(dirPath, { recursive: true });
  } catch (error) {
    console.error('Failed to create directory:', error);
  }
};

const getLogFilePath = (repository: string, date: string) => {
  const logsDir = getLogsDir(repository);
  return path.join(logsDir, `${date}.md`);
};

const formatLogEntry = (entry: WorkLogEntry): string => {
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

const getDailyLogHeader = (repository: string, date: string): string => {
  return `# Work Log - ${repository} - ${date}

Generated automatically by Claude Code workflow system.

---

`;
};

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const repository = searchParams.get('repository');
    const startDate = searchParams.get('startDate');
    const endDate = searchParams.get('endDate');

    if (!repository) {
      return NextResponse.json({ error: 'Repository parameter is required' }, { status: 400 });
    }

    const logsDir = getLogsDir(repository);

    // Check if directory exists
    try {
      await fs.access(logsDir);
    } catch {
      // Directory doesn't exist, return empty array
      return NextResponse.json([]);
    }

    const files = await fs.readdir(logsDir);
    const logFiles = files.filter(file => file.endsWith('.md'));

    const workLogs: DailyWorkLog[] = [];

    for (const file of logFiles) {
      const date = file.replace('.md', '');
      
      // Filter by date range if provided
      if (startDate && date < startDate) continue;
      if (endDate && date > endDate) continue;

      const filePath = path.join(logsDir, file);
      const content = await fs.readFile(filePath, 'utf-8');
      
      // Parse markdown content back to entries (simplified)
      // For now, return basic structure - full parsing can be added later
      workLogs.push({
        date,
        repository,
        entries: [], // TODO: Parse markdown back to entries if needed
      });
    }

    return NextResponse.json(workLogs);
  } catch (error) {
    console.error('Failed to get work logs:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { repository, entry } = body as { repository: string; entry: WorkLogEntry };

    if (!repository || !entry) {
      return NextResponse.json({ error: 'Repository and entry are required' }, { status: 400 });
    }

    const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
    const logsDir = getLogsDir(repository);
    const logFilePath = getLogFilePath(repository, today);

    // Ensure directory exists
    await ensureDirectoryExists(logsDir);

    // Check if log file exists for today
    let existingContent = '';
    try {
      existingContent = await fs.readFile(logFilePath, 'utf-8');
    } catch {
      // File doesn't exist, create with header
      existingContent = getDailyLogHeader(repository, today);
    }

    // Append new entry
    const entryContent = formatLogEntry(entry);
    const updatedContent = existingContent + entryContent;

    // Write updated content
    await fs.writeFile(logFilePath, updatedContent, 'utf-8');

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Failed to write work log entry:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}