import {
  getMessages,
  replaceParams,
  translate,
  isLocaleSupported,
  getBrowserLocale,
  saveLocaleToStorage,
  loadLocaleFromStorage,
  getInitialLocale,
} from '@/locales';
import { Locale } from '@/types/i18n';
import { koMessages } from '@/locales/ko';
import { enMessages } from '@/locales/en';

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

// Mock navigator.languages
const mockNavigator = {
  languages: ['ko-KR', 'ko', 'en-US', 'en'],
  language: 'ko-KR'
};

Object.defineProperty(window, 'navigator', {
  value: mockNavigator,
  configurable: true
});

describe('Locale utilities', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.spyOn(console, 'warn').mockImplementation();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('getMessages', () => {
    it('should return Korean messages for ko locale', () => {
      const messages = getMessages('ko');
      expect(messages).toEqual(koMessages);
      expect(messages.auth.signIn).toBe('로그인');
    });

    it('should return English messages for en locale', () => {
      const messages = getMessages('en');
      expect(messages).toEqual(enMessages);
      expect(messages.auth.signIn).toBe('Sign In');
    });

    it('should fallback to Korean for invalid locale', () => {
      const messages = getMessages('invalid' as Locale);
      expect(messages).toEqual(koMessages);
    });
  });

  describe('replaceParams', () => {
    it('should replace single parameter', () => {
      const template = 'Hello {{name}}!';
      const params = { name: 'World' };
      const result = replaceParams(template, params);
      expect(result).toBe('Hello World!');
    });

    it('should replace multiple parameters', () => {
      const template = '{{greeting}} {{name}}, welcome to {{app}}!';
      const params = { greeting: 'Hello', name: 'User', app: 'App' };
      const result = replaceParams(template, params);
      expect(result).toBe('Hello User, welcome to App!');
    });

    it('should handle missing parameters', () => {
      const template = 'Hello {{name}}, {{missing}} parameter!';
      const params = { name: 'User' };
      const result = replaceParams(template, params);
      expect(result).toBe('Hello User, {{missing}} parameter!');
    });

    it('should handle empty parameters', () => {
      const template = 'Hello {{name}}!';
      const result = replaceParams(template);
      expect(result).toBe('Hello {{name}}!');
    });

    it('should handle number and boolean parameters', () => {
      const template = 'Count: {{count}}, Active: {{active}}';
      const params = { count: 42, active: true };
      const result = replaceParams(template, params);
      expect(result).toBe('Count: 42, Active: true');
    });

    it('should handle template without parameters', () => {
      const template = 'No parameters here';
      const params = { unused: 'value' };
      const result = replaceParams(template, params);
      expect(result).toBe('No parameters here');
    });
  });

  describe('translate', () => {
    it('should translate direct key access', () => {
      const result = translate(koMessages, 'auth.signIn');
      expect(result).toBe('로그인');
    });

    it('should translate with parameters', () => {
      // Create a test message with template
      const testMessages = {
        ...koMessages,
        test: {
          welcome: '{{name}}님, 안녕하세요!'
        }
      } as any;
      
      const result = translate(testMessages, 'test.welcome', { name: '사용자' });
      expect(result).toBe('사용자님, 안녕하세요!');
    });

    it('should return key for missing translation', () => {
      const result = translate(koMessages, 'missing.key');
      expect(result).toBe('missing.key');
    });

    it('should warn for missing keys', () => {
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      translate(koMessages, 'missing.key');
      
      expect(consoleSpy).toHaveBeenCalledWith('Translation missing for key: missing.key');
      
      consoleSpy.mockRestore();
    });

    it('should handle nested object traversal', () => {
      const result = translate(koMessages, 'dashboard.filter.all');
      expect(result).toBe('전체');
    });

    it('should handle deep nesting', () => {
      const result = translate(enMessages, 'dashboard.filter.byLanguage');
      expect(result).toBe('Filter by Language');
    });
  });

  describe('isLocaleSupported', () => {
    it('should return true for supported locales', () => {
      expect(isLocaleSupported('ko')).toBe(true);
      expect(isLocaleSupported('en')).toBe(true);
    });

    it('should return false for unsupported locales', () => {
      expect(isLocaleSupported('fr')).toBe(false);
      expect(isLocaleSupported('ja')).toBe(false);
      expect(isLocaleSupported('invalid')).toBe(false);
    });

    it('should handle empty string', () => {
      expect(isLocaleSupported('')).toBe(false);
    });
  });

  describe('getBrowserLocale', () => {
    it('should detect Korean from browser languages', () => {
      const result = getBrowserLocale();
      expect(result).toBe('ko');
    });

    it('should detect English when available', () => {
      Object.defineProperty(window, 'navigator', {
        value: { languages: ['en-US', 'en'], language: 'en-US' },
        configurable: true
      });
      
      const result = getBrowserLocale();
      expect(result).toBe('en');
    });

    it('should fallback to Korean for unsupported languages', () => {
      Object.defineProperty(window, 'navigator', {
        value: { languages: ['fr-FR', 'fr'], language: 'fr-FR' },
        configurable: true
      });
      
      const result = getBrowserLocale();
      expect(result).toBe('ko');
    });

    it('should handle missing navigator.languages', () => {
      Object.defineProperty(window, 'navigator', {
        value: { language: 'ko-KR' },
        configurable: true
      });
      
      const result = getBrowserLocale();
      expect(result).toBe('ko');
    });

    it('should fallback when server-side (no window)', () => {
      const originalWindow = global.window;
      delete (global as any).window;
      
      const result = getBrowserLocale();
      expect(result).toBe('ko');
      
      global.window = originalWindow;
    });
  });

  describe('saveLocaleToStorage', () => {
    it('should save locale to localStorage', () => {
      saveLocaleToStorage('en');
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('workflow-locale', 'en');
    });

    it('should handle localStorage errors', () => {
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      mockLocalStorage.setItem.mockImplementation(() => {
        throw new Error('Storage error');
      });
      
      saveLocaleToStorage('ko');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'Failed to save locale to localStorage:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle server-side (no window)', () => {
      const originalWindow = global.window;
      delete (global as any).window;
      
      expect(() => {
        saveLocaleToStorage('ko');
      }).not.toThrow();
      
      global.window = originalWindow;
    });
  });

  describe('loadLocaleFromStorage', () => {
    it('should load valid locale from localStorage', () => {
      mockLocalStorage.getItem.mockReturnValue('en');
      const result = loadLocaleFromStorage();
      expect(result).toBe('en');
      expect(mockLocalStorage.getItem).toHaveBeenCalledWith('workflow-locale');
    });

    it('should return null for invalid locale', () => {
      mockLocalStorage.getItem.mockReturnValue('invalid');
      const result = loadLocaleFromStorage();
      expect(result).toBeNull();
    });

    it('should return null when no value stored', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      const result = loadLocaleFromStorage();
      expect(result).toBeNull();
    });

    it('should handle localStorage errors', () => {
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      mockLocalStorage.getItem.mockImplementation(() => {
        throw new Error('Storage error');
      });
      
      const result = loadLocaleFromStorage();
      expect(result).toBeNull();
      expect(consoleSpy).toHaveBeenCalledWith(
        'Failed to load locale from localStorage:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle server-side (no window)', () => {
      const originalWindow = global.window;
      delete (global as any).window;
      
      const result = loadLocaleFromStorage();
      expect(result).toBeNull();
      
      global.window = originalWindow;
    });
  });

  describe('getInitialLocale', () => {
    it('should prioritize stored locale', () => {
      mockLocalStorage.getItem.mockReturnValue('en');
      Object.defineProperty(window, 'navigator', {
        value: { languages: ['ko-KR'], language: 'ko-KR' },
        configurable: true
      });
      
      const result = getInitialLocale();
      expect(result).toBe('en');
    });

    it('should use browser locale when no stored value', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      Object.defineProperty(window, 'navigator', {
        value: { languages: ['en-US'], language: 'en-US' },
        configurable: true
      });
      
      const result = getInitialLocale();
      expect(result).toBe('en');
    });

    it('should use default when no stored or browser locale', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      Object.defineProperty(window, 'navigator', {
        value: { languages: ['fr-FR'], language: 'fr-FR' },
        configurable: true
      });
      
      const result = getInitialLocale('en');
      expect(result).toBe('en');
    });

    it('should use fallback default when no parameters', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      Object.defineProperty(window, 'navigator', {
        value: { languages: ['fr-FR'], language: 'fr-FR' },
        configurable: true
      });
      
      const result = getInitialLocale();
      expect(result).toBe('ko'); // Default fallback
    });
  });
});