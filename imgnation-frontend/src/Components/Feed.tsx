import {
  FC, useCallback, useEffect, useRef, useState, ReactNode
} from "react";
import { useNavigate } from "react-router";
import {MasonryItem} from "../hooks/useFeed.ts";

interface MasonryFeedItemProps {
  item: MasonryItem;
}

interface MasonryFeedProps {
  items: MasonryItem[];
  hasMore: boolean;
  loading: boolean;
  onLoadMoreRequest?: () => void;
  onReloadRequest?: () => void;
  className?: string;
  children: (item: MasonryItem) => ReactNode;
}

export const MasonryFeedItem: FC<MasonryFeedItemProps> = ({ item }) => {
  const navigate = useNavigate();

  const getDisplayName = () =>
    item.uploads.length > 0 ? item.uploads[0].filename : `Item ${item.id.slice(0, 8)}`;

  const formatDate = (dateString: string) =>
    new Date(dateString).toLocaleDateString();

  const getTags = () =>
    item.uploads.length > 0 ? item.uploads[0].tags : [];

  return (
    <div
      onClick={() => navigate('/uploads?id=' + item.id)}
      className={`${item.imageUrl ? 'bg-gray-200' : item.color} rounded-xl shadow-lg hover:scale-[1.02] transition-transform cursor-pointer relative overflow-hidden`}
      style={{ minHeight: item.height }}
    >
      {item.imageUrl ? (
        <img src={item.imageUrl} alt={getDisplayName()} className="w-full h-full object-cover" />
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
                <span key={index} className="px-2 py-1 bg-white/20 rounded-full text-xs">{tag}</span>
              ))}
              {getTags().length > 3 && (
                <span className="px-2 py-1 bg-white/20 rounded-full text-xs">+{getTags().length - 3}</span>
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

export const MasonryFeed: FC<MasonryFeedProps> = ({
  items,
  hasMore,
  loading,
  onLoadMoreRequest,
  onReloadRequest,
  children,
  className = "",
}) => {
  const [masonryItems, setMasonryItems] = useState<MasonryItem[]>(items);
  const [columns, setColumns] = useState<MasonryItem[][]>([]);
  const [windowWidth, setWindowWidth] = useState(globalThis.innerWidth);
  const [windowHeight, setWindowHeight] = useState(globalThis.innerHeight);

  const feedRef = useRef<HTMLDivElement>(null);
  const lastScrollY = useRef<number>(0);
  const scrollTimeoutRef = useRef<number | null>(null);

  const getColumnWidth = useCallback(() =>
    windowWidth > windowHeight ? 300 : 200, [windowWidth, windowHeight]);

  useEffect(() => {
    const resize = () => {
      setWindowWidth(globalThis.innerWidth);
      setWindowHeight(globalThis.innerHeight);
    };
    globalThis.addEventListener("resize", resize);
    return () => globalThis.removeEventListener("resize", resize);
  }, []);

  useEffect(() => {
    const feedWidth = feedRef.current?.offsetWidth || 0;
    const containerWidth = feedWidth - 32;
    const colWidth = getColumnWidth();
    const cols = Math.max(1, Math.floor((containerWidth + 16) / (colWidth + 16)));

    const newCols: MasonryItem[][] = Array.from({ length: cols }, () => []);
    masonryItems.forEach((item, i) => {
      newCols[i % cols].push(item);
    });

    setColumns(newCols);
  }, [masonryItems, windowWidth, windowHeight, getColumnWidth]);

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;
    const y = scrollTop;

    if (scrollTimeoutRef.current) clearTimeout(scrollTimeoutRef.current);

    if (scrollTop <= 0 && lastScrollY.current > scrollTop && onReloadRequest) {
      scrollTimeoutRef.current = setTimeout(() => {
        if (scrollTop <= -50) onReloadRequest();
      }, 100);
    }

    if (scrollHeight - scrollTop - clientHeight < 200 && hasMore && !loading && onLoadMoreRequest) {
      onLoadMoreRequest();
    }

    lastScrollY.current = y;
  };

  return (
    <div
      ref={feedRef}
      className={`h-screen overflow-auto px-4 pb-20 ${className}`}
      onScroll={handleScroll}
      style={{ maxHeight: "calc(100vh - 80px)" }}
    >
      <div className="flex gap-4 justify-center">
        {columns.map((col, i) => (
          <div
            key={i}
            className="flex flex-col gap-4"
            style={{ width: `${getColumnWidth()}px` }}
          >
            {col.map(item => children(item))}
          </div>
        ))}
      </div>

      {loading && (
        <div className="text-center py-4 text-purple-500">Loading...</div>
      )}
      {!hasMore && items.length > 0 && (
        <div className="text-center py-4 text-gray-500">No more items.</div>
      )}
    </div>
  );
};
