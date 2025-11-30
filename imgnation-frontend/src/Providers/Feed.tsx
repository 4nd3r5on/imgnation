import React, {
  createContext,
  useContext,
  useCallback,
  useState,
  useEffect,
  useRef,
  useMemo,
  ReactNode
} from 'react';
import { useAuth } from './Api.tsx';
import { buildQuery } from "../utils/query_params.ts";

// Types
export interface Variant {
  name: string;
  content_type: string;
  size: number;
  metadata: {
    height: number;
    width: number;
  };
  is_chunked: boolean;
  status: string;
}

export interface Upload {
  user_id: string;
  filename: string;
  access: {
    level: number;
  };
  uploaded_at: string;
  tags: string[];
  description: string;
}

export interface FeedItem {
  id: string;
  size: number;
  variants: Variant[];
  uploads: Upload[];
  reactions: { [key: string]: number };
  created_at: string;
}

export interface SearchParams {
  tags?: string[];
  user_id?: string;
  before_date?: string;
}

export interface SearchResponse {
  items: FeedItem[];
  total: number;
  page: number;
  page_size: number;
  order_asc: boolean;
}

export interface MasonryItem extends FeedItem {
  height: number;
  color: string;
  imageUrl?: string;
  loading?: boolean;
}

interface FeedState {
  items: MasonryItem[];
  total: number;
  loading: boolean;
  error: string | null;
  hasMore: boolean;
  page: number;
  searchParams: SearchParams;
  searchTimestamp: string | null;
}

interface FeedContextValue extends FeedState {
  search: (params?: SearchParams) => Promise<void>;
  loadMore: () => Promise<void>;
  reload: () => Promise<void>;
  clearError: () => void;
  getColumnWidth: (windowsHeight: number, windowsWidth: number) => number;
  refreshItem: (itemId: string) => Promise<void>;
  updateItem: (itemId: string, updates: Partial<FeedItem>) => void;
}

// Constants
const PAGE_SIZE = 20;
const GRADIENT_COLORS = [
  'bg-gradient-to-br from-purple-500 to-pink-500',
  'bg-gradient-to-br from-blue-500 to-cyan-500',
  'bg-gradient-to-br from-green-500 to-emerald-500',
  'bg-gradient-to-br from-orange-500 to-red-500',
  'bg-gradient-to-br from-indigo-500 to-purple-500',
  'bg-gradient-to-br from-teal-500 to-blue-500',
  'bg-gradient-to-br from-rose-500 to-pink-500',
  'bg-gradient-to-br from-amber-500 to-orange-500',
] as const;

// Context
const FeedContext = createContext<FeedContextValue | null>(null);

// Custom hook for using the feed context
export const useFeed = () => {
  const context = useContext(FeedContext);
  if (!context) {
    throw new Error('useFeed must be used within a FeedProvider');
  }
  return context;
};

// Utility functions
const getItemColor = (itemId: string): string => {
  const index = Math.abs(itemId.charCodeAt(0)) % GRADIENT_COLORS.length;
  return GRADIENT_COLORS[index];
};

const areSearchParamsEqual = (prev: SearchParams, current: SearchParams): boolean => {
  if (!prev && !current) return true;
  if (!prev || !current) return false;

  const prevTags = prev.tags?.sort() || [];
  const currentTags = current.tags?.sort() || [];

  return (
    prev.user_id === current.user_id &&
    prev.before_date === current.before_date &&
    prevTags.length === currentTags.length &&
    prevTags.every((tag, index) => tag === currentTags[index])
  );
};

// Image cache for blob URLs
class ImageCache {
  private cache = new Map<string, string>();
  private loadingPromises = new Map<string, Promise<string>>();

  async getImageUrl(key: string, loader: () => Promise<string>): Promise<string> {
    if (this.cache.has(key)) {
      return this.cache.get(key)!;
    }

    if (this.loadingPromises.has(key)) {
      return this.loadingPromises.get(key)!;
    }

    const promise = loader().then(url => {
      this.cache.set(key, url);
      this.loadingPromises.delete(key);
      return url;
    }).catch(error => {
      this.loadingPromises.delete(key);
      throw error;
    });

    this.loadingPromises.set(key, promise);
    return promise;
  }

  clear(): void {
    // Revoke all blob URLs to prevent memory leaks
    for (const url of this.cache.values()) {
      if (url.startsWith('blob:')) {
        URL.revokeObjectURL(url);
      }
    }
    this.cache.clear();
    this.loadingPromises.clear();
  }

  revokeUrl(key: string): void {
    const url = this.cache.get(key);
    if (url && url.startsWith('blob:')) {
      URL.revokeObjectURL(url);
    }
    this.cache.delete(key);
  }
}

// Provider component
interface FeedProviderProps {
  children: ReactNode;
  initialSearchParams?: SearchParams;
}

export const FeedProvider: React.FC<FeedProviderProps> = ({
  children,
  initialSearchParams = {}
}) => {
  const { req } = useAuth();
  const imageCache = useRef(new ImageCache());
  const abortControllerRef = useRef<AbortController | null>(null);
  const [windowDimensions, setWindowDimensions] = useState({
    width: globalThis.innerWidth,
    height: globalThis.innerHeight
  });

  // State management
  const [state, setState] = useState<FeedState>({
    items: [],
    total: 0,
    loading: false,
    error: null,
    hasMore: true,
    page: 1,
    searchParams: initialSearchParams,
    searchTimestamp: null,
  });

  // Window resize handler
  useEffect(() => {
    const handleResize = () => {
      setWindowDimensions({
        width: globalThis.innerWidth,
        height: globalThis.innerHeight
      });
    };

    globalThis.addEventListener("resize", handleResize);
    return () => globalThis.removeEventListener("resize", handleResize);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      imageCache.current.clear();
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  // Memoized column width calculation
  const getColumnWidth = (windowsHeight: number, windowsWidth: number) => {
    return windowsWidth > windowsHeight ? 300 : 200;
  };

  // API functions
  const fetchFeed = useCallback(
    async (pageNum: number, params: SearchParams, reuseTimestamp?: string): Promise<SearchResponse> => {
      const before_date = reuseTimestamp || new Date().toISOString();
      const query = buildQuery({
        ...params,
        before_date,
        page: pageNum,
        page_size: PAGE_SIZE,
        tags: params.tags?.join('+') || undefined,
      });

      // Cancel previous request if still pending
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }

      abortControllerRef.current = new AbortController();

      const response = await req(`/uploads?${query}`, {
        method: 'GET',
        signal: abortControllerRef.current.signal
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return await response.json();
    },
    [req]
  );

  const convertToMasonryItem = useCallback(async (item: FeedItem): Promise<MasonryItem> => {
    const color = getItemColor(item.id);

    const previewVariant =
      item.variants.find(v => v.name === 'thumbnail' && v.content_type.startsWith('image/')) ||
      item.variants.find(v => v.name === 'compressed' && v.content_type.startsWith('image/')) ||
      item.variants.find(v => v.name === '' && v.content_type.startsWith('image/'));

    let imageUrl: string | undefined;
    let height = Math.floor(Math.random() * 200) + getColumnWidth(windowDimensions.height, windowDimensions.width);

    if (previewVariant) {
      try {
        const cacheKey = `${item.id}-${previewVariant.name}`;
        imageUrl = await imageCache.current.getImageUrl(cacheKey, async () => {
          const variantParam = previewVariant.name ? `?variant=${previewVariant.name}` : '';
          const res = await req(`/uploads/${item.id}${variantParam}`, {});
          const blob = await res.blob();
          return URL.createObjectURL(blob);
        });

        if (previewVariant.metadata?.height) {
          height = Math.min(previewVariant.metadata.height, 400) + 100;
        }
      } catch (error) {
        console.warn(`Failed to load image for item ${item.id}:`, error);
        // Keep default height and no imageUrl
      }
    }

    return { ...item, height, color, imageUrl };
  }, [req, getColumnWidth, windowDimensions]);

  const convertToMasonryItems = useCallback(async (items: FeedItem[]): Promise<MasonryItem[]> => {
    const promises = items.map(item => convertToMasonryItem(item));
    return Promise.all(promises);
  }, [convertToMasonryItem]);

  // Main actions
  const search = useCallback(async (params: SearchParams = {}) => {
    // Avoid duplicate searches
    if (areSearchParamsEqual(state.searchParams, params) && state.items.length > 0 && !state.error) {
      return;
    }

    setState(prev => ({
      ...prev,
      items: [],
      total: 0,
      hasMore: true,
      page: 1,
      error: null,
      loading: true,
      searchParams: params,
    }));

    try {
      const timestamp = new Date().toISOString();
      const response = await fetchFeed(1, params, timestamp);
      const masonryItems = await convertToMasonryItems(response.items);

      setState(prev => ({
        ...prev,
        items: masonryItems,
        total: response.total,
        searchTimestamp: timestamp,
        hasMore: response.items.length >= PAGE_SIZE && response.items.length < response.total,
        loading: false,
      }));
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        return; // Request was cancelled, don't update state
      }

      const errorMessage = error instanceof Error ? error.message : 'Search failed';
      console.error('Search failed:', errorMessage, { params, error });

      setState(prev => ({
        ...prev,
        error: errorMessage,
        loading: false,
      }));
    }
  }, [state.searchParams, state.items.length, state.error, fetchFeed, convertToMasonryItems]);

  const loadMore = useCallback(async () => {
    if (state.loading || !state.hasMore) return;

    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const nextPage = state.page + 1;
      const response = await fetchFeed(nextPage, state.searchParams, state.searchTimestamp || undefined);
      const newMasonryItems = await convertToMasonryItems(response.items);
      const allItems = [...state.items, ...newMasonryItems];

      setState(prev => ({
        ...prev,
        items: allItems,
        page: nextPage,
        hasMore: response.items.length >= PAGE_SIZE && allItems.length < response.total,
        loading: false,
      }));
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        return;
      }

      const errorMessage = error instanceof Error ? error.message : 'Load more failed';
      setState(prev => ({
        ...prev,
        error: errorMessage,
        loading: false,
      }));
    }
  }, [state.loading, state.hasMore, state.page, state.searchParams, state.searchTimestamp, state.items, fetchFeed, convertToMasonryItems]);

  const reload = useCallback(async () => {
    await search(state.searchParams);
  }, [search, state.searchParams]);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  const refreshItem = useCallback(async (itemId: string) => {
    const itemIndex = state.items.findIndex(item => item.id === itemId);
    if (itemIndex === -1) return;

    try {
      // Mark item as loading
      setState(prev => ({
        ...prev,
        items: prev.items.map((item, index) =>
          index === itemIndex ? { ...item, loading: true } : item
        )
      }));

      // Clear cached image
      imageCache.current.revokeUrl(`${itemId}-thumbnail`);
      imageCache.current.revokeUrl(`${itemId}-compressed`);
      imageCache.current.revokeUrl(`${itemId}-`);

      // Refresh the item
      const originalItem = state.items[itemIndex];
      const refreshedItem = await convertToMasonryItem(originalItem);

      setState(prev => ({
        ...prev,
        items: prev.items.map((item, index) =>
          index === itemIndex ? { ...refreshedItem, loading: false } : item
        )
      }));
    } catch (error) {
      console.error(`Failed to refresh item ${itemId}:`, error);
      setState(prev => ({
        ...prev,
        items: prev.items.map((item, index) =>
          index === itemIndex ? { ...item, loading: false } : item
        )
      }));
    }
  }, [state.items, convertToMasonryItem]);

  const updateItem = useCallback((itemId: string, updates: Partial<FeedItem>) => {
    setState(prev => ({
      ...prev,
      items: prev.items.map(item =>
        item.id === itemId ? { ...item, ...updates } : item
      )
    }));
  }, []);

  // Context value
  const contextValue = useMemo<FeedContextValue>(() => ({
    ...state,
    search,
    loadMore,
    reload,
    clearError,
    getColumnWidth,
    refreshItem,
    updateItem,
  }), [state, search, loadMore, reload, clearError, getColumnWidth, refreshItem, updateItem]);

  return (
    <FeedContext.Provider value={contextValue}>
      {children}
    </FeedContext.Provider>
  );
};