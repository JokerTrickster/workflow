import { NextRequest, NextResponse } from 'next/server';
import { promises as fs } from 'fs';
import path from 'path';
import { WorkLogEntry } from '../../../../services/WorkLogManager';

// This is the endpoint for creating work log entries
// GET requests are handled by the main work-logs route
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { repository, entry } = body as { repository: string; entry: WorkLogEntry };

    if (!repository || !entry) {
      return NextResponse.json({ error: 'Repository and entry are required' }, { status: 400 });
    }

    const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
    const logsDir = path.join(process.cwd(), '.claude', 'logs', repository);
    const logFilePath = path.join(logsDir, `${today}.md`);

    // Ensure directory exists
    await fs.mkdir(logsDir, { recursive: true });

    // Format the log entry as markdown
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

    // Check if log file exists for today
    let existingContent = '';
    try {
      existingContent = await fs.readFile(logFilePath, 'utf-8');
    } catch {
      // File doesn't exist, create with header
      existingContent = `# Work Log - ${repository} - ${today}\n\nGenerated automatically by Claude Code workflow system.\n\n---\n\n`;
    }

    // Append new entry
    const updatedContent = existingContent + markdown;

    // Write updated content
    await fs.writeFile(logFilePath, updatedContent, 'utf-8');

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Failed to write work log entry:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}