import { NextRequest, NextResponse } from 'next/server';
import { promises as fs } from 'fs';
import { WorkLogEntry } from '../../../../services/WorkLogManager';
import { 
  validateRepository, 
  validateWorkLogEntry, 
  getLogFilePath, 
  ensureDirectoryExists, 
  formatLogEntry, 
  getDailyLogHeader,
  getLogsDir
} from '../utils';

// This is the endpoint for creating work log entries
// GET requests are handled by the main work-logs route
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { repository, entry } = body as { repository: string; entry: WorkLogEntry };

    // Validate inputs
    validateRepository(repository);
    validateWorkLogEntry(entry);

    const today = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
    const logsDir = getLogsDir(repository);
    const logFilePath = getLogFilePath(repository, today);

    // Ensure directory exists
    await ensureDirectoryExists(logsDir);

    // Format the log entry as markdown
    const entryMarkdown = formatLogEntry(entry);

    // Check if log file exists for today
    let existingContent = '';
    try {
      existingContent = await fs.readFile(logFilePath, 'utf-8');
    } catch {
      // File doesn't exist, create with header
      existingContent = getDailyLogHeader(repository, today);
    }

    // Append new entry
    const updatedContent = existingContent + entryMarkdown;

    // Write updated content
    await fs.writeFile(logFilePath, updatedContent, 'utf-8');

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Failed to write work log entry:', error);
    
    // Return appropriate error based on error type
    if (error instanceof Error) {
      if (error.message.includes('Invalid repository') || error.message.includes('Entry must have')) {
        return NextResponse.json({ error: error.message }, { status: 400 });
      }
    }
    
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}