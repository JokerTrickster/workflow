'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase/client'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { CheckCircle, AlertCircle, ArrowRight } from 'lucide-react'

type CallbackState = 'loading' | 'success' | 'error'

interface CallbackError {
  message: string
  details?: string
}

export default function AuthCallbackPage() {
  const router = useRouter()
  const [state, setState] = useState<CallbackState>('loading')
  const [error, setError] = useState<CallbackError | null>(null)
  const [redirectTarget, setRedirectTarget] = useState<string>('/dashboard')
  const [countdown, setCountdown] = useState<number>(3)

  useEffect(() => {
    const handleAuthCallback = async () => {
      try {
        setState('loading')
        setError(null)

        console.log('Processing auth callback...')
        console.log('Current URL:', window.location.href)
        console.log('Hash:', window.location.hash)
        console.log('Search:', window.location.search)

        // Check if we have tokens in the URL fragment (implicit flow)
        const hashParams = new URLSearchParams(window.location.hash.substring(1))
        const searchParams = new URLSearchParams(window.location.search)
        
        const accessToken = hashParams.get('access_token')
        const refreshToken = hashParams.get('refresh_token')
        const expiresIn = hashParams.get('expires_in')
        const tokenType = hashParams.get('token_type')
        const code = searchParams.get('code')
        
        console.log('Auth tokens found:', {
          hasAccessToken: !!accessToken,
          hasRefreshToken: !!refreshToken,
          hasCode: !!code,
          expiresIn,
          tokenType
        })

        // Handle implicit flow (tokens in fragment)
        if (accessToken && refreshToken) {
          console.log('Processing implicit flow tokens...')
          
          const { data, error: sessionError } = await supabase.auth.setSession({
            access_token: accessToken,
            refresh_token: refreshToken
          })

          if (sessionError) {
            console.error('Session creation error:', sessionError)
            setError({
              message: 'Failed to create session with tokens.',
              details: sessionError.message
            })
            setState('error')
            return
          }

          if (data.session) {
            console.log('✅ Got session! Redirecting to dashboard...', data.session.user.email)
            
            // Store GitHub access token in cookie for API calls
            const providerToken = data.session.provider_token
            console.log('🔵 Provider token:', providerToken ? 'EXISTS' : 'MISSING')
            
            if (providerToken) {
              const cookieString = `github_token=${providerToken}; path=/; max-age=${60 * 60 * 24 * 7}; secure; samesite=lax`
              document.cookie = cookieString
              console.log('🍪 Cookie set:', cookieString.substring(0, 50) + '...')
            } else {
              console.log('❌ No provider token to store')
            }
            
            // Immediate redirect - no countdown
            window.location.href = '/dashboard'
            return
          }
        }

        // Handle authorization code flow
        if (code) {
          console.log('Processing authorization code...')
          
          const { data, error: exchangeError } = await supabase.auth.exchangeCodeForSession(code)

          if (exchangeError) {
            console.error('Code exchange error:', exchangeError)
            setError({
              message: 'Failed to exchange authorization code.',
              details: exchangeError.message
            })
            setState('error')
            return
          }

          if (data.session) {
            console.log('✅ Got session! Redirecting to dashboard...', data.session.user.email)
            
            // Store GitHub access token in cookie for API calls
            const providerToken = data.session.provider_token
            console.log('🔵 Provider token:', providerToken ? 'EXISTS' : 'MISSING')
            
            if (providerToken) {
              const cookieString = `github_token=${providerToken}; path=/; max-age=${60 * 60 * 24 * 7}; secure; samesite=lax`
              document.cookie = cookieString
              console.log('🍪 Cookie set:', cookieString.substring(0, 50) + '...')
            } else {
              console.log('❌ No provider token to store')
            }
            
            // Immediate redirect - no countdown
            window.location.href = '/dashboard'
            return
          }
        }

        // Handle getSession for any existing session
        const { data: sessionData, error: getSessionError } = await supabase.auth.getSession()
        
        if (getSessionError) {
          console.error('Get session error:', getSessionError)
          setError({
            message: 'Failed to retrieve session.',
            details: getSessionError.message
          })
          setState('error')
          return
        }

        if (sessionData.session) {
          console.log('✅ Existing session found! Redirecting...', sessionData.session.user.email)
          window.location.href = '/dashboard'
          return
        }

        // No valid authentication found
        setError({
          message: 'No valid authentication found.',
          details: 'No access token, code, or existing session available.'
        })
        setState('error')

      } catch (error) {
        console.error('Unexpected auth callback error:', error)
        setError({
          message: 'An unexpected error occurred during authentication.',
          details: error instanceof Error ? error.message : 'Unknown error'
        })
        setState('error')
      }
    }

    handleAuthCallback()
  }, [router])

  const handleRetryAuth = () => {
    router.push('/login')
  }

  const handleManualRedirect = () => {
    // Use window.location to force a full page reload
    window.location.href = redirectTarget
  }

  // Loading state
  if (state === 'loading') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background to-muted/20 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <LoadingSpinner 
              size="lg" 
              message="Completing authentication..." 
              className="py-8"
            />
            <div className="text-center text-sm text-muted-foreground mt-4">
              <p>Please wait while we verify your GitHub account...</p>
              <p className="mt-2">This usually takes just a few seconds.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Success state
  if (state === 'success') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-background to-muted/20 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center pb-2">
            <div className="mx-auto w-12 h-12 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center mb-4">
              <CheckCircle className="w-6 h-6 text-green-600 dark:text-green-400" />
            </div>
            <CardTitle className="text-xl text-green-600 dark:text-green-400">
              Authentication Successful!
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-center">
            <p className="text-sm text-muted-foreground">
              You have successfully signed in with GitHub.
            </p>
            
            <div className="p-4 bg-muted/50 rounded-lg">
              <p className="text-sm font-medium mb-2">
                Redirecting to dashboard in {countdown} seconds...
              </p>
              <Button 
                onClick={handleManualRedirect}
                className="w-full gap-2"
              >
                Continue to Dashboard
                <ArrowRight className="w-4 h-4" />
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Error state
  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-muted/20 flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center pb-2">
          <div className="mx-auto w-12 h-12 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center mb-4">
            <AlertCircle className="w-6 h-6 text-red-600 dark:text-red-400" />
          </div>
          <CardTitle className="text-xl text-red-600 dark:text-red-400">
            Authentication Failed
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              <div className="space-y-2">
                <p className="font-medium">{error?.message}</p>
                {error?.details && (
                  <p className="text-xs opacity-80">{error.details}</p>
                )}
              </div>
            </AlertDescription>
          </Alert>

          <div className="space-y-2">
            <Button 
              onClick={handleRetryAuth}
              className="w-full"
            >
              Try Again
            </Button>
            <Button 
              onClick={() => window.location.reload()}
              variant="outline"
              size="sm"
              className="w-full"
            >
              Refresh Page
            </Button>
          </div>

          <div className="text-center">
            <p className="text-xs text-muted-foreground">
              If the problem persists, please contact support.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}