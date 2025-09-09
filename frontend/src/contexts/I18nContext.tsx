'use client'

import { createContext, useContext, useEffect, useState, useCallback } from 'react';
import type { ReactNode } from 'react';
import { 
  Locale, 
  I18nContextType, 
  I18nProviderProps, 
  TranslationMessages,
  TemplateParams,
  DEFAULT_LOCALE 
} from '@/types/i18n';
import { 
  getMessages, 
  translate, 
  getInitialLocale, 
  saveLocaleToStorage 
} from '@/locales';

// Create I18n Context
const I18nContext = createContext<I18nContextType | undefined>(undefined);

export { I18nContext };

/**
 * useI18n hook - Access internationalization context
 * @returns I18n context with locale, messages, and translation function
 */
export function useI18n(): I18nContextType {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
}

/**
 * I18nProvider - Provides internationalization context to the application
 */
export function I18nProvider({ 
  children, 
  defaultLocale = DEFAULT_LOCALE,
  fallbackLocale = 'ko'
}: I18nProviderProps) {
  const [locale, setCurrentLocale] = useState<Locale>(defaultLocale);
  const [messages, setMessages] = useState<TranslationMessages>(() => 
    getMessages(defaultLocale)
  );
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Initialize locale on mount
  useEffect(() => {
    const initializeLocale = async () => {
      try {
        setIsLoading(true);
        setError(null);
        
        // Determine initial locale
        const initialLocale = getInitialLocale(defaultLocale);
        
        console.log('Initializing I18n...');
        console.log('Default locale:', defaultLocale);
        console.log('Detected locale:', initialLocale);
        
        // Update state
        setCurrentLocale(initialLocale);
        setMessages(getMessages(initialLocale));
        
      } catch (err) {
        console.error('Failed to initialize I18n:', err);
        setError('Failed to initialize internationalization');
        
        // Fallback to default locale
        const fallback = fallbackLocale || DEFAULT_LOCALE;
        setCurrentLocale(fallback);
        setMessages(getMessages(fallback));
      } finally {
        setIsLoading(false);
      }
    };

    initializeLocale();
  }, [defaultLocale, fallbackLocale]);

  /**
   * Change locale and persist to storage
   * @param newLocale - Target locale
   */
  const setLocale = useCallback(async (newLocale: Locale) => {
    try {
      setIsLoading(true);
      setError(null);
      
      console.log(`Changing locale from ${locale} to ${newLocale}`);
      
      // Load messages for new locale
      const newMessages = getMessages(newLocale);
      
      // Update state
      setCurrentLocale(newLocale);
      setMessages(newMessages);
      
      // Persist to localStorage
      saveLocaleToStorage(newLocale);
      
      console.log(`Locale changed to ${newLocale} successfully`);
      
    } catch (err) {
      console.error('Failed to change locale:', err);
      setError('Failed to change language');
    } finally {
      setIsLoading(false);
    }
  }, [locale]);

  /**
   * Translation function with parameter substitution
   * @param key - Message key (supports dot notation)
   * @param params - Optional parameters for template substitution
   * @returns Translated message
   */
  const t = useCallback((
    key: string, 
    params?: TemplateParams
  ): string => {
    try {
      return translate(messages, key, params);
    } catch (err) {
      console.error('Translation error:', err);
      return key; // Return key as fallback
    }
  }, [messages]);

  // Context value
  const value: I18nContextType = {
    locale,
    messages,
    t,
    setLocale,
    isLoading,
    error
  };

  return (
    <I18nContext.Provider value={value}>
      {children}
    </I18nContext.Provider>
  );
}

/**
 * Higher-order component for providing I18n context
 * @param Component - Component to wrap
 * @param options - I18n provider options
 * @returns Component wrapped with I18nProvider
 */
export function withI18n<P extends object>(
  Component: React.ComponentType<P>,
  options?: Omit<I18nProviderProps, 'children'>
) {
  return function I18nWrappedComponent(props: P) {
    return (
      <I18nProvider {...options}>
        <Component {...props} />
      </I18nProvider>
    );
  };
}

/**
 * Utility hook for getting current locale without full context
 * @returns Current locale
 */
export function useLocale(): Locale {
  const { locale } = useI18n();
  return locale;
}

/**
 * Utility hook for getting translation function only
 * @returns Translation function
 */
export function useTranslation() {
  const { t } = useI18n();
  return { t };
}