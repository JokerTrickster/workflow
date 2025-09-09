/**
 * Unit tests for useNetworkStatus hook
 * Tests network connectivity detection and status changes
 */

import { renderHook, act } from '@testing-library/react';
import { useNetworkStatus } from '../../hooks/useNetworkStatus';

// Mock navigator.onLine
Object.defineProperty(navigator, 'onLine', {
  writable: true,
  value: true,
});

// Mock window event listeners
const mockAddEventListener = jest.fn();
const mockRemoveEventListener = jest.fn();

Object.defineProperty(window, 'addEventListener', {
  writable: true,
  value: mockAddEventListener,
});

Object.defineProperty(window, 'removeEventListener', {
  writable: true,
  value: mockRemoveEventListener,
});

describe('useNetworkStatus Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset navigator.onLine to true before each test
    Object.defineProperty(navigator, 'onLine', {
      writable: true,
      value: true,
    });
  });

  describe('Initial State', () => {
    it('should return true when navigator.onLine is true', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true,
      });

      const { result } = renderHook(() => useNetworkStatus());
      expect(result.current).toBe(true);
    });

    it('should return false when navigator.onLine is false', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: false,
      });

      const { result } = renderHook(() => useNetworkStatus());
      expect(result.current).toBe(false);
    });

    it('should return true in server-side rendering environment', () => {
      // Mock window as undefined to simulate SSR
      const originalWindow = global.window;
      delete (global as any).window;

      const { result } = renderHook(() => useNetworkStatus());
      expect(result.current).toBe(true);

      // Restore window
      global.window = originalWindow;
    });
  });

  describe('Event Listeners', () => {
    it('should register online and offline event listeners on mount', () => {
      renderHook(() => useNetworkStatus());

      expect(mockAddEventListener).toHaveBeenCalledWith('online', expect.any(Function));
      expect(mockAddEventListener).toHaveBeenCalledWith('offline', expect.any(Function));
      expect(mockAddEventListener).toHaveBeenCalledTimes(2);
    });

    it('should clean up event listeners on unmount', () => {
      const { unmount } = renderHook(() => useNetworkStatus());

      unmount();

      expect(mockRemoveEventListener).toHaveBeenCalledWith('online', expect.any(Function));
      expect(mockRemoveEventListener).toHaveBeenCalledWith('offline', expect.any(Function));
      expect(mockRemoveEventListener).toHaveBeenCalledTimes(2);
    });
  });

  describe('Network Status Changes', () => {
    it('should update status when going online', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: false,
      });

      const { result } = renderHook(() => useNetworkStatus());

      // Initially offline
      expect(result.current).toBe(false);

      // Simulate going online
      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: true,
        });

        // Get the online event handler and call it
        const onlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'online'
        )[1];
        onlineHandler();
      });

      expect(result.current).toBe(true);
    });

    it('should update status when going offline', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true,
      });

      const { result } = renderHook(() => useNetworkStatus());

      // Initially online
      expect(result.current).toBe(true);

      // Simulate going offline
      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: false,
        });

        // Get the offline event handler and call it
        const offlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'offline'
        )[1];
        offlineHandler();
      });

      expect(result.current).toBe(false);
    });

    it('should handle rapid network status changes', () => {
      const { result } = renderHook(() => useNetworkStatus());

      // Initially online
      expect(result.current).toBe(true);

      // Go offline
      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: false,
        });
        const offlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'offline'
        )[1];
        offlineHandler();
      });

      expect(result.current).toBe(false);

      // Go online
      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: true,
        });
        const onlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'online'
        )[1];
        onlineHandler();
      });

      expect(result.current).toBe(true);

      // Go offline again
      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: false,
        });
        const offlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'offline'
        )[1];
        offlineHandler();
      });

      expect(result.current).toBe(false);
    });
  });

  describe('Console Logging', () => {
    let consoleSpy: jest.SpyInstance;

    beforeEach(() => {
      consoleSpy = jest.spyOn(console, 'log').mockImplementation();
    });

    afterEach(() => {
      consoleSpy.mockRestore();
    });

    it('should log connection restored message when going online', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: false,
      });

      renderHook(() => useNetworkStatus());

      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: true,
        });
        const onlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'online'
        )[1];
        onlineHandler();
      });

      expect(consoleSpy).toHaveBeenCalledWith('Connection restored');
    });

    it('should log connection lost message when going offline', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true,
      });

      renderHook(() => useNetworkStatus());

      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: false,
        });
        const offlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'offline'
        )[1];
        offlineHandler();
      });

      expect(consoleSpy).toHaveBeenCalledWith('Connection lost');
    });
  });

  describe('Multiple Hook Instances', () => {
    it('should maintain independent state for multiple hook instances', () => {
      const { result: result1 } = renderHook(() => useNetworkStatus());
      const { result: result2 } = renderHook(() => useNetworkStatus());

      expect(result1.current).toBe(result2.current);

      // Both should update when network status changes
      act(() => {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: false,
        });
        const offlineHandler = mockAddEventListener.mock.calls.find(
          call => call[0] === 'offline'
        )[1];
        offlineHandler();
      });

      expect(result1.current).toBe(false);
      expect(result2.current).toBe(false);
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing navigator object', () => {
      const originalNavigator = global.navigator;
      delete (global as any).navigator;

      const { result } = renderHook(() => useNetworkStatus());
      
      // Should fallback to true when navigator is not available
      expect(result.current).toBe(true);

      // Restore navigator
      global.navigator = originalNavigator;
    });

    it('should handle navigator.onLine being undefined', () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: undefined,
      });

      const { result } = renderHook(() => useNetworkStatus());
      
      // Should fallback to true when onLine is undefined
      expect(result.current).toBe(true);
    });

    it('should not register event listeners in server environment', () => {
      const originalWindow = global.window;
      delete (global as any).window;

      renderHook(() => useNetworkStatus());

      // Should not attempt to register event listeners when window is undefined
      expect(mockAddEventListener).not.toHaveBeenCalled();

      // Restore window
      global.window = originalWindow;
    });
  });
});