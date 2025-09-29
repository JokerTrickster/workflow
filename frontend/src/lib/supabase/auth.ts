import { supabase } from './client'
import { AuthError, Session, User } from '@supabase/supabase-js'

// GitHub OAuth sign in
export async function signInWithGitHub(): Promise<{ error?: AuthError | null }> {
  try {
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'github',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`
      }
    })

    return { error }
  } catch (error) {
    console.error('GitHub sign in error:', error)
    return { error: error as AuthError }
  }
}

// Sign out
export async function signOut(): Promise<{ error?: AuthError | null }> {
  try {
    const { error } = await supabase.auth.signOut()
    return { error }
  } catch (error) {
    console.error('Sign out error:', error)
    return { error: error as AuthError }
  }
}

// Get current session
export async function getSession(): Promise<{
  session: Session | null
  error?: AuthError | null
}> {
  try {
    const { data: { session }, error } = await supabase.auth.getSession()
    return { session, error }
  } catch (error) {
    console.error('Get session error:', error)
    return { session: null, error: error as AuthError }
  }
}

// Get current user
export async function getUser(): Promise<{
  user: User | null
  error?: AuthError | null
}> {
  try {
    const { data: { user }, error } = await supabase.auth.getUser()
    return { user, error }
  } catch (error) {
    console.error('Get user error:', error)
    return { user: null, error: error as AuthError }
  }
}

// Subscribe to auth state changes
export function onAuthStateChange(
  callback: (event: string, session: Session | null) => void
) {
  const { data: { subscription } } = supabase.auth.onAuthStateChange(callback)
  
  return () => {
    subscription.unsubscribe()
  }
}

// Refresh session
export async function refreshSession(): Promise<{
  session: Session | null
  error?: AuthError | null
}> {
  try {
    const { data: { session }, error } = await supabase.auth.refreshSession()
    return { session, error }
  } catch (error) {
    console.error('Refresh session error:', error)
    return { session: null, error: error as AuthError }
  }
}

// Check if user is authenticated
export async function isAuthenticated(): Promise<boolean> {
  try {
    const { session } = await getSession()
    return !!session && !!session.user
  } catch (error) {
    console.error('Check authentication error:', error)
    return false
  }
}

// Get user profile (extended user data)
export async function getUserProfile(userId: string) {
  try {
    const { data, error } = await supabase
      .from('profiles')
      .select('*')
      .eq('id', userId)
      .single()

    return { data, error }
  } catch (error) {
    console.error('Get user profile error:', error)
    return { data: null, error }
  }
}

// Update user profile
export async function updateUserProfile(userId: string, updates: Record<string, string | number | boolean | null>) {
  try {
    const { data, error } = await supabase
      .from('profiles')
      .update(updates as never)
      .eq('id', userId)
      .select()
      .single()

    return { data, error }
  } catch (error) {
    console.error('Update user profile error:', error)
    return { data: null, error }
  }
}

// Sync GitHub repositories after login
export async function syncGitHubRepositories() {
  try {
    const { session } = await getSession()
    if (!session?.provider_token) {
      console.warn('No GitHub provider token found')
      return { error: 'No GitHub token available' }
    }

    // Get API base URL from config
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:7000/api/v1'

    const response = await fetch(`${apiBaseUrl}/github/sync-repositories`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${session.provider_token}`,
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(errorData.error || `HTTP ${response.status}`)
    }

    const result = await response.json()
    console.log('GitHub repositories synced successfully:', result)
    return { data: result }
  } catch (error) {
    console.error('GitHub repository sync error:', error)
    return { error: error instanceof Error ? error.message : 'Unknown error' }
  }
}