export interface Feed {
  id: number
  url: string
  title: string
  description?: string
  site_link?: string
  last_fetched_at?: string
  created_at: string
  updated_at: string
  unread_count?: number
}

export interface Item {
  id: number
  feed_id: number
  guid: string
  title: string
  url: string
  content?: string
  description?: string
  author?: string
  published_at?: string
  fetched_at: string
  read: boolean
  starred: boolean
}

export interface ItemsResponse {
  items: Item[]
  total: number
  page: number
}
