import { NextRequest, NextResponse } from 'next/server';

interface MergeParams {
  params: {
    owner: string;
    repo: string;
    number: string;
  };
}

export async function PUT(request: NextRequest, { params }: MergeParams) {
  try {
    const { owner, repo, number } = params;
    const body = await request.json();
    
    // Get GitHub token from environment
    const githubToken = process.env.GITHUB_TOKEN;
    if (!githubToken) {
      return NextResponse.json(
        { error: 'GitHub token not configured' },
        { status: 500 }
      );
    }

    // Call GitHub API to merge PR
    const response = await fetch(
      `https://api.github.com/repos/${owner}/${repo}/pulls/${number}/merge`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${githubToken}`,
          'Accept': 'application/vnd.github.v3+json',
          'Content-Type': 'application/json',
          'User-Agent': 'workflow-app',
        },
        body: JSON.stringify({
          commit_title: body.commit_title,
          commit_message: body.commit_message,
          merge_method: body.merge_method || 'merge',
        }),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      return NextResponse.json(
        { error: error.message || 'Failed to merge PR' },
        { status: response.status }
      );
    }

    const result = await response.json();
    
    return NextResponse.json(result);
    
  } catch (error) {
    console.error('PR merge error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}