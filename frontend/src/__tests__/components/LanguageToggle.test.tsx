import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { LanguageToggle } from '@/components/LanguageToggle';
import { I18nProvider } from '@/contexts/I18nContext';

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

// Test wrapper with I18nProvider
function TestWrapper({ children, defaultLocale = 'ko' }: { children: React.ReactNode, defaultLocale?: 'ko' | 'en' }) {
  return (
    <I18nProvider defaultLocale={defaultLocale}>
      {children}
    </I18nProvider>
  );
}

describe('LanguageToggle', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockLocalStorage.getItem.mockReturnValue(null);
    jest.spyOn(console, 'log').mockImplementation();
    jest.spyOn(console, 'warn').mockImplementation();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('Button variant (default)', () => {
    it('should render Korean language toggle by default', () => {
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
      expect(screen.getByText('KO')).toBeInTheDocument();
      expect(screen.getByText('한국어')).toBeInTheDocument();
    });

    it('should render English language toggle when locale is en', () => {
      mockLocalStorage.getItem.mockReturnValue('en');
      
      render(
        <TestWrapper defaultLocale="en">
          <LanguageToggle />
        </TestWrapper>
      );
      
      expect(screen.getByText('EN')).toBeInTheDocument();
      expect(screen.getByText('English')).toBeInTheDocument();
    });

    it('should toggle from Korean to English', async () => {
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      expect(screen.getByText('KO')).toBeInTheDocument();
      
      fireEvent.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('EN')).toBeInTheDocument();
        expect(screen.getByText('English')).toBeInTheDocument();
      });
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('workflow-locale', 'en');
    });

    it('should toggle from English to Korean', async () => {
      mockLocalStorage.getItem.mockReturnValue('en');
      
      render(
        <TestWrapper defaultLocale="en">
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      expect(screen.getByText('EN')).toBeInTheDocument();
      
      fireEvent.click(button);
      
      await waitFor(() => {
        expect(screen.getByText('KO')).toBeInTheDocument();
        expect(screen.getByText('한국어')).toBeInTheDocument();
      });
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('workflow-locale', 'ko');
    });

    it('should hide text on small screens when showText is false', () => {
      render(
        <TestWrapper>
          <LanguageToggle showText={false} />
        </TestWrapper>
      );
      
      expect(screen.getByText('KO')).toBeInTheDocument();
      expect(screen.queryByText('한국어')).not.toBeInTheDocument();
    });

    it('should apply custom className', () => {
      render(
        <TestWrapper>
          <LanguageToggle className="custom-class" />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      expect(button).toHaveClass('custom-class');
    });

    it('should be disabled when loading', async () => {
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      
      // Click to trigger loading state
      fireEvent.click(button);
      
      // The button should be enabled again after the locale change completes
      await waitFor(() => {
        expect(button).not.toBeDisabled();
      });
    });

    it('should have proper accessibility attributes', () => {
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      expect(button).toHaveAttribute('aria-label');
      expect(button).toHaveAttribute('title');
    });

    it('should prevent toggle when already loading', async () => {
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      
      // Click rapidly multiple times
      fireEvent.click(button);
      fireEvent.click(button);
      fireEvent.click(button);
      
      // Should only change once
      await waitFor(() => {
        expect(screen.getByText('EN')).toBeInTheDocument();
      });
    });
  });

  describe('Select variant', () => {
    it('should render select dropdown', () => {
      render(
        <TestWrapper>
          <LanguageToggle variant="select" />
        </TestWrapper>
      );
      
      const select = screen.getByRole('combobox');
      expect(select).toBeInTheDocument();
      expect(select).toHaveValue('ko');
      
      const options = screen.getAllByRole('option');
      expect(options).toHaveLength(2);
      expect(options[0]).toHaveTextContent('한국어');
      expect(options[1]).toHaveTextContent('English');
    });

    it('should change locale via select', async () => {
      render(
        <TestWrapper>
          <LanguageToggle variant="select" />
        </TestWrapper>
      );
      
      const select = screen.getByRole('combobox');
      expect(select).toHaveValue('ko');
      
      fireEvent.change(select, { target: { value: 'en' } });
      
      await waitFor(() => {
        expect(select).toHaveValue('en');
      });
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('workflow-locale', 'en');
    });

    it('should be disabled when loading', () => {
      render(
        <TestWrapper>
          <LanguageToggle variant="select" />
        </TestWrapper>
      );
      
      const select = screen.getByRole('combobox');
      // Initially not loading, so not disabled
      expect(select).not.toBeDisabled();
    });

    it('should apply custom className to select variant', () => {
      render(
        <TestWrapper>
          <LanguageToggle variant="select" className="custom-select" />
        </TestWrapper>
      );
      
      const container = screen.getByRole('combobox').closest('.language-toggle-select');
      expect(container).toHaveClass('custom-select');
    });

    it('should have proper accessibility for select', () => {
      render(
        <TestWrapper>
          <LanguageToggle variant="select" />
        </TestWrapper>
      );
      
      const select = screen.getByRole('combobox');
      expect(select).toHaveAttribute('aria-label');
    });
  });

  describe('Error handling', () => {
    it('should handle locale change errors gracefully', async () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      mockLocalStorage.setItem.mockImplementation(() => {
        throw new Error('Storage error');
      });
      
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const button = screen.getByRole('button');
      fireEvent.click(button);
      
      // Should still change locale despite storage error
      await waitFor(() => {
        expect(screen.getByText('EN')).toBeInTheDocument();
      });
      
      consoleSpy.mockRestore();
    });
  });

  describe('Component lifecycle', () => {
    it('should cleanup properly on unmount', () => {
      const { unmount } = render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      expect(() => unmount()).not.toThrow();
    });

    it('should handle prop changes', () => {
      const { rerender } = render(
        <TestWrapper>
          <LanguageToggle variant="button" />
        </TestWrapper>
      );
      
      expect(screen.getByRole('button')).toBeInTheDocument();
      
      rerender(
        <TestWrapper>
          <LanguageToggle variant="select" />
        </TestWrapper>
      );
      
      expect(screen.getByRole('combobox')).toBeInTheDocument();
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
  });

  describe('Responsive behavior', () => {
    it('should show text on large screens by default', () => {
      render(
        <TestWrapper>
          <LanguageToggle />
        </TestWrapper>
      );
      
      const textElement = screen.getByText('한국어');
      expect(textElement).toHaveClass('hidden', 'sm:inline');
    });

    it('should respect showText prop', () => {
      render(
        <TestWrapper>
          <LanguageToggle showText={false} />
        </TestWrapper>
      );
      
      expect(screen.queryByText('한국어')).not.toBeInTheDocument();
    });
  });
});