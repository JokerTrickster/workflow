'use client';

import { useCallback, useEffect, useRef, useState } from 'react';

/**
 * Configuration options for the IntersectionObserver
 */
export interface UseIntersectionObserverOptions {
  /** The element that acts as the viewport for checking visibility */
  root?: Element | null;
  /** Margin around the root. Can have values similar to the CSS margin property */
  rootMargin?: string;
  /** Either a single number or an array of numbers between 0.0 and 1.0 */
  threshold?: number | number[];
  /** Whether the observer should be enabled */
  enabled?: boolean;
  /** Delay in ms before triggering the callback (for performance optimization) */
  delay?: number;
  /** Whether to trigger only once when element becomes visible */
  triggerOnce?: boolean;
}

/**
 * Return type for the useIntersectionObserver hook
 */
export interface UseIntersectionObserverReturn {
  /** Ref to attach to the target element */
  targetRef: (node: Element | null) => void;
  /** Whether the target element is currently intersecting */
  isIntersecting: boolean;
  /** The intersection ratio (0.0 to 1.0) */
  intersectionRatio: number;
  /** The IntersectionObserverEntry object if available */
  entry: IntersectionObserverEntry | null;
  /** Manually disconnect the observer */
  disconnect: () => void;
  /** Reconnect the observer if it was disconnected */
  reconnect: () => void;
  /** Whether the observer is currently active */
  isObserving: boolean;
}

/**
 * Custom hook for observing element intersection with viewport
 * Optimized for infinite scrolling and large lists
 * 
 * @param options - Configuration options for the observer
 * @param callback - Optional callback function to execute when intersection changes
 * @returns Object with target ref, intersection state, and control methods
 */
export function useIntersectionObserver(
  options: UseIntersectionObserverOptions = {},
  callback?: (entry: IntersectionObserverEntry) => void
): UseIntersectionObserverReturn {
  const {
    root = null,
    rootMargin = '0px',
    threshold = 0,
    enabled = true,
    delay = 0,
    triggerOnce = false,
  } = options;

  // State for intersection data
  const [isIntersecting, setIsIntersecting] = useState(false);
  const [intersectionRatio, setIntersectionRatio] = useState(0);
  const [entry, setEntry] = useState<IntersectionObserverEntry | null>(null);
  const [isObserving, setIsObserving] = useState(false);
  const [hasTriggered, setHasTriggered] = useState(false);

  // Refs for managing the observer and target element
  const observerRef = useRef<IntersectionObserver | null>(null);
  const targetElementRef = useRef<Element | null>(null);
  const delayTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Cleanup function for delay timeout
  const clearDelayTimeout = useCallback(() => {
    if (delayTimeoutRef.current) {
      clearTimeout(delayTimeoutRef.current);
      delayTimeoutRef.current = null;
    }
  }, []);

  // Handle intersection changes
  const handleIntersection = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [intersectionEntry] = entries;
      if (!intersectionEntry) return;

      // Clear any existing delay timeout
      clearDelayTimeout();

      const processIntersection = () => {
        setEntry(intersectionEntry);
        setIsIntersecting(intersectionEntry.isIntersecting);
        setIntersectionRatio(intersectionEntry.intersectionRatio);

        // Call the optional callback
        if (callback) {
          callback(intersectionEntry);
        }

        // If triggerOnce is enabled and element is intersecting, mark as triggered
        if (triggerOnce && intersectionEntry.isIntersecting && !hasTriggered) {
          setHasTriggered(true);
        }
      };

      // Apply delay if specified
      if (delay > 0) {
        delayTimeoutRef.current = setTimeout(processIntersection, delay);
      } else {
        processIntersection();
      }
    },
    [callback, delay, triggerOnce, hasTriggered, clearDelayTimeout]
  );

  // Create the IntersectionObserver
  const createObserver = useCallback(() => {
    // Don't create observer in SSR or if not enabled
    if (typeof window === 'undefined' || !enabled) return null;

    // If triggerOnce is enabled and already triggered, don't create observer
    if (triggerOnce && hasTriggered) return null;

    try {
      const observer = new IntersectionObserver(handleIntersection, {
        root,
        rootMargin,
        threshold,
      });

      console.log('🔍 IntersectionObserver created with options:', {
        root: root ? 'custom' : 'viewport',
        rootMargin,
        threshold,
        enabled,
        delay,
        triggerOnce,
      });

      return observer;
    } catch (error) {
      console.error('❌ Failed to create IntersectionObserver:', error);
      return null;
    }
  }, [root, rootMargin, threshold, enabled, handleIntersection, triggerOnce, hasTriggered, delay]);

  // Disconnect the observer
  const disconnect = useCallback(() => {
    clearDelayTimeout();
    if (observerRef.current) {
      observerRef.current.disconnect();
      observerRef.current = null;
      setIsObserving(false);
      console.log('🔌 IntersectionObserver disconnected');
    }
  }, [clearDelayTimeout]);

  // Reconnect the observer
  const reconnect = useCallback(() => {
    disconnect();
    
    // Reset triggered state if reconnecting
    if (triggerOnce) {
      setHasTriggered(false);
    }
    
    const newObserver = createObserver();
    if (newObserver && targetElementRef.current) {
      observerRef.current = newObserver;
      newObserver.observe(targetElementRef.current);
      setIsObserving(true);
      console.log('🔄 IntersectionObserver reconnected');
    }
  }, [disconnect, createObserver, triggerOnce]);

  // Target ref callback for attaching to DOM elements
  const targetRef = useCallback(
    (node: Element | null) => {
      // Disconnect existing observer
      if (observerRef.current) {
        observerRef.current.disconnect();
        setIsObserving(false);
      }

      // Clear any existing delay timeout
      clearDelayTimeout();

      // Store the target element reference
      targetElementRef.current = node;

      // If no node or not enabled, don't observe
      if (!node || !enabled) {
        return;
      }

      // If triggerOnce is enabled and already triggered, don't observe
      if (triggerOnce && hasTriggered) {
        return;
      }

      // Create and start observing
      const observer = createObserver();
      if (observer) {
        observerRef.current = observer;
        observer.observe(node);
        setIsObserving(true);
        console.log('👀 Started observing element for intersection');
      }
    },
    [enabled, createObserver, triggerOnce, hasTriggered, clearDelayTimeout]
  );

  // Cleanup effect
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  // Effect to handle observer recreation when dependencies change
  useEffect(() => {
    if (targetElementRef.current && enabled && !(triggerOnce && hasTriggered)) {
      // Recreate observer with new options
      targetRef(targetElementRef.current);
    } else if (!enabled || (triggerOnce && hasTriggered)) {
      // Disconnect if disabled or already triggered
      disconnect();
    }
  }, [root, rootMargin, threshold, enabled, triggerOnce, hasTriggered, targetRef, disconnect]);

  return {
    targetRef,
    isIntersecting,
    intersectionRatio,
    entry,
    disconnect,
    reconnect,
    isObserving,
  };
}

/**
 * Specialized hook for infinite scrolling scenarios
 * Pre-configured with optimal settings for loading more content
 * 
 * @param onLoadMore - Callback function to execute when more content should be loaded
 * @param options - Optional configuration to override defaults
 * @returns Same interface as useIntersectionObserver
 */
export function useInfiniteScroll(
  onLoadMore: () => void,
  options: Partial<UseIntersectionObserverOptions> = {}
): UseIntersectionObserverReturn {
  const infiniteScrollOptions: UseIntersectionObserverOptions = {
    rootMargin: '100px', // Load content 100px before element comes into view
    threshold: 0.1, // Trigger when 10% of element is visible
    delay: 250, // Debounce to prevent rapid firing
    ...options,
  };

  const handleIntersection = useCallback(
    (entry: IntersectionObserverEntry) => {
      if (entry.isIntersecting) {
        console.log('📜 Infinite scroll triggered - loading more content');
        onLoadMore();
      }
    },
    [onLoadMore]
  );

  return useIntersectionObserver(infiniteScrollOptions, handleIntersection);
}

/**
 * Specialized hook for lazy loading elements
 * Pre-configured for one-time visibility detection
 * 
 * @param onVisible - Callback function to execute when element becomes visible
 * @param options - Optional configuration to override defaults
 * @returns Same interface as useIntersectionObserver
 */
export function useLazyLoad(
  onVisible: () => void,
  options: Partial<UseIntersectionObserverOptions> = {}
): UseIntersectionObserverReturn {
  const lazyLoadOptions: UseIntersectionObserverOptions = {
    rootMargin: '50px', // Start loading 50px before element comes into view
    threshold: 0, // Trigger as soon as element enters viewport
    triggerOnce: true, // Only trigger once
    delay: 100, // Small delay to batch rapid scroll events
    ...options,
  };

  const handleVisibility = useCallback(
    (entry: IntersectionObserverEntry) => {
      if (entry.isIntersecting) {
        console.log('👁️ Element became visible - triggering lazy load');
        onVisible();
      }
    },
    [onVisible]
  );

  return useIntersectionObserver(lazyLoadOptions, handleVisibility);
}