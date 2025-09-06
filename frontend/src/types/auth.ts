import { User, Session } from '@supabase/supabase-js'

export interface AuthUser extends User {
  // Extended user properties
  email_confirmed_at?: string
  phone_confirmed_at?: string
  last_sign_in_at?: string
  updated_at?: string
}

export interface AuthSession extends Session {
  // Extended session properties
  user: User
}

export interface UserProfile {
  id: string
  email: string
  full_name?: string
  avatar_url?: string
  username?: string
  github_username?: string
  bio?: string
  website?: string
  location?: string
  created_at: string
  updated_at: string
}

export interface AuthState {
  user: AuthUser | null
  session: AuthSession | null
  loading: boolean
  error: string | null
}

export interface SignInCredentials {
  email?: string
  password?: string
  provider?: 'github' | 'google'
}

export interface SignUpCredentials {
  email: string
  password: string
  full_name?: string
}

export interface AuthResponse {
  user?: AuthUser | null
  session?: AuthSession | null
  error?: AuthErrorType | { message: string } | null
}

export interface AuthContextType {
  user: AuthUser | null
  session: AuthSession | null
  loading: boolean
  error: string | null
  signInWithGitHub: () => Promise<AuthResponse>
  signOut: () => Promise<{ error?: AuthErrorType | { message: string } | null }>
  refreshSession: () => Promise<AuthResponse>
  isAuthenticated: boolean
}

export interface ProviderSettings {
  github: {
    enabled: boolean
    client_id?: string
  }
  google: {
    enabled: boolean
    client_id?: string
  }
}

export interface AuthConfig {
  providers: ProviderSettings
  redirectTo?: string
  autoRefreshToken?: boolean
  persistSession?: boolean
  detectSessionInUrl?: boolean
}

// Auth event types
export type AuthEventType = 
  | 'SIGNED_IN' 
  | 'SIGNED_OUT' 
  | 'TOKEN_REFRESHED' 
  | 'USER_UPDATED' 
  | 'PASSWORD_RECOVERY'

export interface AuthEvent {
  event: AuthEventType
  session: AuthSession | null
  user: AuthUser | null
}

// Auth errors
export interface AuthErrorType {
  code: string
  message: string
  details?: Record<string, unknown>
}

export const AUTH_ERRORS = {
  INVALID_CREDENTIALS: 'invalid_credentials',
  USER_NOT_FOUND: 'user_not_found',
  SIGNUP_DISABLED: 'signup_disabled',
  RATE_LIMIT_EXCEEDED: 'rate_limit_exceeded',
  INVALID_EMAIL: 'invalid_email',
  WEAK_PASSWORD: 'weak_password',
  EMAIL_NOT_CONFIRMED: 'email_not_confirmed',
  SESSION_EXPIRED: 'session_expired',
  NETWORK_ERROR: 'network_error',
  UNKNOWN_ERROR: 'unknown_error'
} as const

export type AuthErrorCode = typeof AUTH_ERRORS[keyof typeof AUTH_ERRORS]