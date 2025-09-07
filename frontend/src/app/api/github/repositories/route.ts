import { NextRequest, NextResponse } from 'next/server'
import { createServerClient } from '@supabase/ssr'
import { cookies } from 'next/headers'
import { createGitHubService, extractGitHubToken, GitHubAPIError } from '@/services/github'
import { RepositoriesQuery } from '@/types/github'

export async function GET(request: NextRequest) {
  try {
    // Get cookies store
    const cookieStore = await cookies()
    
    // Create Supabase client to verify authentication
    const supabase = createServerClient(
      process.env.NEXT_PUBLIC_SUPABASE_URL!,
      process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
      {
        cookies: {
          getAll() {
            return cookieStore.getAll()
          },
          setAll(cookiesToSet) {
            try {
              cookiesToSet.forEach(({ name, value, options }) =>
                cookieStore.set(name, value, options)
              )
            } catch {
              // This can be ignored if you have middleware refreshing sessions
            }
          },
        },
      }
    )

    // Get current user (not session)
    const { data: { user }, error: authError } = await supabase.auth.getUser()
    
    if (authError || !user) {
      return NextResponse.json(
        { error: 'Authentication required' },
        { status: 401 }
      )
    }

    // Get session for the GitHub token
    const { data: { session } } = await supabase.auth.getSession()
    if (!session) {
      return NextResponse.json(
        { error: 'No active session' },
        { status: 401 }
      )
    }

    // Extract GitHub token from session
    const githubToken = extractGitHubToken(session)
    
    if (!githubToken) {
      return NextResponse.json(
        { error: 'GitHub token not found in session' },
        { status: 401 }
      )
    }

    // Parse query parameters
    const { searchParams } = new URL(request.url)
    const query: RepositoriesQuery = {
      page: searchParams.get('page') ? parseInt(searchParams.get('page')!, 10) : 1,
      per_page: searchParams.get('per_page') ? parseInt(searchParams.get('per_page')!, 10) : 30,
      sort: (searchParams.get('sort') as RepositoriesQuery['sort']) || 'updated',
      direction: (searchParams.get('direction') as RepositoriesQuery['direction']) || 'desc',
      type: (searchParams.get('type') as RepositoriesQuery['type']) || 'all',
    }

    // Validate query parameters
    if (query.page && (query.page < 1 || query.page > 1000)) {
      return NextResponse.json(
        { error: 'Page must be between 1 and 1000' },
        { status: 400 }
      )
    }

    if (query.per_page && (query.per_page < 1 || query.per_page > 100)) {
      return NextResponse.json(
        { error: 'Per page must be between 1 and 100' },
        { status: 400 }
      )
    }

    // Create GitHub service and fetch repositories
    const githubService = createGitHubService(githubToken)
    const repositoriesResponse = await githubService.getRepositories(query)

    // Return successful response with repositories
    return NextResponse.json(repositoriesResponse)

  } catch (error) {
    console.error('GitHub API error:', error)

    // Handle GitHub API specific errors
    if (error instanceof GitHubAPIError) {
      return NextResponse.json(
        { 
          error: error.message,
          details: error.response 
        },
        { status: error.status }
      )
    }

    // Handle other errors
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}

// Handle POST request for bulk operations (if needed)
export async function POST(request: NextRequest) {
  try {
    // Get cookies store
    const cookieStore = await cookies()
    
    // Create Supabase client to verify authentication
    const supabase = createServerClient(
      process.env.NEXT_PUBLIC_SUPABASE_URL!,
      process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
      {
        cookies: {
          getAll() {
            return cookieStore.getAll()
          },
          setAll(cookiesToSet) {
            try {
              cookiesToSet.forEach(({ name, value, options }) =>
                cookieStore.set(name, value, options)
              )
            } catch {
              // This can be ignored if you have middleware refreshing sessions
            }
          },
        },
      }
    )

    // Get current user (not session)
    const { data: { user }, error: authError } = await supabase.auth.getUser()
    
    if (authError || !user) {
      return NextResponse.json(
        { error: 'Authentication required' },
        { status: 401 }
      )
    }

    // Get session for the GitHub token
    const { data: { session } } = await supabase.auth.getSession()
    if (!session) {
      return NextResponse.json(
        { error: 'No active session' },
        { status: 401 }
      )
    }

    // Extract GitHub token from session
    const githubToken = extractGitHubToken(session)
    
    if (!githubToken) {
      return NextResponse.json(
        { error: 'GitHub token not found in session' },
        { status: 401 }
      )
    }

    // Parse request body
    const body = await request.json()
    const { action, maxPages = 10 } = body

    // Create GitHub service
    const githubService = createGitHubService(githubToken)

    switch (action) {
      case 'fetchAll': {
        // Fetch all repositories with pagination
        const allRepositories = await githubService.getAllRepositories(
          {
            sort: body.sort || 'updated',
            direction: body.direction || 'desc',
            type: body.type || 'all'
          },
          maxPages
        )

        return NextResponse.json({
          repositories: allRepositories,
          totalCount: allRepositories.length,
          message: `Fetched ${allRepositories.length} repositories`
        })
      }

      case 'search': {
        const { query: searchQuery, ...searchOptions } = body
        
        if (!searchQuery) {
          return NextResponse.json(
            { error: 'Search query is required' },
            { status: 400 }
          )
        }

        const searchResults = await githubService.searchRepositories(searchQuery, searchOptions)
        return NextResponse.json(searchResults)
      }

      default:
        return NextResponse.json(
          { error: 'Invalid action. Supported actions: fetchAll, search' },
          { status: 400 }
        )
    }

  } catch (error) {
    console.error('GitHub API POST error:', error)

    // Handle GitHub API specific errors
    if (error instanceof GitHubAPIError) {
      return NextResponse.json(
        { 
          error: error.message,
          details: error.response 
        },
        { status: error.status }
      )
    }

    // Handle other errors
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    )
  }
}