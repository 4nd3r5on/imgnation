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
