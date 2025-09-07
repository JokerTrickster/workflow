import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { createSupabaseMiddleware } from '@/utils/supabase/middleware'

export async function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname

  // Skip middleware for static files and API routes
  if (
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api') ||
    pathname.includes('.') // Static files
  ) {
    return NextResponse.next()
  }

  const { supabase, response } = createSupabaseMiddleware(request)
  
  // Skip middleware for auth callback
  if (pathname.startsWith('/auth/callback')) {
    console.log('[Middleware] Skipping auth callback')
    return NextResponse.next()
  }
  
  console.log('[Middleware]', pathname, 'Checking auth...')

  // Simple auth check - just check if we have github_token cookie
  const githubToken = request.cookies.get('github_token')
  const isAuthenticated = !!githubToken
  
  console.log('[Middleware]', pathname, { isAuthenticated, hasGithubToken: !!githubToken })

  // Protected routes that require authentication
  const protectedRoutes = ['/dashboard']
  const isProtectedRoute = protectedRoutes.some(route => pathname.startsWith(route))
  
  if (isProtectedRoute && !isAuthenticated) {
    console.log('[Middleware] No github token, redirecting to login')
    const loginUrl = new URL('/login', request.url)
    loginUrl.searchParams.set('redirect', pathname)
    return NextResponse.redirect(loginUrl)
  }

  // Redirect authenticated users away from auth pages
  const authRoutes = ['/login']
  const isAuthRoute = authRoutes.some(route => pathname === route)
  
  if (isAuthRoute && isAuthenticated) {
    console.log('[Middleware] Already authenticated, redirecting to dashboard')
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  // Redirect root to dashboard if authenticated, otherwise to login
  if (pathname === '/') {
    if (isAuthenticated) {
      return NextResponse.redirect(new URL('/dashboard', request.url))
    } else {
      return NextResponse.redirect(new URL('/login', request.url))
    }
  }

  return response
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder files
     */
    '/((?!_next/static|_next/image|favicon.ico|manifest.json|icons/).*)',
  ],
}