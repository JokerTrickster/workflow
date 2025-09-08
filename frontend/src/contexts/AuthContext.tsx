'use client'

import { createContext, useContext, useEffect, useState, useCallback } from 'react'
import type { ReactNode } from 'react'
// Removed unused User and Session imports - using types from auth.ts instead
import { supabase } from '@/lib/supabase/client'
import { AuthContextType, AuthUser, AuthSession } from '@/types/auth'

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [session, setSession] = useState<AuthSession | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Initialize auth state
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        setLoading(true)
        
        console.log('Initializing auth...')
        console.log('Supabase URL:', process.env.NEXT_PUBLIC_SUPABASE_URL)
        console.log('Supabase Anon Key exists:', !!process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY)
        
        // Get initial session
        const { data: { session }, error } = await supabase.auth.getSession()
        
        if (error) {
          console.error('Supabase auth error:', error)
          console.error('Error details:', JSON.stringify(error, null, 2))
          setError(error.message)
        } else {
          console.log('Auth session:', session)
          setSession(session as AuthSession | null)
          setUser(session?.user as AuthUser | null)
        }
      } catch (error) {
        console.error('Auth initialization catch error:', error)
        console.error('Error stack:', (error as Error)?.stack)
        setError('Failed to initialize authentication')
      } finally {
        setLoading(false)
      }
    }

    initializeAuth()
  }, [])

  // Listen for auth state changes
  useEffect(() => {
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      async (event, session) => {
        console.log('Auth state changed:', event, session)
        
        setSession(session as AuthSession | null)
        setUser(session?.user as AuthUser | null)
        setLoading(false)
        setError(null)

        // Handle specific auth events
        switch (event) {
          case 'SIGNED_IN':
            console.log('User signed in:', session?.user?.email)
            break
          case 'SIGNED_OUT':
            console.log('User signed out')
            break
          case 'TOKEN_REFRESHED':
            console.log('Token refreshed')
            break
          case 'USER_UPDATED':
            console.log('User updated')
            break
        }
      }
    )

    return () => subscription.unsubscribe()
  }, [])

  // Sign in with GitHub
  const signInWithGitHub = useCallback(async () => {
    try {
      setError(null)
      const { error } = await supabase.auth.signInWithOAuth({
        provider: 'github',
        options: {
          redirectTo: `${window.location.origin}/auth/callback`
        }
      })

      if (error) {
        setError(error.message)
        return { error }
      }

      return {}
    } catch (error: unknown) {
      const errorMessage = (error as Error)?.message || 'Failed to sign in with GitHub'
      setError(errorMessage)
      return { error: { message: errorMessage } }
    }
  }, [])

  // Sign out
  const signOut = useCallback(async () => {
    try {
      setError(null)
      const { error } = await supabase.auth.signOut()

      if (error) {
        setError(error.message)
        return { error }
      }

      setUser(null)
      setSession(null)
      return {}
    } catch (error: unknown) {
      const errorMessage = (error as Error)?.message || 'Failed to sign out'
      setError(errorMessage)
      return { error: { message: errorMessage } }
    }
  }, [])

  // Refresh session
  const refreshSession = useCallback(async () => {
    try {
      setError(null)
      const { data: { session }, error } = await supabase.auth.refreshSession()

      if (error) {
        setError(error.message)
        return { error }
      }

      setSession(session as AuthSession | null)
      setUser(session?.user as AuthUser | null)
      return { session: session as AuthSession | null, user: session?.user as AuthUser | null }
    } catch (error: unknown) {
      const errorMessage = (error as Error)?.message || 'Failed to refresh session'
      setError(errorMessage)
      return { error: { message: errorMessage } }
    }
  }, [])

  const value: AuthContextType = {
    user,
    session,
    loading,
    error,
    signInWithGitHub,
    signOut,
    refreshSession,
    isAuthenticated: !!user && !!session
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}