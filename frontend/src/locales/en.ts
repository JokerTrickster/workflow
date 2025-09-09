import { TranslationMessages } from '@/types/i18n';

/**
 * English translation messages
 */
export const enMessages: TranslationMessages = {
  // Authentication & User
  auth: {
    signIn: 'Sign In',
    signOut: 'Sign Out',
    signInWithGitHub: 'Sign in with GitHub',
    signOutConfirmation: 'Are you sure you want to sign out?',
    welcome: 'Welcome',
    welcomeBack: 'Welcome back',
    loading: 'Authenticating...',
    error: 'Authentication error occurred',
    unauthorized: 'Authentication required',
    sessionExpired: 'Session expired. Please sign in again',
  },
  
  // Dashboard & Navigation
  dashboard: {
    title: 'Dashboard',
    repositories: 'Repositories',
    search: 'Search',
    searchPlaceholder: 'Search repositories...',
    noRepositories: 'No repositories found',
    loading: 'Loading...',
    error: 'An error occurred',
    refresh: 'Refresh',
    filter: {
      all: 'All',
      connected: 'Connected',
      notConnected: 'Not Connected',
      byLanguage: 'Filter by Language',
    },
  },
  
  // Repository management
  repository: {
    connect: 'Connect',
    disconnect: 'Disconnect',
    connecting: 'Connecting...',
    connected: 'Connected',
    notConnected: 'Not Connected',
    lastUpdated: 'Last Updated',
    createdAt: 'Created At',
    language: 'Language',
    stars: 'Stars',
    forks: 'Forks',
    private: 'Private',
    public: 'Public',
    selectRepository: 'Select Repository',
    connectionSuccess: 'Repository connected successfully',
    connectionError: 'Failed to connect repository',
  },
  
  // Activity & Logging
  activity: {
    title: 'Activity',
    recent: 'Recent Activity',
    noActivity: 'No activity found',
    loading: 'Loading activity...',
    refresh: 'Refresh Activity',
    viewAll: 'View All',
    githubSync: 'GitHub Sync',
    apiCall: 'API Call',
    repositoryConnection: 'Repository Connection',
    userAction: 'User Action',
    systemEvent: 'System Event',
  },
  
  // GitHub Integration
  github: {
    connecting: 'Connecting to GitHub...',
    syncingRepositories: 'Syncing repositories...',
    fetchingIssues: 'Fetching issues...',
    fetchingPullRequests: 'Fetching pull requests...',
    rateLimit: 'GitHub API Rate Limit',
    rateLimitWarning: 'API rate limit will be exceeded soon',
    apiError: 'GitHub API error occurred',
    unauthorized: 'GitHub authentication required',
    repositoryNotFound: 'Repository not found',
  },
  
  // Theme & Settings
  theme: {
    light: 'Light Mode',
    dark: 'Dark Mode',
    system: 'System Setting',
    toggleTheme: 'Toggle Theme',
    themeChanged: 'Theme changed',
  },
  
  // Common UI elements
  common: {
    save: 'Save',
    cancel: 'Cancel',
    delete: 'Delete',
    edit: 'Edit',
    view: 'View',
    close: 'Close',
    back: 'Back',
    next: 'Next',
    previous: 'Previous',
    loading: 'Loading...',
    error: 'Error',
    success: 'Success',
    warning: 'Warning',
    info: 'Info',
    retry: 'Retry',
    confirm: 'Confirm',
    yes: 'Yes',
    no: 'No',
    settings: 'Settings',
    language: 'Language',
    filter: 'Filters',
  },
  
  // Error messages
  errors: {
    generic: 'An unknown error occurred',
    network: 'Please check your network connection',
    unauthorized: 'Unauthorized access',
    forbidden: 'Access denied',
    notFound: 'Requested resource not found',
    serverError: 'Server error occurred',
    validationError: 'Please check your input',
    requiredField: 'This field is required',
    invalidFormat: 'Invalid format',
  },
  
  // Success messages  
  success: {
    saved: 'Saved successfully',
    updated: 'Updated successfully',
    deleted: 'Deleted successfully',
    connected: 'Connected successfully',
    disconnected: 'Disconnected successfully',
    refreshed: 'Refreshed successfully',
  },

  // Activity logging messages
  logs: {
    title: 'Activity Logs',
    noLogs: 'No activity logs found',
    noLogsMessage: 'Try adjusting your filters to see more results or activity logs will appear here when tasks are executed and repository interactions occur',
    showingLogs: 'Showing {{count}} log{{pluralSuffix}}',
    exportTitle: 'Export',
    clearTitle: 'Clear',
    clearConfirmation: 'Are you sure you want to clear all activity logs? This action cannot be undone.',
    search: 'Search activity logs...',
    filterByType: 'Filter by type',
    selectTimeRange: 'Select time range',
    
    // Time ranges
    timeRanges: {
      lastHour: 'Last Hour',
      last24Hours: 'Last 24 Hours', 
      last7Days: 'Last 7 Days',
      last30Days: 'Last 30 Days',
      allTime: 'All Time'
    },

    // Activity types
    types: {
      all: 'All Activities',
      connection: 'Connection',
      task: 'Tasks',
      github: 'GitHub',
      navigation: 'Navigation'
    },

    // Repository connection messages
    repository_connected: 'Successfully connected to {{repositoryName}} repository',
    repository_disconnected: 'Disconnected from {{repositoryName}} repository',
    repository_connection_failed: 'Failed to connect to {{repositoryName}}: {{error}}',

    // Task messages
    task_created: 'New task "{{taskTitle}}" created for {{repositoryName}}',
    task_started: 'Task "{{taskTitle}}" execution started',
    task_completed: 'Task "{{taskTitle}}" completed successfully',
    task_failed: 'Task "{{taskTitle}}" failed: {{error}}',
    task_cancelled: 'Task "{{taskTitle}}" was cancelled',

    // GitHub messages
    github_sync_started: 'GitHub synchronization started for {{repositoryName}}',
    github_sync_completed: 'GitHub synchronization completed for {{repositoryName}}',
    github_sync_failed: 'GitHub synchronization failed for {{repositoryName}}',
    github_api_call: 'GitHub API call: {{method}} {{endpoint}}',
    github_rate_limit: 'GitHub API rate limit: {{remaining}} requests remaining. Resets at {{resetTime}}',
    github_rate_limit_warning: 'GitHub API rate limit will be exceeded soon',
    github_rate_limit_exceeded: 'GitHub API rate limit exceeded',

    // Navigation messages
    tab_switched: 'Switched from {{previousTab}} to {{currentTab}} tab in {{repositoryName}}',
    workspace_opened: 'Accessed workspace for {{repositoryName}}',
    workspace_closed: 'Closed workspace for {{repositoryName}}',

    // User actions
    user_actions: {
      repository_connected: 'Repository Connected',
      repository_disconnected: 'Repository Disconnected',
      repository_connection_failed: 'Repository Connection Failed',
      task_created: 'Task Created',
      task_started: 'Task Started',
      task_completed: 'Task Completed',
      task_failed: 'Task Failed',
      task_cancelled: 'Task Cancelled',
      github_sync_started: 'GitHub Sync Started',
      github_sync_completed: 'GitHub Sync Completed',
      github_sync_failed: 'GitHub Sync Failed',
      github_api_call: 'GitHub API Call',
      tab_switched: 'Tab Switched',
      workspace_opened: 'Workspace Opened',
      workspace_closed: 'Workspace Closed',
    }
  },
};