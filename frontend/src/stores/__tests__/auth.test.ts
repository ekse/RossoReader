import { setActivePinia, createPinia } from 'pinia'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useAuthStore } from '@/stores/auth'

vi.mock('@/api/client', () => ({
  getMe: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
}))

import * as api from '@/api/client'

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts unauthenticated', () => {
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.isAdmin).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('login stores the user', async () => {
    const user = { id: 1, username: 'alice', is_admin: true }
    ;(api.login as any).mockResolvedValue(user)
    const auth = useAuthStore()
    await auth.login('alice', 'pw')
    expect(auth.user).toEqual(user)
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.isAdmin).toBe(true)
  })

  it('logout clears the user', async () => {
    ;(api.logout as any).mockResolvedValue(undefined)
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'alice', is_admin: true }
    await auth.logout()
    expect(auth.user).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
  })

  it('fetchMe sets user on success, clears on failure', async () => {
    ;(api.getMe as any).mockResolvedValue({ id: 1, username: 'alice', is_admin: false })
    const auth = useAuthStore()
    await auth.fetchMe()
    expect(auth.user?.username).toBe('alice')

    ;(api.getMe as any).mockRejectedValue(new Error('unauthorized'))
    await auth.fetchMe()
    expect(auth.user).toBeNull()
  })
})