'use client'

import { useAuth } from '@/contexts/AuthContext'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { GitHubButton } from '@/components/GitHubButton'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState, Suspense } from 'react'
import { AlertCircle, ShieldCheck } from 'lucide-react'

function LoginContent() {
  const { signInWithGitHub, loading, error, isAuthenticated } = useAuth()
  const router = useRouter()
  const searchParams = useSearchParams()
  const [urlError, setUrlError] = useState<string | null>(null)

  // Check for errors from URL parameters
  useEffect(() => {
    const errorParam = searchParams.get('error')
    if (errorParam) {
      switch (errorParam) {
        case 'auth_callback_failed':
          setUrlError('Authentication failed during callback. Please try again.')
          break
        case 'no_session':
          setUrlError('No valid session found. Please try signing in again.')
          break
        case 'unexpected_error':
          setUrlError('An unexpected error occurred during authentication.')
          break
        default:
          setUrlError('An error occurred during authentication.')
      }
      // Clean URL by removing error parameter
      const newUrl = new URL(window.location.href)
      newUrl.searchParams.delete('error')
      window.history.replaceState({}, '', newUrl.toString())
    }
  }, [searchParams])

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      const redirectTo = searchParams.get('redirect') || '/dashboard'
      router.push(redirectTo)
    }
  }, [isAuthenticated, router, searchParams])

  const handleGitHubSignIn = async () => {
    setUrlError(null) // Clear any previous URL errors
    await signInWithGitHub()
  }

  if (loading) {
    return (
      <LoadingSpinner 
        fullscreen 
        size="lg" 
        message="Checking authentication..." 
      />
    )
  }

  const displayError = error || urlError

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-muted/20 flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6">
        {/* Logo/Brand Section */}
        <div className="text-center space-y-2">
          <div className="mx-auto w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
            <ShieldCheck className="w-6 h-6 text-primary-foreground" />
          </div>
          <h1 className="text-3xl font-bold tracking-tight">
            Welcome Back
          </h1>
          <p className="text-muted-foreground text-sm">
            Sign in to access your workflow dashboard
          </p>
        </div>

        {/* Main Login Card */}
        <Card className="border-border/50 shadow-lg">
          <CardHeader className="space-y-2 text-center">
            <CardTitle className="text-xl">Sign in to your account</CardTitle>
            <CardDescription>
              Authenticate with your GitHub account to continue
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Error Display */}
            {displayError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription className="text-sm">
                  {displayError}
                </AlertDescription>
              </Alert>
            )}

            {/* GitHub Sign In Button */}
            <GitHubButton
              onClick={handleGitHubSignIn}
              className="w-full h-11"
              size="lg"
            />

            {/* Additional Info */}
            <div className="text-center text-xs text-muted-foreground space-y-2">
              <p>
                By signing in, you agree to our terms of service and privacy policy.
              </p>
              <div className="flex items-center justify-center gap-1">
                <ShieldCheck className="w-3 h-3" />
                <span>Secure OAuth authentication</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Help Text */}
        <div className="text-center text-sm text-muted-foreground space-y-2">
          <p>
            Need help? Contact support or check our documentation.
          </p>
        </div>
      </div>
    </div>
  )
}

export default function LoginPage() {
  return (
    <Suspense fallback={
      <LoadingSpinner 
        fullscreen 
        size="lg" 
        message="Loading..." 
      />
    }>
      <LoginContent />
    </Suspense>
  )
}