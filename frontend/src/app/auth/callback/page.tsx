'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { supabase } from '@/lib/supabase/client'

export default function AuthCallbackPage() {
  const router = useRouter()

  useEffect(() => {
    const handleAuthCallback = async () => {
      try {
        const { data, error } = await supabase.auth.getSession()
        
        if (error) {
          console.error('Auth callback error:', error)
          router.push('/login?error=auth_callback_failed')
          return
        }

        if (data.session) {
          // Successful authentication - redirect to dashboard or intended page
          const redirectTo = new URLSearchParams(window.location.search).get('redirect') || '/dashboard'
          router.push(redirectTo)
        } else {
          // No session found - redirect to login
          router.push('/login?error=no_session')
        }
      } catch (error) {
        console.error('Unexpected auth callback error:', error)
        router.push('/login?error=unexpected_error')
      }
    }

    handleAuthCallback()
  }, [router])

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="space-y-4 text-center">
        <div className="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        <h1 className="text-xl font-semibold">Completing authentication...</h1>
        <p className="text-sm text-muted-foreground">
          Please wait while we complete your sign-in process.
        </p>
      </div>
    </div>
  )
}