import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { Feed } from '@/types'
import * as api from '@/api/client'

export const useFeedsStore = defineStore('feeds', () => {
  const feeds = ref<Feed[]>([])
  const loading = ref(false)

  const feedNames = computed(() => {
    const map: Record<number, string> = {}
    for (const f of feeds.value) {
      map[f.id] = f.title || f.url
    }
    return map
  })

  async function loadFeeds() {
    loading.value = true
    try {
      feeds.value = await api.fetchFeeds()
    } finally {
      loading.value = false
    }
  }

  async function addFeed(url: string) {
    const feed = await api.addFeed(url)
    feeds.value.push(feed)
    return feed
  }

  async function removeFeed(id: number) {
    await api.deleteFeed(id)
    feeds.value = feeds.value.filter(f => f.id !== id)
  }

  async function refreshFeed(id: number) {
    await api.refreshFeed(id)
    await loadFeeds()
  }

  return { feeds, loading, feedNames, loadFeeds, addFeed, removeFeed, refreshFeed }
})
