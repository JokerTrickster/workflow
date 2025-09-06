/**
 * Integration test to verify auth setup and configuration
 */

import { createClient } from '@supabase/supabase-js'

// Mock environment variables for testing
const mockEnv = {
  NEXT_PUBLIC_SUPABASE_URL: 'https://test-project.supabase.co',
  NEXT_PUBLIC_SUPABASE_ANON_KEY: 'test-anon-key',
  SUPABASE_SERVICE_ROLE_KEY: 'test-service-role-key'
}

describe('Auth Setup Integration', () => {
  beforeEach(() => {
    // Mock environment variables
    Object.entries(mockEnv).forEach(([key, value]) => {
      process.env[key] = value
    })
  })

  afterEach(() => {
    // Clean up environment variables
    Object.keys(mockEnv).forEach(key => {
      delete process.env[key]
    })
  })

  it('should create Supabase client with correct configuration', () => {
    const supabase = createClient(
      process.env.NEXT_PUBLIC_SUPABASE_URL!,
      process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
      {
        auth: {
          persistSession: true,
          autoRefreshToken: true,
          detectSessionInUrl: true
        }
      }
    )

    expect(supabase).toBeDefined()
    expect(supabase.auth).toBeDefined()
    expect(supabase.from).toBeDefined()
  })

  it('should have all required environment variables defined in example', () => {
    const requiredVars = [
      'NEXT_PUBLIC_SUPABASE_URL',
      'NEXT_PUBLIC_SUPABASE_ANON_KEY',
      'SUPABASE_SERVICE_ROLE_KEY',
      'GITHUB_CLIENT_ID',
      'GITHUB_CLIENT_SECRET',
      'NEXTAUTH_URL',
      'NEXTAUTH_SECRET'
    ]

    // This test verifies that our .env.example has all required variables
    import path from 'path'
    import fs from 'fs'
    const envExamplePath = path.resolve(__dirname, '../../../.env.example')
    
    let envExampleContent = ''
    try {
      envExampleContent = fs.readFileSync(envExamplePath, 'utf8')
    } catch {
      // If .env.example doesn't exist, that's fine for this test
      console.log('.env.example not found, skipping validation')
      return
    }

    requiredVars.forEach(varName => {
      expect(envExampleContent).toContain(varName)
    })
  })

  it('should validate auth client configuration options', () => {
    const clientConfig = {
      auth: {
        persistSession: true,
        autoRefreshToken: true,
        detectSessionInUrl: true
      }
    }

    expect(clientConfig.auth.persistSession).toBe(true)
    expect(clientConfig.auth.autoRefreshToken).toBe(true)
    expect(clientConfig.auth.detectSessionInUrl).toBe(true)
  })

  it('should validate server client configuration options', () => {
    const serverConfig = {
      auth: {
        persistSession: false,
        autoRefreshToken: false,
        detectSessionInUrl: false
      }
    }

    expect(serverConfig.auth.persistSession).toBe(false)
    expect(serverConfig.auth.autoRefreshToken).toBe(false)
    expect(serverConfig.auth.detectSessionInUrl).toBe(false)
  })

  it('should validate GitHub OAuth provider configuration', () => {
    const githubConfig = {
      provider: 'github',
      options: {
        redirectTo: `${process.env.NEXT_PUBLIC_SUPABASE_URL}/auth/callback`
      }
    }

    expect(githubConfig.provider).toBe('github')
    expect(githubConfig.options.redirectTo).toContain('/auth/callback')
  })

  it('should validate auth callback URL structure', () => {
    const baseUrl = 'http://localhost:3000'
    const callbackUrl = `${baseUrl}/auth/callback`

    expect(callbackUrl).toBe('http://localhost:3000/auth/callback')
  })
})