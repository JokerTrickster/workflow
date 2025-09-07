'use client'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'

interface GitHubButtonProps {
  onClick: () => Promise<void>
  disabled?: boolean
  loading?: boolean
  className?: string
  size?: 'default' | 'sm' | 'lg'
  variant?: 'default' | 'outline'
  children?: React.ReactNode
}

// GitHub logo SVG component
const GitHubIcon = ({ className }: { className?: string }) => (
  <svg
    viewBox="0 0 16 16"
    fill="currentColor"
    className={cn("w-4 h-4", className)}
    aria-hidden="true"
  >
    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
  </svg>
)

export function GitHubButton({
  onClick,
  disabled = false,
  loading = false,
  className,
  size = 'default',
  variant = 'outline',
  children = 'Continue with GitHub'
}: GitHubButtonProps) {
  const [isLoading, setIsLoading] = useState(false)
  
  const handleClick = async () => {
    if (disabled || loading || isLoading) return
    
    try {
      setIsLoading(true)
      await onClick()
    } catch (error) {
      console.error('GitHub sign-in error:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const isButtonLoading = loading || isLoading

  return (
    <Button
      onClick={handleClick}
      disabled={disabled || isButtonLoading}
      variant={variant}
      size={size}
      className={cn(
        'relative transition-all duration-200',
        // GitHub-specific styling
        variant === 'outline' && [
          'border-gray-300 bg-white text-gray-900 hover:bg-gray-50',
          'dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100 dark:hover:bg-gray-700',
          'focus:ring-2 focus:ring-gray-500 focus:ring-offset-2',
          'dark:focus:ring-gray-400 dark:focus:ring-offset-gray-800'
        ],
        variant === 'default' && [
          'bg-gray-900 text-white hover:bg-gray-800',
          'dark:bg-gray-100 dark:text-gray-900 dark:hover:bg-gray-200'
        ],
        isButtonLoading && 'opacity-80 cursor-not-allowed',
        className
      )}
      aria-label="Sign in with GitHub"
    >
      <div className="flex items-center justify-center gap-2">
        {isButtonLoading ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : (
          <GitHubIcon />
        )}
        <span className="font-medium">
          {isButtonLoading ? 'Signing in...' : children}
        </span>
      </div>
    </Button>
  )
}

export default GitHubButton