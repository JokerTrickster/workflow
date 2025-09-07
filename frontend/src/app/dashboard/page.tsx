'use client'

import { useAuth } from '@/contexts/AuthContext'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { format } from 'date-fns'
import { Calendar, ExternalLink, Github, LogOut, User } from 'lucide-react'
import { RepositoriesTest } from '@/components/RepositoriesTest'

export default function DashboardPage() {
  const { user, session, loading, error, signOut, isAuthenticated } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      router.push('/login')
    }
  }, [loading, isAuthenticated, router])

  const handleSignOut = async () => {
    const { error } = await signOut()
    if (!error) {
      router.push('/login')
    }
  }

  if (loading) {
    return (
      <LoadingSpinner 
        fullscreen 
        size="lg" 
        message="Loading dashboard..." 
      />
    )
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-center text-destructive">Error</CardTitle>
          </CardHeader>
          <CardContent className="text-center">
            <p className="text-sm text-muted-foreground mb-4">{error}</p>
            <Button onClick={() => router.push('/login')} variant="outline">
              Back to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!isAuthenticated || !user) {
    return null
  }

  const userMetadata = user.user_metadata || {}
  const avatarUrl = userMetadata.avatar_url
  const fullName = userMetadata.full_name
  const username = userMetadata.user_name || userMetadata.preferred_username
  const publicRepos = userMetadata.public_repos
  const followers = userMetadata.followers
  const following = userMetadata.following

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">Dashboard</h1>
              <Badge variant="secondary">GitHub Connected</Badge>
            </div>
            <Button 
              onClick={handleSignOut} 
              variant="outline" 
              size="sm"
              className="gap-2"
            >
              <LogOut className="w-4 h-4" />
              Sign Out
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {/* User Profile Card */}
          <Card className="md:col-span-2 lg:col-span-2">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="w-5 h-5" />
                User Profile
              </CardTitle>
              <CardDescription>
                Your GitHub account information
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center gap-4">
                <Avatar className="w-16 h-16">
                  <AvatarImage src={avatarUrl} alt={fullName || 'User avatar'} />
                  <AvatarFallback className="text-lg">
                    {fullName ? fullName.charAt(0).toUpperCase() : 'U'}
                  </AvatarFallback>
                </Avatar>
                <div className="space-y-1">
                  <h3 className="text-lg font-semibold">
                    {fullName || 'GitHub User'}
                  </h3>
                  {username && (
                    <p className="text-sm text-muted-foreground">
                      @{username}
                    </p>
                  )}
                  <p className="text-sm text-muted-foreground">
                    {user.email}
                  </p>
                </div>
              </div>

              <Separator />

              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {typeof publicRepos === 'number' && (
                  <div className="text-center">
                    <div className="text-2xl font-bold">{publicRepos}</div>
                    <div className="text-xs text-muted-foreground">Repositories</div>
                  </div>
                )}
                {typeof followers === 'number' && (
                  <div className="text-center">
                    <div className="text-2xl font-bold">{followers}</div>
                    <div className="text-xs text-muted-foreground">Followers</div>
                  </div>
                )}
                {typeof following === 'number' && (
                  <div className="text-center">
                    <div className="text-2xl font-bold">{following}</div>
                    <div className="text-xs text-muted-foreground">Following</div>
                  </div>
                )}
              </div>

              {username && (
                <div className="flex justify-center">
                  <Button 
                    variant="outline" 
                    size="sm" 
                    className="gap-2"
                    asChild
                  >
                    <a 
                      href={`https://github.com/${username}`}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Github className="w-4 h-4" />
                      View GitHub Profile
                      <ExternalLink className="w-3 h-3" />
                    </a>
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Session Info Card */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Calendar className="w-5 h-5" />
                Session Info
              </CardTitle>
              <CardDescription>
                Current authentication details
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-sm font-medium">Provider</label>
                <p className="text-sm text-muted-foreground capitalize">
                  {user.app_metadata?.provider || 'GitHub'}
                </p>
              </div>
              
              <div>
                <label className="text-sm font-medium">Last Sign In</label>
                <p className="text-sm text-muted-foreground">
                  {user.last_sign_in_at 
                    ? format(new Date(user.last_sign_in_at), 'PPp')
                    : 'Unknown'
                  }
                </p>
              </div>

              <div>
                <label className="text-sm font-medium">Session Expires</label>
                <p className="text-sm text-muted-foreground">
                  {session?.expires_at 
                    ? format(new Date(session.expires_at * 1000), 'PPp')
                    : 'Unknown'
                  }
                </p>
              </div>

              <div>
                <label className="text-sm font-medium">User ID</label>
                <p className="text-xs text-muted-foreground font-mono break-all">
                  {user.id}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* GitHub API Test */}
        <div className="mt-6">
          <RepositoriesTest />
        </div>

        {/* Quick Actions */}
        <Card className="mt-6">
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>
              Common tasks and navigation
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              <Button variant="outline" disabled>
                View Projects
                <span className="ml-2 text-xs text-muted-foreground">(Coming Soon)</span>
              </Button>
              <Button variant="outline" disabled>
                Settings
                <span className="ml-2 text-xs text-muted-foreground">(Coming Soon)</span>
              </Button>
              <Button variant="outline" disabled>
                Profile
                <span className="ml-2 text-xs text-muted-foreground">(Coming Soon)</span>
              </Button>
            </div>
          </CardContent>
        </Card>
      </main>
    </div>
  )
}