'use client'

import { Github, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { LogoutButton } from '@/components/LogoutButton'
import { ThemeToggle } from '@/components/ThemeToggle'
import { LanguageToggle } from '@/components/LanguageToggle'
import { useAuth } from '@/contexts/AuthContext'
import { useI18n } from '@/contexts/I18nContext'

interface HeaderProps {
  title: string
  onRefresh?: () => void
  isLoading?: boolean
  showSettings?: boolean
  children?: React.ReactNode
}

export function Header({ 
  title, 
  onRefresh, 
  isLoading = false, 
  showSettings = false,
  children 
}: HeaderProps) {
  const { user, isAuthenticated } = useAuth()
  const { t } = useI18n()

  return (
    <header className="border-b bg-background">
      <div className="container mx-auto max-w-7xl px-4 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Github className="h-8 w-8" />
            <h1 className="text-2xl font-bold">{title}</h1>
          </div>
          
          <div className="flex items-center gap-4">
            {/* Custom action buttons */}
            {children}
            
            {/* Language toggle */}
            <LanguageToggle />
            
            {/* Theme toggle */}
            <ThemeToggle />
            
            {/* Refresh button */}
            {onRefresh && (
              <Button
                variant="outline"
                size="sm"
                onClick={onRefresh}
                disabled={isLoading}
                className="gap-2"
              >
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                {t('dashboard.refresh')}
              </Button>
            )}
            
            {/* Settings button */}
            {showSettings && (
              <Button variant="outline" size="sm">
                {t('common.settings', 'Settings')}
              </Button>
            )}
            
            {/* User info and logout */}
            {isAuthenticated && user && (
              <div className="flex items-center gap-3">
                <div className="hidden sm:block text-sm text-muted-foreground">
                  {user.email}
                </div>
                <LogoutButton />
              </div>
            )}
            
            {/* Sign in button for unauthenticated users */}
            {!isAuthenticated && (
              <Button 
                onClick={() => window.location.href = '/login'}
                className="gap-2"
              >
                <Github className="h-4 w-4" />
                {t('auth.signIn')}
              </Button>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}