'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '../components/ui/button';
import { Github, ArrowRight } from 'lucide-react';

export default function HomePage() {
  const router = useRouter();

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="text-center space-y-6 max-w-2xl mx-auto px-4">
        <div className="space-y-4">
          <Github className="h-16 w-16 mx-auto opacity-80" />
          <h1 className="text-4xl font-bold">AI Git Workbench</h1>
          <p className="text-xl text-muted-foreground">
            Manage your GitHub repositories with AI assistance
          </p>
          <p className="text-muted-foreground">
            Connect, organize, and streamline your workflow across multiple repositories
          </p>
        </div>
        
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Button 
            onClick={() => router.push('/dashboard')} 
            size="lg" 
            className="gap-2"
          >
            <ArrowRight className="h-5 w-5" />
            Go to Dashboard
          </Button>
          <Button variant="outline" size="lg" className="gap-2">
            <Github className="h-5 w-5" />
            Login with GitHub
          </Button>
        </div>

        <div className="grid md:grid-cols-3 gap-6 mt-12 pt-12 border-t">
          <div className="text-center space-y-2">
            <div className="h-12 w-12 bg-primary/10 rounded-lg mx-auto flex items-center justify-center">
              <Github className="h-6 w-6" />
            </div>
            <h3 className="font-semibold">Repository Management</h3>
            <p className="text-sm text-muted-foreground">
              View and manage all your GitHub repositories in one place
            </p>
          </div>
          
          <div className="text-center space-y-2">
            <div className="h-12 w-12 bg-primary/10 rounded-lg mx-auto flex items-center justify-center">
              <ArrowRight className="h-6 w-6" />
            </div>
            <h3 className="font-semibold">AI-Powered Workflow</h3>
            <p className="text-sm text-muted-foreground">
              Get intelligent suggestions and automate common tasks
            </p>
          </div>
          
          <div className="text-center space-y-2">
            <div className="h-12 w-12 bg-primary/10 rounded-lg mx-auto flex items-center justify-center">
              <Github className="h-6 w-6" />
            </div>
            <h3 className="font-semibold">Seamless Integration</h3>
            <p className="text-sm text-muted-foreground">
              Connect directly with GitHub for real-time synchronization
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
