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
};