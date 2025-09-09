'use client'

import { useState } from 'react'
import { LogOut } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface LogoutButtonProps {
  variant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  className?: string
  showText?: boolean
}

export function LogoutButton({ 
  variant = 'outline', 
  size = 'sm', 
  className,
  showText = true 
}: LogoutButtonProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const { signOut, user } = useAuth()

  const handleLogoutClick = () => {
    setIsDialogOpen(true)
  }

  const handleConfirmLogout = async () => {
    try {
      setIsLoggingOut(true)
      await signOut()
      setIsDialogOpen(false)
    } catch (error) {
      console.error('Logout error:', error)
      // Error is handled by the AuthContext
    } finally {
      setIsLoggingOut(false)
    }
  }

  const handleCancelLogout = () => {
    setIsDialogOpen(false)
  }

  if (!user) {
    return null
  }

  return (
    <>
      <Button
        variant={variant}
        size={size}
        onClick={handleLogoutClick}
        className={className}
        disabled={isLoggingOut}
      >
        <LogOut className="h-4 w-4" />
        {showText && <span className="ml-2">Logout</span>}
      </Button>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Logout</DialogTitle>
            <DialogDescription>
              Are you sure you want to sign out? This will clear your session and any cached data.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={handleCancelLogout}
              disabled={isLoggingOut}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmLogout}
              disabled={isLoggingOut}
              className="gap-2"
            >
              {isLoggingOut && <div className="h-4 w-4 animate-spin rounded-full border-2 border-background border-t-transparent" />}
              <LogOut className="h-4 w-4" />
              Sign Out
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}