import { useCallback, useState, useEffect } from 'react';
import { FeedItem, SearchParams, SearchResponse } from './feedTypes.ts';
import { useAuth } from '../Providers/Api.tsx';
import {buildQuery} from "../utils/query_params.ts";

const PAGE_SIZE = 20;

export interface UseFeedReturn {
  items: MasonryItem[];
  total: number;
  loading: boolean;
  error: string | null;
  hasMore: boolean;
  search: (params?: SearchParams) => Promise<void>;
  loadMore: () => Promise<void>;
  reload: () => Promise<void>;
  clearError: () => void;
}

export interface MasonryItem extends FeedItem {
  height: number;
  color: string;
  imageUrl?: string;
}

export const useFeed = (): UseFeedReturn => {
  const { req } = useAuth();

  // Split state into individual pieces
  const [items, setItems] = useState<MasonryItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [page, setPage] = useState(1);
  const [searchTimestamp, setSearchTimestamp] = useState<string | null>(null);
  const [searchParams, setSearchParams] = useState<SearchParams>({});
  const [prevSearchParams, setPrevSearchParams] = useState<SearchParams>({});
  const [windowWidth, setWindowWidth] = useState(globalThis.innerWidth);
  const [windowHeight, setWindowHeight] = useState(globalThis.innerHeight);

  useEffect(() => {
    const resize = () => {
      setWindowWidth(globalThis.innerWidth);
      setWindowHeight(globalThis.innerHeight);
    };
    globalThis.addEventListener("resize", resize);
    return () => globalThis.removeEventListener("resize", resize);
  }, []);

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
      const response = await req(`/uploads?${query}`, { method: 'GET' });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      return await response.json();
    },
    [req, buildQuery]
  );

  const getColumnWidth = useCallback(() =>
    windowWidth > windowHeight ? 300 : 200, [windowWidth, windowHeight]);

  const convertToMasonryItem = useCallback(async (item: FeedItem): Promise<MasonryItem> => {
    const colors = [
      'bg-gradient-to-br from-purple-500 to-pink-500',
      'bg-gradient-to-br from-blue-500 to-cyan-500',
      'bg-gradient-to-br from-green-500 to-emerald-500',
      'bg-gradient-to-br from-orange-500 to-red-500',
    ];
    const color = colors[Math.abs(item.id.charCodeAt(0)) % colors.length];

    const previewVariant =
      item.variants.find(v => v.name === 'thumbnail' && v.content_type.startsWith('image/')) ||
      item.variants.find(v => v.name === 'compressed' && v.content_type.startsWith('image/')) ||
      item.variants.find(v => v.name === '' && v.content_type.startsWith('image/'))

    const variantParam = previewVariant?.name ? `?variant=${previewVariant.name}` : '';
    const res = await req(`/uploads/${item.id}${variantParam}`, {})
    const blob = await res.blob()
    const imageUrl = URL.createObjectURL(blob)

    const height= previewVariant?.metadata?.height
      ? Math.min(previewVariant.metadata.height, 400) + 100
      : Math.floor(Math.random() * 200) + getColumnWidth();
    return { ...item, height, color, imageUrl };
  }, [getColumnWidth]);

  const convertToMasonryItems = async (items: FeedItem[]): Promise<MasonryItem[]> => {
    const convertedItems: MasonryItem[] = []
    for (const item of items) {
      convertedItems.push(await convertToMasonryItem(item))
    }
    return convertedItems;
  }

  const loadMore = useCallback(async () => {
    if (loading || !hasMore) return;

    setLoading(true);
    setError(null);

    try {
      const nextPage = page + 1;
      const res = await fetchFeed(nextPage, searchParams, searchTimestamp || undefined);
      const newItems: MasonryItem[] = [...items, ...(await convertToMasonryItems(res.items))];

      setItems(newItems);
      setPage(nextPage);
      setHasMore(res.items.length >= PAGE_SIZE && newItems.length < res.total);
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Load more failed');
      setLoading(false);
    }
  }, [fetchFeed, loading, hasMore, page, searchParams, searchTimestamp, items]);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Helper function to deep compare search parameters
  const areSearchParamsEqual = (prev: SearchParams, current: SearchParams): boolean => {
    if (!prev && !current) return true;
    if (!prev || !current) return false;
    const prevTags = prev.tags || [];
    const currentTags = current.tags || [];

    if (prevTags.length !== currentTags.length) return false;
    return prevTags.every((tag, index) => tag === currentTags[index]);
  };

  const search = useCallback(async (params: SearchParams = {}) => {
    if (areSearchParamsEqual(prevSearchParams, params) && items.length > 0) return

    // Clean up current results
    setItems([]);
    setTotal(0);
    setHasMore(true);
    setPage(1);
    setError(null);

    // Update search params immediately
    setSearchParams(params);
    setLoading(true);

    try {
      const timestamp = new Date().toISOString();
      const res = await fetchFeed(1, params, timestamp);

      setItems(await convertToMasonryItems(res.items));
      setTotal(res.total);
      setSearchTimestamp(timestamp);
      setHasMore(res.items.length >= PAGE_SIZE && res.items.length < res.total);
      setLoading(false);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Search error';
      console.error('Search failed:', errorMessage, { params, error: err });
      setError(errorMessage);
      setLoading(false);
    }

    setPrevSearchParams(params);
  }, [fetchFeed]);

  const reload = useCallback(async () => {
    await search(searchParams);
  }, [search, searchParams]);

  return {
    items,
    total,
    loading,
    error,
    hasMore,
    search,
    loadMore,
    reload,
    clearError,
  };
};