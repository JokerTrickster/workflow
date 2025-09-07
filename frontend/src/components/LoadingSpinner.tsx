'use client'

import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
  message?: string
  fullscreen?: boolean
}

const sizeClasses = {
  sm: 'w-4 h-4',
  md: 'w-6 h-6', 
  lg: 'w-8 h-8',
  xl: 'w-12 h-12'
}

export function LoadingSpinner({ 
  size = 'md', 
  className, 
  message,
  fullscreen = false 
}: LoadingSpinnerProps) {
  const spinner = (
    <div className={cn(
      'flex flex-col items-center justify-center gap-3',
      fullscreen && 'min-h-screen',
      className
    )}>
      <Loader2 
        className={cn(
          'animate-spin text-primary',
          sizeClasses[size]
        )} 
      />
      {message && (
        <p className={cn(
          'text-muted-foreground font-medium',
          size === 'sm' && 'text-xs',
          size === 'md' && 'text-sm',
          size === 'lg' && 'text-base',
          size === 'xl' && 'text-lg'
        )}>
          {message}
        </p>
      )}
    </div>
  )

  if (fullscreen) {
    return (
      <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
        {spinner}
      </div>
    )
  }

  return spinner
}

export default LoadingSpinner