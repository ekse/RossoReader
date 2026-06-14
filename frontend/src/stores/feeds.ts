import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Feed } from '@/types'
import * as api from '@/api/client'

export const useFeedsStore = defineStore('feeds', () => {
  const feeds = ref<Feed[]>([])
  const loading = ref(false)

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

  return { feeds, loading, loadFeeds, addFeed, removeFeed, refreshFeed }
})
