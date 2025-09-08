/**
 * @jest-environment jsdom
 */

// Removed unused createClient import
import { 
  signInWithGitHub, 
  signOut, 
  getSession, 
  getUser, 
  refreshSession, 
  isAuthenticated, 
  getUserProfile, 
  updateUserProfile 
} from '@/lib/supabase/auth'

// Mock Supabase client
jest.mock('@/lib/supabase/client', () => {
  const mockSupabase = {
    auth: {
      signInWithOAuth: jest.fn(),
      signOut: jest.fn(),
      getSession: jest.fn(),
      getUser: jest.fn(),
      refreshSession: jest.fn(),
      onAuthStateChange: jest.fn(() => ({
        data: { subscription: { unsubscribe: jest.fn() } }
      }))
    },
    from: jest.fn(() => ({
      select: jest.fn().mockReturnThis(),
      eq: jest.fn().mockReturnThis(),
      single: jest.fn(),
      update: jest.fn().mockReturnThis()
    }))
  }
  return { supabase: mockSupabase }
})

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    origin: 'http://localhost:3000'
  },
  writable: true
})

import { supabase } from '@/lib/supabase/client'
const mockSupabase = supabase as jest.Mocked<typeof supabase>

describe('Auth Functions', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Reset console.error mock
    jest.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  describe('signInWithGitHub', () => {
    it('should successfully initiate GitHub OAuth sign in', async () => {
      const mockResponse = { error: null }
      mockSupabase.auth.signInWithOAuth.mockResolvedValue(mockResponse)

      const result = await signInWithGitHub()

      expect(mockSupabase.auth.signInWithOAuth).toHaveBeenCalledWith({
        provider: 'github',
        options: {
          redirectTo: 'http://localhost:3000/auth/callback'
        }
      })
      expect(result).toEqual({ error: null })
    })

    it('should handle GitHub OAuth sign in error', async () => {
      const mockError = { message: 'OAuth failed', code: 'oauth_error' }
      mockSupabase.auth.signInWithOAuth.mockResolvedValue({ error: mockError })

      const result = await signInWithGitHub()

      expect(result).toEqual({ error: mockError })
    })

    it('should handle unexpected errors during GitHub sign in', async () => {
      const unexpectedError = new Error('Network error')
      mockSupabase.auth.signInWithOAuth.mockRejectedValue(unexpectedError)

      const result = await signInWithGitHub()

      expect(console.error).toHaveBeenCalledWith('GitHub sign in error:', unexpectedError)
      expect(result).toEqual({ error: unexpectedError })
    })
  })

  describe('signOut', () => {
    it('should successfully sign out user', async () => {
      const mockResponse = { error: null }
      mockSupabase.auth.signOut.mockResolvedValue(mockResponse)

      const result = await signOut()

      expect(mockSupabase.auth.signOut).toHaveBeenCalled()
      expect(result).toEqual({ error: null })
    })

    it('should handle sign out error', async () => {
      const mockError = { message: 'Sign out failed', code: 'signout_error' }
      mockSupabase.auth.signOut.mockResolvedValue({ error: mockError })

      const result = await signOut()

      expect(result).toEqual({ error: mockError })
    })

    it('should handle unexpected errors during sign out', async () => {
      const unexpectedError = new Error('Network error')
      mockSupabase.auth.signOut.mockRejectedValue(unexpectedError)

      const result = await signOut()

      expect(console.error).toHaveBeenCalledWith('Sign out error:', unexpectedError)
      expect(result).toEqual({ error: unexpectedError })
    })
  })

  describe('getSession', () => {
    it('should successfully get current session', async () => {
      const mockSession = {
        access_token: 'token',
        refresh_token: 'refresh',
        user: { id: 'user-id', email: 'test@example.com' }
      }
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: mockSession },
        error: null
      })

      const result = await getSession()

      expect(mockSupabase.auth.getSession).toHaveBeenCalled()
      expect(result).toEqual({ session: mockSession, error: null })
    })

    it('should handle null session', async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: null },
        error: null
      })

      const result = await getSession()

      expect(result).toEqual({ session: null, error: null })
    })

    it('should handle session retrieval error', async () => {
      const mockError = { message: 'Session retrieval failed', code: 'session_error' }
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: null },
        error: mockError
      })

      const result = await getSession()

      expect(result).toEqual({ session: null, error: mockError })
    })

    it('should handle unexpected errors during session retrieval', async () => {
      const unexpectedError = new Error('Network error')
      mockSupabase.auth.getSession.mockRejectedValue(unexpectedError)

      const result = await getSession()

      expect(console.error).toHaveBeenCalledWith('Get session error:', unexpectedError)
      expect(result).toEqual({ session: null, error: unexpectedError })
    })
  })

  describe('getUser', () => {
    it('should successfully get current user', async () => {
      const mockUser = { id: 'user-id', email: 'test@example.com' }
      mockSupabase.auth.getUser.mockResolvedValue({
        data: { user: mockUser },
        error: null
      })

      const result = await getUser()

      expect(mockSupabase.auth.getUser).toHaveBeenCalled()
      expect(result).toEqual({ user: mockUser, error: null })
    })

    it('should handle null user', async () => {
      mockSupabase.auth.getUser.mockResolvedValue({
        data: { user: null },
        error: null
      })

      const result = await getUser()

      expect(result).toEqual({ user: null, error: null })
    })

    it('should handle user retrieval error', async () => {
      const mockError = { message: 'User retrieval failed', code: 'user_error' }
      mockSupabase.auth.getUser.mockResolvedValue({
        data: { user: null },
        error: mockError
      })

      const result = await getUser()

      expect(result).toEqual({ user: null, error: mockError })
    })
  })

  describe('refreshSession', () => {
    it('should successfully refresh session', async () => {
      const mockSession = {
        access_token: 'new-token',
        refresh_token: 'new-refresh',
        user: { id: 'user-id', email: 'test@example.com' }
      }
      mockSupabase.auth.refreshSession.mockResolvedValue({
        data: { session: mockSession },
        error: null
      })

      const result = await refreshSession()

      expect(mockSupabase.auth.refreshSession).toHaveBeenCalled()
      expect(result).toEqual({ session: mockSession, error: null })
    })

    it('should handle refresh session error', async () => {
      const mockError = { message: 'Token refresh failed', code: 'refresh_error' }
      mockSupabase.auth.refreshSession.mockResolvedValue({
        data: { session: null },
        error: mockError
      })

      const result = await refreshSession()

      expect(result).toEqual({ session: null, error: mockError })
    })
  })

  describe('isAuthenticated', () => {
    it('should return true when user is authenticated', async () => {
      const mockSession = {
        access_token: 'token',
        user: { id: 'user-id', email: 'test@example.com' }
      }
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: mockSession },
        error: null
      })

      const result = await isAuthenticated()

      expect(result).toBe(true)
    })

    it('should return false when no session exists', async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: null },
        error: null
      })

      const result = await isAuthenticated()

      expect(result).toBe(false)
    })

    it('should return false when session has no user', async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: { access_token: 'token', user: null } },
        error: null
      })

      const result = await isAuthenticated()

      expect(result).toBe(false)
    })

    it('should return false on authentication check error', async () => {
      const unexpectedError = new Error('Network error')
      mockSupabase.auth.getSession.mockRejectedValue(unexpectedError)

      const result = await isAuthenticated()

      expect(console.error).toHaveBeenCalledWith('Check authentication error:', unexpectedError)
      expect(result).toBe(false)
    })
  })

  describe('getUserProfile', () => {
    it('should successfully get user profile', async () => {
      const mockProfile = {
        id: 'user-id',
        email: 'test@example.com',
        full_name: 'Test User'
      }
      
      const mockQuery = mockSupabase.from()
      mockQuery.single.mockResolvedValue({ data: mockProfile, error: null })

      const result = await getUserProfile('user-id')

      expect(mockSupabase.from).toHaveBeenCalledWith('profiles')
      expect(mockQuery.select).toHaveBeenCalledWith('*')
      expect(mockQuery.eq).toHaveBeenCalledWith('id', 'user-id')
      expect(mockQuery.single).toHaveBeenCalled()
      expect(result).toEqual({ data: mockProfile, error: null })
    })

    it('should handle profile retrieval error', async () => {
      const mockError = { message: 'Profile not found', code: 'profile_error' }
      
      const mockQuery = mockSupabase.from()
      mockQuery.single.mockResolvedValue({ data: null, error: mockError })

      const result = await getUserProfile('user-id')

      expect(result).toEqual({ data: null, error: mockError })
    })
  })

  describe('updateUserProfile', () => {
    it('should successfully update user profile', async () => {
      const mockUpdatedProfile = {
        id: 'user-id',
        email: 'test@example.com',
        full_name: 'Updated Name'
      }
      const updates = { full_name: 'Updated Name' }
      
      const mockQuery = mockSupabase.from()
      mockQuery.single.mockResolvedValue({ data: mockUpdatedProfile, error: null })

      const result = await updateUserProfile('user-id', updates)

      expect(mockSupabase.from).toHaveBeenCalledWith('profiles')
      expect(mockQuery.update).toHaveBeenCalledWith(updates)
      expect(mockQuery.eq).toHaveBeenCalledWith('id', 'user-id')
      expect(mockQuery.select).toHaveBeenCalled()
      expect(mockQuery.single).toHaveBeenCalled()
      expect(result).toEqual({ data: mockUpdatedProfile, error: null })
    })

    it('should handle profile update error', async () => {
      const mockError = { message: 'Update failed', code: 'update_error' }
      const updates = { full_name: 'Updated Name' }
      
      const mockQuery = mockSupabase.from()
      mockQuery.single.mockResolvedValue({ data: null, error: mockError })

      const result = await updateUserProfile('user-id', updates)

      expect(result).toEqual({ data: null, error: mockError })
    })
  })
})