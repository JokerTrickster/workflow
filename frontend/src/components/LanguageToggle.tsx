'use client'

import React, { memo } from 'react';
import { useI18n } from '@/contexts/I18nContext';
import { Locale, LOCALE_NAMES } from '@/types/i18n';

interface LanguageToggleProps {
  className?: string;
  variant?: 'button' | 'select';
  showText?: boolean;
}

/**
 * LanguageToggle - Component for switching between Korean and English
 */
const LanguageToggle = memo<LanguageToggleProps>(function LanguageToggle({ 
  className = '',
  variant = 'button',
  showText = true
}) {
  const { locale, setLocale, isLoading, t } = useI18n();
  
  const handleLocaleChange = async (newLocale: Locale) => {
    if (newLocale !== locale && !isLoading) {
      await setLocale(newLocale);
    }
  };
  
  const toggleLocale = () => {
    const newLocale: Locale = locale === 'ko' ? 'en' : 'ko';
    handleLocaleChange(newLocale);
  };

  if (variant === 'select') {
    return (
      <div className={`language-toggle-select ${className}`}>
        <select
          value={locale}
          onChange={(e) => handleLocaleChange(e.target.value as Locale)}
          disabled={isLoading}
          className="
            px-3 py-1.5 
            bg-white dark:bg-gray-800 
            border border-gray-300 dark:border-gray-600 
            rounded-md 
            text-sm 
            focus:outline-none focus:ring-2 focus:ring-blue-500 
            transition-colors duration-200
            disabled:opacity-50 disabled:cursor-not-allowed
          "
          aria-label={t('common.language', { fallback: 'Language' })}
        >
          <option value="ko">{LOCALE_NAMES.ko}</option>
          <option value="en">{LOCALE_NAMES.en}</option>
        </select>
      </div>
    );
  }
  
  // Button variant (default)
  return (
    <button
      onClick={toggleLocale}
      disabled={isLoading}
      className={`
        language-toggle-button
        inline-flex items-center gap-2
        px-3 py-1.5
        bg-transparent
        hover:bg-gray-100 dark:hover:bg-gray-800
        border border-gray-300 dark:border-gray-600
        rounded-md
        text-sm font-medium
        text-gray-700 dark:text-gray-300
        transition-all duration-200
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1
        disabled:opacity-50 disabled:cursor-not-allowed
        ${className}
      `}
      aria-label={t('theme.toggleTheme', { fallback: 'Toggle Language' })}
      title={`${t('common.language', { fallback: 'Language' })}: ${LOCALE_NAMES[locale]}`}
    >
      {/* Language indicator */}
      <span className="
        flex items-center justify-center
        w-6 h-6
        rounded-full
        bg-blue-100 dark:bg-blue-900
        text-blue-600 dark:text-blue-300
        text-xs font-bold
        transition-colors duration-200
      ">
        {locale.toUpperCase()}
      </span>
      
      {/* Text label (optional) */}
      {showText && (
        <span className="hidden sm:inline">
          {LOCALE_NAMES[locale]}
        </span>
      )}
      
      {/* Loading indicator */}
      {isLoading && (
        <span className="inline-block w-3 h-3">
          <svg 
            className="animate-spin w-full h-full text-gray-400" 
            fill="none" 
            viewBox="0 0 24 24"
          >
            <circle 
              className="opacity-25" 
              cx="12" 
              cy="12" 
              r="10" 
              stroke="currentColor" 
              strokeWidth="4"
            />
            <path 
              className="opacity-75" 
              fill="currentColor" 
              d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </span>
      )}
    </button>
  );
});

export { LanguageToggle };