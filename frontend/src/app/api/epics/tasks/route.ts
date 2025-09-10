import { NextRequest, NextResponse } from 'next/server';
import fs from 'fs/promises';
import path from 'path';
import matter from 'gray-matter';

export interface TaskFileMetadata {
  id: string;
  title: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  epic: string;
  branch?: string;
  createdAt: string;
  updatedAt: string;
  startedAt?: string;
  completedAt?: string;
  tokensUsed: number;
  githubIssue?: number;
  prUrl?: string;
  buildStatus?: 'pending' | 'success' | 'failure';
  lintStatus?: 'pending' | 'success' | 'failure';
}

export interface TaskFile {
  metadata: TaskFileMetadata;
  content: string;
}

const TASKS_DIR = path.resolve(process.cwd(), '../.claude/epics/tasks');

// Ensure tasks directory exists
async function ensureTasksDir() {
  try {
    await fs.mkdir(TASKS_DIR, { recursive: true });
  } catch (error) {
    console.error('Failed to create tasks directory:', error);
  }
}

// Get all task files
export async function GET() {
  try {
    await ensureTasksDir();
    
    const files = await fs.readdir(TASKS_DIR);
    const taskFiles = files.filter(file => file.endsWith('.md') && file !== 'task-template.md');
    
    const tasks: TaskFile[] = [];
    
    for (const filename of taskFiles) {
      try {
        const filePath = path.join(TASKS_DIR, filename);
        const content = await fs.readFile(filePath, 'utf-8');
        const { data: frontMatter, content: markdownContent } = matter(content);
        
        const taskFile: TaskFile = {
          metadata: frontMatter as TaskFileMetadata,
          content: markdownContent.trim(),
        };
        
        tasks.push(taskFile);
      } catch (error) {
        console.error(`Failed to read task file ${filename}:`, error);
      }
    }
    
    // Sort by updated date (newest first)
    tasks.sort((a, b) => new Date(b.metadata.updatedAt).getTime() - new Date(a.metadata.updatedAt).getTime());
    
    return NextResponse.json(tasks);
  } catch (error) {
    console.error('Failed to load tasks:', error);
    return NextResponse.json({ error: 'Failed to load tasks' }, { status: 500 });
  }
}

// Create new task file
export async function POST(request: NextRequest) {
  try {
    await ensureTasksDir();
    
    const taskFile: TaskFile = await request.json();
    
    if (!taskFile.metadata.id || !taskFile.metadata.title) {
      return NextResponse.json({ error: 'Task ID and title are required' }, { status: 400 });
    }
    
    const filename = `${taskFile.metadata.id}.md`;
    const filePath = path.join(TASKS_DIR, filename);
    
    // Check if file already exists
    try {
      await fs.access(filePath);
      return NextResponse.json({ error: 'Task file already exists' }, { status: 409 });
    } catch {
      // File doesn't exist, continue
    }
    
    // Generate markdown content with frontmatter
    const frontMatter = Object.entries(taskFile.metadata)
      .map(([key, value]) => {
        if (value === undefined || value === null) return null;
        if (typeof value === 'string' && value.includes('\n')) {
          return `${key}: |\n  ${value.replace(/\n/g, '\n  ')}`;
        }
        return `${key}: ${JSON.stringify(value)}`;
      })
      .filter(Boolean)
      .join('\n');
    
    const fileContent = `---\n${frontMatter}\n---\n\n${taskFile.content}`;
    
    await fs.writeFile(filePath, fileContent, 'utf-8');
    
    return NextResponse.json(taskFile, { status: 201 });
  } catch (error) {
    console.error('Failed to create task:', error);
    return NextResponse.json({ error: 'Failed to create task' }, { status: 500 });
  }
}