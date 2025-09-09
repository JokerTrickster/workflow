import { Locale, TranslationMessages } from '@/types/i18n';
import { koMessages } from './ko';
import { enMessages } from './en';

/**
 * Locale message loader and utilities
 */

// Message dictionary
export const messages: Record<Locale, TranslationMessages> = {
  ko: koMessages,
  en: enMessages,
};

/**
 * Load messages for a specific locale
 * @param locale - Target locale
 * @returns Translation messages for the locale
 */
export function getMessages(locale: Locale): TranslationMessages {
  return messages[locale] || messages.ko; // Fallback to Korean
}

/**
 * Get value from nested object using dot notation
 * @param obj - Source object
 * @param path - Dot-separated path (e.g., "auth.signIn")
 * @returns Value at the path or undefined
 */
function getNestedValue(obj: any, path: string): string | undefined {
  return path.split('.').reduce((current, key) => {
    return current && typeof current === 'object' ? current[key] : undefined;
  }, obj);
}

/**
 * Replace template parameters in a message
 * @param template - Message template with {{param}} placeholders
 * @param params - Parameters to substitute
 * @returns Message with parameters replaced
 */
export function replaceParams(template: string, params?: Record<string, any>): string {
  if (!params || Object.keys(params).length === 0) {
    return template;
  }
  
  return template.replace(/\{\{(\w+)\}\}/g, (match, key) => {
    const value = params[key];
    return value !== undefined ? String(value) : match;
  });
}

/**
 * Get translated message with parameter substitution
 * @param messages - Translation messages
 * @param key - Message key (supports dot notation)
 * @param params - Optional parameters for substitution
 * @returns Translated and parameterized message
 */
export function translate(
  messages: TranslationMessages, 
  key: string, 
  params?: Record<string, any>
): string {
  // Try direct key access first
  const directValue = (messages as any)[key];
  if (typeof directValue === 'string') {
    return replaceParams(directValue, params);
  }
  
  // Try nested key access with dot notation
  const nestedValue = getNestedValue(messages, key);
  if (typeof nestedValue === 'string') {
    return replaceParams(nestedValue, params);
  }
  
  // Return the key if translation not found (development aid)
  console.warn(`Translation missing for key: ${key}`);
  return key;
}

/**
 * Check if a locale is supported
 * @param locale - Locale to check
 * @returns True if locale is supported
 */
export function isLocaleSupported(locale: string): locale is Locale {
  return ['ko', 'en'].includes(locale);
}

/**
 * Get browser locale with fallback
 * @returns Detected locale or fallback
 */
export function getBrowserLocale(): Locale {
  if (typeof window === 'undefined') {
    return 'ko'; // Server-side fallback
  }
  
  // Check navigator languages
  const languages = navigator.languages || [navigator.language];
  
  for (const lang of languages) {
    // Extract language code (e.g., 'ko-KR' -> 'ko')
    const langCode = lang.split('-')[0].toLowerCase();
    if (isLocaleSupported(langCode)) {
      return langCode;
    }
  }
  
  // Fallback to Korean
  return 'ko';
}

/**
 * Save locale to localStorage
 * @param locale - Locale to save
 */
export function saveLocaleToStorage(locale: Locale): void {
  if (typeof window === 'undefined') return;
  
  try {
    localStorage.setItem('workflow-locale', locale);
  } catch (error) {
    console.warn('Failed to save locale to localStorage:', error);
  }
}

/**
 * Load locale from localStorage
 * @returns Saved locale or null
 */
export function loadLocaleFromStorage(): Locale | null {
  if (typeof window === 'undefined') return null;
  
  try {
    const saved = localStorage.getItem('workflow-locale');
    return saved && isLocaleSupported(saved) ? saved : null;
  } catch (error) {
    console.warn('Failed to load locale from localStorage:', error);
    return null;
  }
}

/**
 * Determine initial locale based on storage, browser, and fallback
 * @param defaultLocale - Default locale if none found
 * @returns Resolved initial locale
 */
export function getInitialLocale(defaultLocale: Locale = 'ko'): Locale {
  // 1. Check localStorage first
  const storedLocale = loadLocaleFromStorage();
  if (storedLocale) {
    return storedLocale;
  }
  
  // 2. Check browser locale
  const browserLocale = getBrowserLocale();
  if (browserLocale) {
    return browserLocale;
  }
  
  // 3. Use default fallback
  return defaultLocale;
}