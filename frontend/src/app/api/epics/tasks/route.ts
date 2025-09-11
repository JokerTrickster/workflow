import { NextRequest, NextResponse } from 'next/server';
import fs from 'fs/promises';
import path from 'path';
import matter from 'gray-matter';

// Force dynamic rendering - no static generation
export const dynamic = 'force-dynamic';
export const revalidate = 0;

export interface TaskFileMetadata {
  id: string;
  title: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  repository: string;
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

const EPICS_BASE_DIR = path.resolve(process.cwd(), '../.claude/epics');

// Get current repository name from git remote or default
async function getCurrentRepository(): Promise<string> {
  try {
    // Try to get from environment or default to workflow
    return process.env.REPOSITORY_NAME || 'workflow';
  } catch (error) {
    console.warn('Failed to get repository name, using default:', error);
    return 'workflow';
  }
}

// Get tasks directory for specific repository
function getTasksDir(repository: string): string {
  return path.join(EPICS_BASE_DIR, 'repositories', repository, 'tasks');
}

// Ensure tasks directory exists for repository
async function ensureTasksDir(repository: string) {
  try {
    const tasksDir = getTasksDir(repository);
    await fs.mkdir(tasksDir, { recursive: true });
  } catch (error) {
    console.error('Failed to create tasks directory:', error);
  }
}

// Get all task files for current repository
export async function GET(request: NextRequest) {
  try {
    // Get repository from query parameter or default to 'workflow'
    const { searchParams } = new URL(request.url);
    const repository = searchParams.get('repository') || 'workflow';
    const TASKS_DIR = getTasksDir(repository);
    
    await ensureTasksDir(repository);
    
    // Check if directory exists
    try {
      await fs.access(TASKS_DIR);
    } catch {
      // Directory doesn't exist, return empty array
      return NextResponse.json([]);
    }
    
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
        
        // Ensure repository field is set correctly
        if (!taskFile.metadata.repository) {
          taskFile.metadata.repository = repository;
        }
        
        tasks.push(taskFile);
      } catch (error) {
        console.error(`Failed to read task file ${filename}:`, error);
      }
    }
    
    // Sort by updated date (newest first)
    tasks.sort((a, b) => new Date(b.metadata.updatedAt).getTime() - new Date(a.metadata.updatedAt).getTime());
    
    const response = NextResponse.json(tasks);
    
    // Add ultra-aggressive cache-busting headers
    response.headers.set('Cache-Control', 'no-cache, no-store, must-revalidate, max-age=0, private, no-transform');
    response.headers.set('Pragma', 'no-cache');
    response.headers.set('Expires', '0');
    response.headers.set('Last-Modified', new Date().toUTCString());
    response.headers.set('ETag', `"${Date.now()}-${Math.random().toString(36).substring(2)}-${repository}"`);
    response.headers.set('X-Content-Type-Options', 'nosniff');
    response.headers.set('X-Force-Fresh', 'true');
    response.headers.set('X-No-Cache', 'true');
    response.headers.set('X-Repository', repository);
    response.headers.set('X-Timestamp', Date.now().toString());
    response.headers.set('Vary', 'Accept, Accept-Encoding, Accept-Language, User-Agent');
    response.headers.set('X-Cache-Status', 'MISS');
    response.headers.set('X-Served-By', 'fresh-data');
    console.log('🚛 API Response: Serving fresh data for repository:', repository, 'with', tasks.length, 'tasks');
    
    return response;
  } catch (error) {
    console.error('Failed to load tasks:', error);
    return NextResponse.json({ error: 'Failed to load tasks' }, { status: 500 });
  }
}

// Create new task file
export async function POST(request: NextRequest) {
  try {
    // Get repository from query parameter or default to 'workflow'
    const { searchParams } = new URL(request.url);
    const repository = searchParams.get('repository') || 'workflow';
    const TASKS_DIR = getTasksDir(repository);
    
    await ensureTasksDir(repository);
    
    const taskFile: TaskFile = await request.json();
    
    if (!taskFile.metadata.id || !taskFile.metadata.title) {
      return NextResponse.json({ error: 'Task ID and title are required' }, { status: 400 });
    }
    
    // Ensure repository field is set correctly
    if (!taskFile.metadata.repository) {
      taskFile.metadata.repository = repository;
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