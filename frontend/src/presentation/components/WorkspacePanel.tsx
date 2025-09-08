'use client';

import { useState, useEffect } from 'react';
import { Repository } from '../../domain/entities/Repository';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../components/ui/tabs';
import { Button } from '../../components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card';
import { Badge } from '../../components/ui/badge';
import { ExternalLink, Activity, BarChart3, AlertCircle, CheckCircle } from 'lucide-react';
import { TaskTab } from './tabs/TaskTab';

interface WorkspacePanelProps {
  repository: Repository;
  onClose: () => void;
}

// Placeholder components for the other two tabs

const LogsTab = () => {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Activity className="h-5 w-5" />
          Activity Logs
        </h2>
      </div>
      <div className="text-center py-12">
        <Activity className="h-12 w-12 mx-auto mb-4 opacity-50" />
        <h3 className="text-lg font-semibold mb-2">No logs available yet</h3>
        <p className="text-sm text-muted-foreground">
          Activity logs will appear here when tasks are executed
        </p>
      </div>
    </div>
  );
};

const DashboardTab = () => {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <BarChart3 className="h-5 w-5" />
          Dashboard
        </h2>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Tasks</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">12</div>
            <p className="text-xs text-muted-foreground">
              +2 from last week
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Completed</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">8</div>
            <p className="text-xs text-muted-foreground">
              66% completion rate
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">In Progress</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">3</div>
            <p className="text-xs text-muted-foreground">
              Active tasks
            </p>
          </CardContent>
        </Card>
      </div>
      <div className="text-center py-8">
        <BarChart3 className="h-12 w-12 mx-auto mb-4 opacity-50" />
        <h3 className="text-lg font-semibold mb-2">Dashboard Coming Soon</h3>
        <p className="text-sm text-muted-foreground">
          Detailed analytics and insights will be available here
        </p>
      </div>
    </div>
  );
};


export function WorkspacePanel({ repository, onClose }: WorkspacePanelProps) {
  const [activeTab, setActiveTab] = useState<string>('tasks');

  // Tab state persistence
  useEffect(() => {
    const savedTab = localStorage.getItem(`workspace-tab-${repository.id}`);
    if (savedTab && ['tasks', 'logs', 'dashboard'].includes(savedTab)) {
      setActiveTab(savedTab);
    }
  }, [repository.id]);

  const handleTabChange = (value: string) => {
    setActiveTab(value);
    localStorage.setItem(`workspace-tab-${repository.id}`, value);
  };

  // Show connection message for non-connected repositories
  if (!repository.is_connected) {
    return (
      <div className="fixed inset-0 bg-background z-50 flex flex-col safe-area-inset">
        {/* Header */}
        <div className="border-b bg-card">
          <div className="container mx-auto max-w-7xl px-4 py-4 px-safe-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <h1 className="text-xl md:text-2xl font-bold">{repository.name}</h1>
                  <Badge variant={repository.private ? "secondary" : "outline"}>
                    {repository.private ? 'Private' : 'Public'}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">{repository.description}</p>
              </div>
              <Button variant="outline" onClick={onClose} className="shrink-0">
                ← Back
              </Button>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 container mx-auto max-w-7xl px-4 py-6 px-safe-4">
          <div className="flex flex-col items-center justify-center h-full text-center space-y-6">
            <AlertCircle className="h-16 w-16 text-muted-foreground" />
            <div className="space-y-2">
              <h2 className="text-2xl font-bold">Repository Not Connected</h2>
              <p className="text-muted-foreground max-w-md">
                This repository needs to be connected before you can access the workspace features.
                Please connect the repository first to view tasks, logs, and dashboard.
              </p>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={onClose}>
                ← Back to Repositories
              </Button>
              <Button variant="outline" asChild>
                <a href={repository.html_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4 mr-2" />
                  View on GitHub
                </a>
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // For connected repositories, show the 3-tab interface
  return (
    <div className="fixed inset-0 bg-background z-50 flex flex-col safe-area-inset">
      {/* Header */}
      <div className="border-b bg-card">
        <div className="container mx-auto max-w-7xl px-4 py-4 px-safe-4">
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <h1 className="text-xl md:text-2xl font-bold">{repository.name}</h1>
                <Badge variant={repository.private ? "secondary" : "outline"}>
                  {repository.private ? 'Private' : 'Public'}
                </Badge>
                <Badge variant="default" className="bg-green-100 text-green-800">
                  Connected
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">{repository.description}</p>
            </div>
            <div className="flex gap-2 items-center">
              <Button variant="outline" size="sm" asChild>
                <a href={repository.html_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4 mr-2" />
                  View on GitHub
                </a>
              </Button>
              <Button variant="outline" onClick={onClose} className="shrink-0">
                ← Back
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 container mx-auto max-w-7xl px-4 py-6 px-safe-4">
        <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
          <TabsList className="grid w-full grid-cols-3 mb-6">
            <TabsTrigger value="tasks" className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4" />
              <span className="hidden sm:inline">Tasks</span>
            </TabsTrigger>
            <TabsTrigger value="logs" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              <span className="hidden sm:inline">Logs</span>
            </TabsTrigger>
            <TabsTrigger value="dashboard" className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              <span className="hidden sm:inline">Dashboard</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="tasks" className="space-y-4">
            <TaskTab repository={repository} />
          </TabsContent>

          <TabsContent value="logs" className="space-y-4">
            <LogsTab />
          </TabsContent>

          <TabsContent value="dashboard" className="space-y-4">
            <DashboardTab />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}