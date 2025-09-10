import { NextRequest, NextResponse } from 'next/server';
import fs from 'fs/promises';
import path from 'path';
import matter from 'gray-matter';
import { TaskFile, TaskFileMetadata } from '../route';

const EPICS_BASE_DIR = path.resolve(process.cwd(), '../.claude/epics');

// Get current repository name
async function getCurrentRepository(): Promise<string> {
  return process.env.REPOSITORY_NAME || 'workflow';
}

// Get tasks directory for specific repository
function getTasksDir(repository: string): string {
  return path.join(EPICS_BASE_DIR, 'repositories', repository, 'tasks');
}

// Get specific task file
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const repository = await getCurrentRepository();
    const TASKS_DIR = getTasksDir(repository);
    const { id } = params;
    const filename = `${id}.md`;
    const filePath = path.join(TASKS_DIR, filename);
    
    const content = await fs.readFile(filePath, 'utf-8');
    const { data: frontMatter, content: markdownContent } = matter(content);
    
    const taskFile: TaskFile = {
      metadata: frontMatter as TaskFileMetadata,
      content: markdownContent.trim(),
    };
    
    // Ensure repository field is set
    if (!taskFile.metadata.repository) {
      taskFile.metadata.repository = repository;
    }
    
    return NextResponse.json(taskFile);
  } catch (error) {
    console.error(`Failed to get task ${params.id}:`, error);
    
    if ((error as any).code === 'ENOENT') {
      return NextResponse.json({ error: 'Task not found' }, { status: 404 });
    }
    
    return NextResponse.json({ error: 'Failed to get task' }, { status: 500 });
  }
}

// Update task file
export async function PUT(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const repository = await getCurrentRepository();
    const TASKS_DIR = getTasksDir(repository);
    const { id } = params;
    const filename = `${id}.md`;
    const filePath = path.join(TASKS_DIR, filename);
    
    const updatedTaskFile: TaskFile = await request.json();
    
    // Ensure the ID matches
    if (updatedTaskFile.metadata.id !== id) {
      return NextResponse.json({ error: 'Task ID mismatch' }, { status: 400 });
    }
    
    // Ensure repository field is set correctly
    if (!updatedTaskFile.metadata.repository) {
      updatedTaskFile.metadata.repository = repository;
    }
    
    // Generate markdown content with frontmatter
    const frontMatter = Object.entries(updatedTaskFile.metadata)
      .map(([key, value]) => {
        if (value === undefined || value === null) return null;
        if (typeof value === 'string' && value.includes('\n')) {
          return `${key}: |\n  ${value.replace(/\n/g, '\n  ')}`;
        }
        return `${key}: ${JSON.stringify(value)}`;
      })
      .filter(Boolean)
      .join('\n');
    
    const fileContent = `---\n${frontMatter}\n---\n\n${updatedTaskFile.content}`;
    
    await fs.writeFile(filePath, fileContent, 'utf-8');
    
    return NextResponse.json(updatedTaskFile);
  } catch (error) {
    console.error(`Failed to update task ${params.id}:`, error);
    
    if ((error as any).code === 'ENOENT') {
      return NextResponse.json({ error: 'Task not found' }, { status: 404 });
    }
    
    return NextResponse.json({ error: 'Failed to update task' }, { status: 500 });
  }
}

// Delete task file
export async function DELETE(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const repository = await getCurrentRepository();
    const TASKS_DIR = getTasksDir(repository);
    const { id } = params;
    const filename = `${id}.md`;
    const filePath = path.join(TASKS_DIR, filename);
    
    await fs.unlink(filePath);
    
    return NextResponse.json({ success: true });
  } catch (error) {
    console.error(`Failed to delete task ${params.id}:`, error);
    
    if ((error as any).code === 'ENOENT') {
      return NextResponse.json({ error: 'Task not found' }, { status: 404 });
    }
    
    return NextResponse.json({ error: 'Failed to delete task' }, { status: 500 });
  }
}