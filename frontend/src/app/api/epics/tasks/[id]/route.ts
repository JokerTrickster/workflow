import { NextRequest, NextResponse } from 'next/server';
import { getApiBaseUrl } from '../../../../../config/api';

// Force dynamic rendering - no static generation
export const dynamic = 'force-dynamic';
export const revalidate = 0;

// Get specific task - proxy to backend
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {
    const response = await fetch(`${getApiBaseUrl()}/tasks/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      return NextResponse.json(
        { error: 'Task not found in backend' },
        { status: response.status }
      );
    }

    const taskData = await response.json();
    return NextResponse.json(taskData);
  } catch (error) {
    console.error(`Failed to get task ${id}:`, error);
    return NextResponse.json({ error: 'Failed to get task' }, { status: 500 });
  }
}

// Update task - proxy to backend
export async function PUT(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {
    const requestBody = await request.json();

    const response = await fetch(`${getApiBaseUrl()}/tasks/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const errorData = await response.text();
      return NextResponse.json(
        { error: 'Failed to update task in backend', details: errorData },
        { status: response.status }
      );
    }

    const result = await response.json();
    return NextResponse.json(result);
  } catch (error) {
    console.error(`Failed to update task ${id}:`, error);
    return NextResponse.json({ error: 'Failed to update task' }, { status: 500 });
  }
}

// Execute task
export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {

    // Call the backend API to execute the task
    const backendUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:7000/api/v1';
    const executeUrl = `${backendUrl}/tasks/${id}/execute`;

    const response = await fetch(executeUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const errorData = await response.text();
      console.error('Backend execute API failed:', errorData);
      return NextResponse.json(
        { error: 'Failed to execute task on backend', details: errorData },
        { status: response.status }
      );
    }

    const result = await response.json();
    return NextResponse.json(result);
  } catch (error) {
    console.error(`Failed to execute task ${id}:`, error);
    return NextResponse.json({ error: 'Failed to execute task' }, { status: 500 });
  }
}

// Delete task - proxy to backend
export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {
    const response = await fetch(`${getApiBaseUrl()}/tasks/${id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const errorData = await response.text();
      return NextResponse.json(
        { error: 'Failed to delete task in backend', details: errorData },
        { status: response.status }
      );
    }

    const result = await response.json();
    return NextResponse.json(result);
  } catch (error) {
    console.error(`Failed to delete task ${id}:`, error);
    return NextResponse.json({ error: 'Failed to delete task' }, { status: 500 });
  }
}