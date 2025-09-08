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