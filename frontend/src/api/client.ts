import axios from 'axios'
import type { Feed, Item, ItemsResponse, DiscoveredFeed, User, Passkey } from '@/types'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  withCredentials: true,
})

// On 401, redirect to the login page (unless already there).
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response && err.response.status === 401) {
      const path = window.location.pathname
      if (path !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(err)
  },
)

// Auth

export async function login(username: string, password: string): Promise<User> {
  const res = await api.post('/api/auth/login', { username, password })
  return res.data
}

export async function logout(): Promise<void> {
  await api.post('/api/auth/logout')
}

export async function getMe(): Promise<User> {
  const res = await api.get('/api/auth/me')
  return res.data
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  await api.put('/api/auth/password', {
    current_password: currentPassword,
    new_password: newPassword,
  })
}

export async function listUsers(): Promise<User[]> {
  const res = await api.get('/api/users')
  return res.data
}

export async function createUser(username: string, password: string, isAdmin: boolean): Promise<User> {
  const res = await api.post('/api/users', { username, password, is_admin: isAdmin })
  return res.data
}

export async function deleteUser(id: number): Promise<void> {
  await api.delete(`/api/users/${id}`)
}

// Feeds

export async function fetchFeeds(): Promise<Feed[]> {
  const res = await api.get('/api/feeds')
  return res.data
}

export async function addFeed(url: string): Promise<Feed> {
  const res = await api.post('/api/feeds', { url })
  return res.data
}

export async function discoverFeeds(url: string): Promise<DiscoveredFeed[]> {
  const res = await api.post('/api/feeds/discover', { url })
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

// Items

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

// Settings

export async function fetchSettings(): Promise<Record<string, string>> {
  const res = await api.get('/api/settings')
  return res.data
}

export async function updateSettings(data: Record<string, string>): Promise<Record<string, string>> {
  const res = await api.patch('/api/settings', data)
  return res.data
}

// Passkeys

export async function passkeyRegisterBegin(): Promise<{ state_id: string; options: any }> {
  const res = await api.post('/api/auth/passkey/register/begin')
  return res.data
}

export async function passkeyRegisterFinish(stateId: string, name: string, credential: any): Promise<Passkey> {
  const res = await api.post('/api/auth/passkey/register/finish', { state_id: stateId, name, credential })
  return res.data
}

export async function passkeyLoginBegin(): Promise<{ state_id: string; options: any }> {
  const res = await api.post('/api/auth/passkey/login/begin')
  return res.data
}

export async function passkeyLoginFinish(stateId: string, credential: any): Promise<User> {
  const res = await api.post('/api/auth/passkey/login/finish', { state_id: stateId, credential })
  return res.data
}

export async function listPasskeys(): Promise<Passkey[]> {
  const res = await api.get('/api/auth/passkeys')
  return res.data
}

export async function deletePasskey(id: number): Promise<void> {
  await api.delete(`/api/auth/passkeys/${id}`)
}