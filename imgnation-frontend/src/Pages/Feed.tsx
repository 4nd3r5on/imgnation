import { FC, useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router';
import { Plus } from 'lucide-react';
import { useFeed, MasonryItem } from '../Providers/Feed.tsx';
import { useInfiniteScroll, useMasonryLayout } from '../hooks/feed.ts';
import {Header} from "../Components/Header.tsx";
import {buildQuery} from "../utils/query_params.ts";

// Individual Feed Item Component
interface MasonryFeedItemProps {
  item: MasonryItem;
  isLast?: boolean;
  lastElementRef?: (node: HTMLElement | null) => void;
}

export const MasonryFeedItem: FC<MasonryFeedItemProps> = ({
  item,
  isLast,
  lastElementRef
}) => {
  const navigate = useNavigate();

  const getDisplayName = () =>
    item.uploads.length > 0 ? item.uploads[0].filename : `Item ${item.id.slice(0, 8)}`;

  const formatDate = (dateString: string) =>
    new Date(dateString).toLocaleDateString();

  const getTags = () =>
    item.uploads.length > 0 ? item.uploads[0].tags : [];

  return (
    <div
      ref={isLast ? lastElementRef : undefined}
      onClick={() => navigate('/uploads?id=' + item.id)}
      className={`${
        item.imageUrl ? 'bg-gray-200' : item.color
      } rounded-xl shadow-lg hover:scale-[1.02] transition-transform cursor-pointer relative overflow-hidden`}
      style={{ minHeight: item.height }}
    >
      {item.imageUrl ? (
        <img
          src={item.imageUrl}
          alt={getDisplayName()}
          className="w-full h-full object-cover"
          loading="lazy"
        />
      ) : (
        <div className={`w-full h-full ${item.color}`}></div>
      )}

      <div className="absolute inset-0 bg-black/50 opacity-0 hover:opacity-100 transition-opacity flex flex-col justify-end p-4">
        <div className="text-white">
          <h3 className="font-semibold text-lg mb-1">{getDisplayName()}</h3>
          <p className="text-white/80 text-sm mb-2">{formatDate(item.created_at)}</p>
          <p className="text-white/70 text-xs mb-2">
            Size: {(item.size / 1024 / 1024).toFixed(2)} MB
          </p>

          {getTags().length > 0 && (
            <div className="flex flex-wrap gap-1">
              {getTags().slice(0, 3).map((tag, index) => (
                <span key={index} className="px-2 py-1 bg-white/20 rounded-full text-xs">
                  {tag}
                </span>
              ))}
              {getTags().length > 3 && (
                <span className="px-2 py-1 bg-white/20 rounded-full text-xs">
                  +{getTags().length - 3}
                </span>
              )}
            </div>
          )}

          <div className="flex justify-between items-center mt-2 text-xs text-white/60">
            <span>👁 {item.reactions["view"] || 0}</span>
            <span>❤️ {item.reactions["like"] || 0}</span>
            <span>👎 {item.reactions["dislike"] || 0}</span>
          </div>
        </div>
      </div>

      {!item.imageUrl && (
        <>
          <div className="absolute -bottom-4 -right-4 w-16 h-16 bg-white/10 rounded-full"></div>
          <div className="absolute -top-2 -left-2 w-8 h-8 bg-white/20 rounded-full"></div>
        </>
      )}
    </div>
  );
};

// Masonry Feed Component
interface MasonryFeedProps {
  className?: string;
}

export const MasonryFeed: FC<MasonryFeedProps> = ({ className = '' }) => {
  const [windowSize, setWindowSize] = useState({
    width: globalThis.innerWidth,
    height: globalThis.innerHeight,
  });
  const feed = useFeed()
  const [columns, setColumns] = useState<MasonryItem[][]>([]);
  const [columnWidth, setColumnWidth] = useState<number>(0);
  const { items, loading, hasMore, error } = useFeed();
  const { lastElementRef } = useInfiniteScroll();

  useEffect(() => {
    const gap = 18
    const columnWidth = feed.getColumnWidth(windowSize.height, windowSize.width)
    const columnsCount = Math.floor((windowSize.width+gap*2) / (columnWidth+gap))

    const columns: MasonryItem[][] = Array.from({ length: columnsCount }, () => []);
    const columnHeights = new Array(columnsCount).fill(0);

    items.forEach((item) => {
      // Find the shortest column
      const shortestColumnIndex = columnHeights.indexOf(Math.min(...columnHeights));

      // Add item to shortest column
      columns[shortestColumnIndex].push(item);
      columnHeights[shortestColumnIndex] += item.height + gap;
    });
    setColumns(columns)
    setColumnWidth(columnWidth)
  }, [windowSize, items]);

  useEffect(() => {
    const handleResize = () => {
      setWindowSize({
        width: globalThis.innerWidth,
        height: globalThis.innerHeight,
      });
    };

    globalThis.addEventListener('resize', handleResize);
    return () => {
      globalThis.removeEventListener('resize', handleResize);
    };
  }, []);

  return (
    <div className={`px-4 pb-20 ${className}`}>
      {/* Error display */}
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          Error: {error}
        </div>
      )}

      {/* Masonry layout */}
      <div className="flex gap-4 justify-center">
        {columns.map((column, columnIndex) => (
          <div
            key={columnIndex}
            className="flex flex-col gap-4"
            style={{ width: `${columnWidth}px` }}
          >
            {column.map((item, itemIndex) => {
              const isLastInColumn = itemIndex === column.length - 1;
              const isLastOverall = columnIndex === columns.length - 1 && isLastInColumn;

              return (
                <MasonryFeedItem
                  key={item.id}
                  item={item}
                  isLast={isLastOverall}
                  lastElementRef={isLastOverall ? lastElementRef : undefined}
                />
              );
            })}
          </div>
        ))}
      </div>

      {/* Loading indicator */}
      {loading && (
        <div className="text-center py-8">
          <div className="inline-flex items-center gap-2 text-purple-500">
            <div className="w-4 h-4 border-2 border-purple-500 border-t-transparent rounded-full animate-spin"></div>
            Loading...
          </div>
        </div>
      )}

      {/* End of results */}
      {!hasMore && items.length > 0 && (
        <div className="text-center py-8 text-gray-500">
          No more items.
        </div>
      )}

      {/* Empty state */}
      {!loading && items.length === 0 && !error && (
        <div className="text-center py-12 text-gray-500">
          No items found.
        </div>
      )}
    </div>
  );
};

// Main Feed Page Component
export const FeedPage: FC = () => {
  const navigate = useNavigate();
  const { search } = useFeed();

  const [searchParams] = useSearchParams();
  const [searchInput, setSearchInput] = useState<string>("");


  // Initialize search from URL params
  useEffect(() => {
    const searchQuery = searchParams.get('search');
    if (searchQuery) {
      const tags = searchQuery.split('+').map(tag => decodeURIComponent(tag));
      setSearchInput(tags.join(' '))
      search({ tags });
    } else {
      search({});
    }
  }, [searchParams]); // Removed search from deps to prevent infinite loop

  const handleSearchChange = (newValue: string) => {
    setSearchInput(newValue)
  };

  const handleSearchSubmit = () => {
    // TODO:
    const tags = searchInput.split(' ')
    const searchQuery = tags.join('+');
    if (searchQuery.length > 0) {
      navigate("/?search="+searchQuery)
    } else {
      navigate('/')
    }
  };

  const handleAccountClick = () => {
    navigate("/me");
  };

  const handleUploadClick = () => {
    navigate("/upload");
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        searchValue={searchInput}
        onSearch={handleSearchSubmit}
        onSearchChange={handleSearchChange}
        onAccountClick={handleAccountClick}
      />

      <main className="mt-5">
        <MasonryFeed />
      </main>

      {/* Upload FAB */}
      <button
        type="button"
        className="fixed bottom-8 left-1/2 -translate-x-1/2 w-14 h-14 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full shadow-lg hover:scale-110 transition-all duration-200 flex items-center justify-center z-50 group"
        onClick={handleUploadClick}
      >
        <Plus className="w-6 h-6 text-white group-hover:rotate-90 transition-transform duration-300" />
      </button>
    </div>
  );
};