import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useFeed, SearchParams, MasonryItem } from '../Providers/Feed.tsx';

// Hook for infinite scroll functionality
export const useInfiniteScroll = (threshold = 200) => {
  const { loadMore, hasMore, loading } = useFeed();
  const observerRef = useRef<IntersectionObserver>(null);
  const loadingRef = useRef(loading);

  // Keep loading state in sync
  useEffect(() => {
    loadingRef.current = loading;
  }, [loading]);

  const lastElementRef = useCallback((node: HTMLElement | null) => {
    if (loadingRef.current) return;

    if (observerRef.current) {
      observerRef.current.disconnect();
    }

    if (node && hasMore) {
      observerRef.current = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting && !loadingRef.current && hasMore) {
            loadMore();
          }
        },
        { rootMargin: `${threshold}px` }
      );
      observerRef.current.observe(node);
    }
  }, [hasMore, loadMore, threshold]);

  useEffect(() => {
    return () => {
      if (observerRef.current) {
        observerRef.current.disconnect();
      }
    };
  }, []);

  return { lastElementRef };
};

// Hook for search functionality with debouncing
export const useSearch = (debounceMs = 300) => {
  const { search, loading, error } = useFeed();
  const timeoutRef = useRef<number>(null);

  const debouncedSearch = useCallback((params: SearchParams) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(() => {
      search(params);
    }, debounceMs);
  }, [search, debounceMs]);

  const immediateSearch = useCallback((params: SearchParams) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    search(params);
  }, [search]);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return {
    debouncedSearch,
    immediateSearch,
    loading,
    error
  };
};

// Hook for masonry layout calculations
export const useMasonryLayout = (columnCount: number = 3, gap: number = 16, columnWidth: number = 200) => {

};

// Hook for performance monitoring
export const useFeedStats = () => {
  const { items, total, loading, error } = useFeed();
  const loadStartTime = useRef<number>(0);
  const [loadTime, setLoadTime] = useState<number>(0);
  const [loadHistory, setLoadHistory] = useState<number[]>([]);

  useEffect(() => {
    if (loading) {
      loadStartTime.current = Date.now();
    } else if (loadStartTime.current > 0) {
      const duration = Date.now() - loadStartTime.current;
      setLoadTime(duration);
      setLoadHistory(prev => [...prev.slice(-9), duration]); // Keep last 10 loads
    }
  }, [loading]);

  const averageLoadTime = useMemo(() => {
    if (loadHistory.length === 0) return 0;
    return loadHistory.reduce((sum, time) => sum + time, 0) / loadHistory.length;
  }, [loadHistory]);

  const itemsPerSecond = useMemo(() => {
    if (loadTime === 0 || items.length === 0) return 0;
    return Math.round((items.length / loadTime) * 1000);
  }, [loadTime, items.length]);

  return {
    itemCount: items.length,
    totalItems: total,
    loadTime,
    averageLoadTime: Math.round(averageLoadTime),
    itemsPerSecond,
    hasError: !!error,
    loadingProgress: total > 0 ? (items.length / total) * 100 : 0,
    performanceGrade: averageLoadTime < 500 ? 'A' : averageLoadTime < 1000 ? 'B' : averageLoadTime < 2000 ? 'C' : 'D'
  };
};

// Hook for virtualization (for large datasets)
export const useVirtualization = (containerHeight: number, itemHeight: number = 200) => {
  const { items } = useFeed();
  const [scrollTop, setScrollTop] = useState(0);

  const visibleRange = useMemo(() => {
    const start = Math.floor(scrollTop / itemHeight);
    const visibleCount = Math.ceil(containerHeight / itemHeight) + 2; // Buffer
    const end = Math.min(start + visibleCount, items.length);

    return { start, end };
  }, [scrollTop, itemHeight, containerHeight, items.length]);

  const visibleItems = useMemo(() => {
    return items.slice(visibleRange.start, visibleRange.end).map((item, index) => ({
      ...item,
      virtualIndex: visibleRange.start + index,
      offsetY: (visibleRange.start + index) * itemHeight
    }));
  }, [items, visibleRange, itemHeight]);

  const totalHeight = items.length * itemHeight;

  const handleScroll = useCallback((event: Event) => {
    const target = event.target as HTMLElement;
    setScrollTop(target.scrollTop);
  }, []);

  return {
    visibleItems,
    totalHeight,
    handleScroll,
    visibleRange
  };
};

// Hook for offline support
export const useOfflineSupport = () => {
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [pendingActions, setPendingActions] = useState<Array<{
    id: string;
    action: string;
    data: unknown;
    timestamp: number;
  }>>([]);

  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    globalThis.addEventListener('online', handleOnline);
    globalThis.addEventListener('offline', handleOffline);

    return () => {
      globalThis.removeEventListener('online', handleOnline);
      globalThis.removeEventListener('offline', handleOffline);
    };
  }, []);

  const queueAction = useCallback((action: string, data: unknown) => {
    if (!isOnline) {
      const actionItem = {
        id: crypto.randomUUID(),
        action,
        data,
        timestamp: Date.now()
      };
      setPendingActions(prev => [...prev, actionItem]);
      return false; // Action queued
    }
    return true; // Execute immediately
  }, [isOnline]);

  const processPendingActions = useCallback(() => {
    if (!isOnline || pendingActions.length === 0) return;

    // Process pending actions when back online
    for (const action of pendingActions) {
      try {
        // Process the action
        console.log('Processing pending action:', action);
        // await executeAction(action);
      } catch (error) {
        console.error('Failed to process pending action:', error);
      }
    }

    setPendingActions([]);
  }, [isOnline, pendingActions]);

  useEffect(() => {
    if (isOnline) {
      processPendingActions();
    }
  }, [isOnline, processPendingActions]);

  return {
    isOnline,
    pendingActions,
    queueAction,
    hasPendingActions: pendingActions.length > 0
  };
};