import axios from 'axios'
import type { Feed, Item, ItemsResponse } from '@/types'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
})

export async function fetchFeeds(): Promise<Feed[]> {
  const res = await api.get('/api/feeds')
  return res.data
}

export async function addFeed(url: string): Promise<Feed> {
  const res = await api.post('/api/feeds', { url })
  return res.data
}

export async function deleteFeed(id: number): Promise<void> {
  await api.delete(`/api/feeds/${id}`)
}

export async function refreshFeed(id: number): Promise<void> {
  await api.post(`/api/feeds/${id}/refresh`)
}

export async function markFeedRead(id: number): Promise<void> {
  await api.post(`/api/feeds/${id}/read-all`)
}

export async function fetchItems(params?: {
  page?: number
  per_page?: number
  feed_id?: number
  read?: boolean
  starred?: boolean
}): Promise<ItemsResponse> {
  const res = await api.get('/api/items', { params })
  return res.data
}

export async function markAllItemsRead(): Promise<void> {
  await api.post('/api/items/read-all')
}

export async function updateItem(id: number, data: { read?: boolean; starred?: boolean }): Promise<Item> {
  const res = await api.patch(`/api/items/${id}`, data)
  return res.data
}

export async function fetchSettings(): Promise<Record<string, string>> {
  const res = await api.get('/api/settings')
  return res.data
}

export async function updateSettings(data: Record<string, string>): Promise<Record<string, string>> {
  const res = await api.patch('/api/settings', data)
  return res.data
}
