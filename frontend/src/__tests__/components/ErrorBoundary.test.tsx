/**
 * Unit tests for ErrorBoundary component
 * Tests error catching, fallback UI, and recovery mechanisms
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ErrorBoundary } from '../../components/ErrorBoundary';

// Test component that throws errors
const ThrowError = ({ shouldError }: { shouldError: boolean }) => {
  if (shouldError) {
    throw new Error('Test error message');
  }
  return <div>Component rendered successfully</div>;
};

// Mock console.error to avoid noise in tests
const originalError = console.error;
beforeAll(() => {
  console.error = jest.fn();
});

afterAll(() => {
  console.error = originalError;
});

describe('ErrorBoundary Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Normal Operation', () => {
    it('should render children when no error occurs', () => {
      render(
        <ErrorBoundary level="component">
          <ThrowError shouldError={false} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Component rendered successfully')).toBeInTheDocument();
    });

    it('should not show error UI when children render normally', () => {
      render(
        <ErrorBoundary level="component">
          <div>Normal content</div>
        </ErrorBoundary>
      );

      expect(screen.getByText('Normal content')).toBeInTheDocument();
      expect(screen.queryByText(/Something went wrong/)).not.toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should catch errors and show default fallback UI', () => {
      render(
        <ErrorBoundary level="component">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/encountered an unexpected error/)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Try Again/i })).toBeInTheDocument();
    });

    it('should show custom fallback when provided', () => {
      const customFallback = (error: Error, resetError: () => void) => (
        <div>
          <h2>Custom Error UI</h2>
          <p>Error: {error.message}</p>
          <button onClick={resetError}>Custom Reset</button>
        </div>
      );

      render(
        <ErrorBoundary level="component" fallback={customFallback}>
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Custom Error UI')).toBeInTheDocument();
      expect(screen.getByText('Error: Test error message')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Custom Reset/i })).toBeInTheDocument();
    });

    it('should show error details in development mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'development';

      render(
        <ErrorBoundary level="component" showDetails={true}>
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Error Details')).toBeInTheDocument();
      expect(screen.getByText('Test error message')).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });

    it('should hide error details in production mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'production';

      render(
        <ErrorBoundary level="component" showDetails={false}>
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.queryByText('Error Details')).not.toBeInTheDocument();
      expect(screen.queryByText('Test error message')).not.toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });
  });

  describe('Error Recovery', () => {
    it('should reset error state when reset button is clicked', () => {
      const { rerender } = render(
        <ErrorBoundary level="component">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      // Error should be displayed
      expect(screen.getByText('Something went wrong')).toBeInTheDocument();

      // Click reset button
      const resetButton = screen.getByRole('button', { name: /Try Again/i });
      fireEvent.click(resetButton);

      // Re-render with non-erroring component
      rerender(
        <ErrorBoundary level="component">
          <ThrowError shouldError={false} />
        </ErrorBoundary>
      );

      // Should show normal content
      expect(screen.getByText('Component rendered successfully')).toBeInTheDocument();
      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
    });

    it('should call onError callback when error occurs', () => {
      const onErrorSpy = jest.fn();

      render(
        <ErrorBoundary level="component" onError={onErrorSpy}>
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(onErrorSpy).toHaveBeenCalledWith(
        expect.any(Error),
        expect.objectContaining({
          componentStack: expect.any(String),
        })
      );
    });

    it('should handle multiple error recovery cycles', () => {
      let shouldError = true;
      const TestComponent = () => <ThrowError shouldError={shouldError} />;

      const { rerender } = render(
        <ErrorBoundary level="component">
          <TestComponent />
        </ErrorBoundary>
      );

      // First error
      expect(screen.getByText('Something went wrong')).toBeInTheDocument();

      // Reset first error
      fireEvent.click(screen.getByRole('button', { name: /Try Again/i }));
      shouldError = false;
      rerender(
        <ErrorBoundary level="component">
          <TestComponent />
        </ErrorBoundary>
      );

      expect(screen.getByText('Component rendered successfully')).toBeInTheDocument();

      // Cause second error
      shouldError = true;
      rerender(
        <ErrorBoundary level="component">
          <TestComponent />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });
  });

  describe('Different Error Levels', () => {
    it('should handle component-level errors appropriately', () => {
      render(
        <ErrorBoundary level="component">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/component encountered/)).toBeInTheDocument();
    });

    it('should handle page-level errors with appropriate messaging', () => {
      render(
        <ErrorBoundary level="page">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/page encountered/)).toBeInTheDocument();
    });

    it('should handle application-level errors with comprehensive messaging', () => {
      render(
        <ErrorBoundary level="application">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/application encountered/)).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle errors with no message', () => {
      const ThrowEmptyError = () => {
        throw new Error('');
      };

      render(
        <ErrorBoundary level="component">
          <ThrowEmptyError />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Try Again/i })).toBeInTheDocument();
    });

    it('should handle non-Error objects being thrown', () => {
      const ThrowString = () => {
        throw 'String error'; // eslint-disable-line no-throw-literal
      };

      render(
        <ErrorBoundary level="component">
          <ThrowString />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });

    it('should handle nested ErrorBoundaries correctly', () => {
      render(
        <ErrorBoundary level="application">
          <div>Outer content</div>
          <ErrorBoundary level="component">
            <ThrowError shouldError={true} />
          </ErrorBoundary>
          <div>More outer content</div>
        </ErrorBoundary>
      );

      // Only the inner boundary should catch the error
      expect(screen.getByText('Outer content')).toBeInTheDocument();
      expect(screen.getByText('More outer content')).toBeInTheDocument();
      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/component encountered/)).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should provide proper ARIA attributes for error state', () => {
      render(
        <ErrorBoundary level="component">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      const errorContainer = screen.getByRole('alert');
      expect(errorContainer).toBeInTheDocument();
      expect(errorContainer).toHaveAttribute('aria-live', 'polite');
    });

    it('should provide accessible button labels', () => {
      render(
        <ErrorBoundary level="component">
          <ThrowError shouldError={true} />
        </ErrorBoundary>
      );

      const resetButton = screen.getByRole('button', { name: /Try Again/i });
      expect(resetButton).toBeInTheDocument();
      expect(resetButton).toBeEnabled();
    });
  });
});