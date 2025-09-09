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
    <header className="border-b bg-background" role="banner">
      <div className="container mx-auto max-w-7xl px-4 py-4 px-safe-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 min-w-0">
            <Github className="h-6 w-6 sm:h-8 sm:w-8 shrink-0" aria-hidden="true" />
            <h1 className="text-lg sm:text-2xl font-bold truncate">{title}</h1>
          </div>
          
          <nav className="flex items-center gap-2 sm:gap-4" role="navigation" aria-label="사이트 내비게이션">
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
                className="gap-2 touch-target hidden sm:flex"
                aria-label={`${t('dashboard.refresh')} ${isLoading ? '(로딩 중)' : ''}`}
              >
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} aria-hidden="true" />
                {t('dashboard.refresh')}
              </Button>
            )}
            
            {/* Mobile refresh button - icon only */}
            {onRefresh && (
              <Button
                variant="outline"
                size="sm"
                onClick={onRefresh}
                disabled={isLoading}
                className="touch-target sm:hidden"
                aria-label={`${t('dashboard.refresh')} ${isLoading ? '(로딩 중)' : ''}`}
              >
                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
              </Button>
            )}
            
            {/* Settings button */}
            {showSettings && (
              <Button 
                variant="outline" 
                size="sm" 
                className="touch-target"
                aria-label={t('common.settings', 'Settings')}
              >
                <span className="hidden sm:inline">{t('common.settings', 'Settings')}</span>
                <span className="sm:hidden">⚙️</span>
              </Button>
            )}
            
            {/* User info and logout */}
            {isAuthenticated && user && (
              <div className="flex items-center gap-2 sm:gap-3">
                <div className="hidden md:block text-sm text-muted-foreground truncate max-w-[150px]" title={user.email}>
                  {user.email}
                </div>
                <LogoutButton />
              </div>
            )}
            
            {/* Sign in button for unauthenticated users */}
            {!isAuthenticated && (
              <Button 
                onClick={() => window.location.href = '/login'}
                className="gap-2 touch-target"
                aria-label={t('auth.signIn')}
              >
                <Github className="h-4 w-4" aria-hidden="true" />
                <span className="hidden sm:inline">{t('auth.signIn')}</span>
              </Button>
            )}
          </nav>
        </div>
      </div>
    </header>
  )
}