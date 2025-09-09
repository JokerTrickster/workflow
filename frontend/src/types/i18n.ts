/**
 * Internationalization types for Korean localization system
 */

// Supported locales
export type Locale = 'ko' | 'en';

// Template parameter type for dynamic message substitution
export interface TemplateParams {
  [key: string]: string | number | boolean;
}

// Translation message structure
export interface TranslationMessages {
  // Authentication & User
  auth: {
    signIn: string;
    signOut: string;
    signInWithGitHub: string;
    signOutConfirmation: string;
    welcome: string;
    welcomeBack: string;
    loading: string;
    error: string;
    unauthorized: string;
    sessionExpired: string;
  };
  
  // Dashboard & Navigation
  dashboard: {
    title: string;
    repositories: string;
    search: string;
    searchPlaceholder: string;
    noRepositories: string;
    loading: string;
    error: string;
    refresh: string;
    filter: {
      all: string;
      connected: string;
      notConnected: string;
      byLanguage: string;
    };
  };
  
  // Repository management
  repository: {
    connect: string;
    disconnect: string;
    connecting: string;
    connected: string;
    notConnected: string;
    lastUpdated: string;
    createdAt: string;
    language: string;
    stars: string;
    forks: string;
    private: string;
    public: string;
    selectRepository: string;
    connectionSuccess: string;
    connectionError: string;
  };
  
  // Activity & Logging
  activity: {
    title: string;
    recent: string;
    noActivity: string;
    loading: string;
    refresh: string;
    viewAll: string;
    githubSync: string;
    apiCall: string;
    repositoryConnection: string;
    userAction: string;
    systemEvent: string;
  };
  
  // GitHub Integration
  github: {
    connecting: string;
    syncingRepositories: string;
    fetchingIssues: string;
    fetchingPullRequests: string;
    rateLimit: string;
    rateLimitWarning: string;
    apiError: string;
    unauthorized: string;
    repositoryNotFound: string;
  };
  
  // Theme & Settings
  theme: {
    light: string;
    dark: string;
    system: string;
    toggleTheme: string;
    themeChanged: string;
  };
  
  // Common UI elements
  common: {
    save: string;
    cancel: string;
    delete: string;
    edit: string;
    view: string;
    close: string;
    back: string;
    next: string;
    previous: string;
    loading: string;
    error: string;
    success: string;
    warning: string;
    info: string;
    retry: string;
    confirm: string;
    yes: string;
    no: string;
  };
  
  // Error messages
  errors: {
    generic: string;
    network: string;
    unauthorized: string;
    forbidden: string;
    notFound: string;
    serverError: string;
    validationError: string;
    requiredField: string;
    invalidFormat: string;
  };
  
  // Success messages  
  success: {
    saved: string;
    updated: string;
    deleted: string;
    connected: string;
    disconnected: string;
    refreshed: string;
  };
}

// I18n Context interface
export interface I18nContextType {
  locale: Locale;
  messages: TranslationMessages;
  t: (key: keyof TranslationMessages | string, params?: TemplateParams) => string;
  setLocale: (locale: Locale) => void;
  isLoading: boolean;
  error: string | null;
}

// I18n Provider props
export interface I18nProviderProps {
  children: React.ReactNode;
  defaultLocale?: Locale;
  fallbackLocale?: Locale;
}

// I18n hook return type (same as context for consistency)
export type UseI18nReturn = I18nContextType;

// Message dictionary structure for type safety
export type MessageKey = 
  | keyof TranslationMessages
  | `${keyof TranslationMessages}.${string}`
  | string;

// Utility type for nested message keys
export type DeepKeys<T> = T extends object
  ? {
      [K in keyof T]: T[K] extends object
        ? K | `${K & string}.${DeepKeys<T[K]> & string}`
        : K;
    }[keyof T]
  : never;

// Type-safe message key extraction
export type I18nKeys = DeepKeys<TranslationMessages>;

// Storage key for locale preference
export const LOCALE_STORAGE_KEY = 'workflow-locale' as const;

// Default locale fallback
export const DEFAULT_LOCALE: Locale = 'ko' as const;

// Available locales list
export const AVAILABLE_LOCALES: readonly Locale[] = ['ko', 'en'] as const;

// Locale display names
export const LOCALE_NAMES: Record<Locale, string> = {
  ko: '한국어',
  en: 'English'
} as const;