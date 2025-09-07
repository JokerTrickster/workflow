import { createServerClient } from '@supabase/ssr'
import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url)
  const code = searchParams.get('code')
  const next = searchParams.get('next') ?? '/'

  if (code) {
    const cookieStore = await cookies()
    const supabase = createServerClient(
      process.env.NEXT_PUBLIC_SUPABASE_URL!,
      process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
      {
        cookies: {
          getAll() {
            return cookieStore.getAll()
          },
          setAll(cookiesToSet) {
            try {
              cookiesToSet.forEach(({ name, value, options }) => {
                cookieStore.set(name, value, options)
              })
            } catch (error) {
              console.error('Cookie setting error:', error)
            }
          },
        },
      }
    )

    console.log('Attempting to exchange code for session:', code)
    const { data, error } = await supabase.auth.exchangeCodeForSession(code)
    
    if (!error && data.session) {
      console.log('Session created successfully:', data.session.user.email)
      const forwardedHost = request.headers.get('x-forwarded-host')
      const isLocalEnv = process.env.NODE_ENV === 'development'
      
      if (isLocalEnv) {
        // Local environment: simple redirect
        return NextResponse.redirect(`${origin}/dashboard`)
      } else if (forwardedHost) {
        // Production environment with forwarded host
        return NextResponse.redirect(`https://${forwardedHost}${next}`)
      } else {
        // Fallback
        return NextResponse.redirect(`${origin}${next}`)
      }
    } else {
      console.error('Auth callback error:', error)
      console.error('Error details:', {
        message: error?.message,
        name: error?.name,
        status: error?.status,
        code: error?.code
      })
    }
  }

  // Return the user to an error page with some instructions
  return NextResponse.redirect(`${origin}/login?error=auth_callback_failed`)
}