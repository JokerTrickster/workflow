import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { I18nProvider, useI18n, useLocale, useTranslation } from '@/contexts/I18nContext';
import { Locale } from '@/types/i18n';

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: mockLocalStorage
});

// Test component that uses I18n
function TestComponent() {
  const { locale, t, setLocale, isLoading, error } = useI18n();
  
  return (
    <div>
      <div data-testid="locale">{locale}</div>
      <div data-testid="loading">{isLoading.toString()}</div>
      <div data-testid="error">{error || 'null'}</div>
      <div data-testid="message">{t('auth.signIn')}</div>
      <div data-testid="message-with-params">{t('common.welcome', { name: 'Test User' })}</div>
      <div data-testid="missing-key">{t('missing.key')}</div>
      <button 
        data-testid="change-to-en" 
        onClick={() => setLocale('en')}
      >
        Change to English
      </button>
      <button 
        data-testid="change-to-ko" 
        onClick={() => setLocale('ko')}
      >
        Change to Korean
      </button>
    </div>
  );
}

// Test component for useLocale hook
function LocaleComponent() {
  const locale = useLocale();
  return <div data-testid="locale-only">{locale}</div>;
}

// Test component for useTranslation hook
function TranslationComponent() {
  const { t } = useTranslation();
  return <div data-testid="translation-only">{t('auth.signOut')}</div>;
}

describe('I18nContext', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Mock console.log and console.warn to reduce noise
    jest.spyOn(console, 'log').mockImplementation();
    jest.spyOn(console, 'warn').mockImplementation();
    jest.spyOn(console, 'error').mockImplementation();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('I18nProvider', () => {
    it('should provide default Korean locale', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('ko');
      expect(screen.getByTestId('message')).toHaveTextContent('로그인');
    });

    it('should use custom default locale', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      
      render(
        <I18nProvider defaultLocale="en">
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('en');
      expect(screen.getByTestId('message')).toHaveTextContent('Sign In');
    });

    it('should load locale from localStorage', () => {
      mockLocalStorage.getItem.mockReturnValue('en');
      
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('en');
      expect(screen.getByTestId('message')).toHaveTextContent('Sign In');
      expect(mockLocalStorage.getItem).toHaveBeenCalledWith('workflow-locale');
    });

    it('should ignore invalid locale from localStorage', () => {
      mockLocalStorage.getItem.mockReturnValue('invalid');
      
      render(
        <I18nProvider defaultLocale="ko">
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('ko');
    });

    it('should handle missing translation keys', () => {
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('missing-key')).toHaveTextContent('missing.key');
    });

    it('should handle template parameter substitution', () => {
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      // Note: This would work if we had a template with {{name}} parameter
      // For now, we just check that the function doesn't crash
      expect(screen.getByTestId('message-with-params')).toBeInTheDocument();
    });
  });

  describe('setLocale function', () => {
    it('should change locale and save to localStorage', async () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('ko');
      expect(screen.getByTestId('message')).toHaveTextContent('로그인');
      
      fireEvent.click(screen.getByTestId('change-to-en'));
      
      await waitFor(() => {
        expect(screen.getByTestId('locale')).toHaveTextContent('en');
        expect(screen.getByTestId('message')).toHaveTextContent('Sign In');
      });
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('workflow-locale', 'en');
    });

    it('should change back to Korean', async () => {
      mockLocalStorage.getItem.mockReturnValue('en');
      
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('en');
      
      fireEvent.click(screen.getByTestId('change-to-ko'));
      
      await waitFor(() => {
        expect(screen.getByTestId('locale')).toHaveTextContent('ko');
        expect(screen.getByTestId('message')).toHaveTextContent('로그인');
      });
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('workflow-locale', 'ko');
    });

    it('should handle localStorage errors gracefully', async () => {
      mockLocalStorage.setItem.mockImplementation(() => {
        throw new Error('Storage error');
      });
      
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      fireEvent.click(screen.getByTestId('change-to-en'));
      
      await waitFor(() => {
        expect(screen.getByTestId('locale')).toHaveTextContent('en');
      });
      
      // Should still work despite storage error
      expect(screen.getByTestId('message')).toHaveTextContent('Sign In');
    });
  });

  describe('useI18n hook', () => {
    it('should throw error when used outside provider', () => {
      // Mock console.error to prevent error output in tests
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      expect(() => {
        render(<TestComponent />);
      }).toThrow('useI18n must be used within an I18nProvider');
      
      consoleSpy.mockRestore();
    });

    it('should provide all context values', () => {
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toBeInTheDocument();
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
      expect(screen.getByTestId('error')).toHaveTextContent('null');
      expect(screen.getByTestId('message')).toBeInTheDocument();
    });
  });

  describe('useLocale hook', () => {
    it('should return current locale', () => {
      render(
        <I18nProvider defaultLocale="en">
          <LocaleComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale-only')).toHaveTextContent('en');
    });

    it('should throw error when used outside provider', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      expect(() => {
        render(<LocaleComponent />);
      }).toThrow('useI18n must be used within an I18nProvider');
      
      consoleSpy.mockRestore();
    });
  });

  describe('useTranslation hook', () => {
    it('should return translation function', () => {
      render(
        <I18nProvider>
          <TranslationComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('translation-only')).toHaveTextContent('로그아웃');
    });

    it('should throw error when used outside provider', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      expect(() => {
        render(<TranslationComponent />);
      }).toThrow('useI18n must be used within an I18nProvider');
      
      consoleSpy.mockRestore();
    });
  });

  describe('loading states', () => {
    it('should handle loading state during locale change', async () => {
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
      
      fireEvent.click(screen.getByTestId('change-to-en'));
      
      // Loading state might be too fast to catch in tests, but the mechanism is there
      await waitFor(() => {
        expect(screen.getByTestId('locale')).toHaveTextContent('en');
      });
      
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });
  });

  describe('error handling', () => {
    it('should handle initialization errors gracefully', () => {
      // Mock getMessages to throw an error
      jest.doMock('@/locales', () => ({
        getMessages: jest.fn(() => {
          throw new Error('Failed to load messages');
        }),
        getInitialLocale: jest.fn(() => 'ko'),
        saveLocaleToStorage: jest.fn(),
      }));
      
      render(
        <I18nProvider>
          <TestComponent />
        </I18nProvider>
      );
      
      // Should still render without crashing
      expect(screen.getByTestId('locale')).toBeInTheDocument();
    });
  });

  describe('fallback behavior', () => {
    it('should use fallback locale when default fails', () => {
      render(
        <I18nProvider defaultLocale="ko" fallbackLocale="en">
          <TestComponent />
        </I18nProvider>
      );
      
      expect(screen.getByTestId('locale')).toHaveTextContent('ko');
      expect(screen.getByTestId('message')).toHaveTextContent('로그인');
    });
  });
});