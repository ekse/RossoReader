export const DEFAULT_ITEMS_LIMIT = 150;
export const DEFAULT_FEEDS_LIMIT = 200;

export interface Feed {
  id: number;
  url: string;
  title: string;
  description?: string;
  site_link?: string;
  icon_url?: string;
  last_fetched_at?: string;
  last_fetch_error?: string;
  created_at: string;
  updated_at: string;
  unread_count?: number;
}

export interface Label {
  id: number;
  user_id: number;
  name: string;
  created_at: string;
}

export interface LabelGroup {
  label: Label;
  feeds: Feed[];
}

export interface GroupedFeedsResponse {
  label_groups: LabelGroup[];
  unlabeled_feeds: Feed[];
}

export interface Item {
  id: number;
  feed_id: number;
  guid: string;
  title: string;
  url: string;
  content?: string;
  description?: string;
  author?: string;
  published_at?: string;
  fetched_at: string;
  read: boolean;
  starred: boolean;
}

export interface ItemsResponse {
  items: Item[];
  total: number;
  page: number;
}

export interface DiscoveredFeed {
  url: string;
  title: string;
}

export interface User {
  id: number;
  username: string;
  is_admin: boolean;
}

export interface Passkey {
  id: number;
  name: string;
  transports: string[];
  backup_eligible: boolean;
  created_at: string;
}

export interface AdminSettings {
  items_limit: number;
  feeds_limit: number;
}
