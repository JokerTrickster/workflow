import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:7000/api/v1';

export async function POST(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const { id } = params;

    const response = await fetch(`${BACKEND_URL}/tasks/${id}/execute`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error('Task execute API proxy error:', error);
    return NextResponse.json(
      { error: 'Failed to execute task' },
      { status: 500 }
    );
  }
}