<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useFeedsStore } from '@/stores/feeds'
import AddFeedDialog from './AddFeedDialog.vue'
import ThemeToggle from './ThemeToggle.vue'

const feedsStore = useFeedsStore()
const router = useRouter()
const route = useRoute()
const showAddFeed = ref(false)

onMounted(() => {
  feedsStore.loadFeeds()
})

function selectFeed(id: number) {
  router.push(`/feed/${id}`)
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString()
}

const lastUpdate = computed(() => {
  const dates = feedsStore.feeds
    .map(f => f.last_fetched_at)
    .filter(Boolean) as string[]
  if (dates.length === 0) return null
  dates.sort()
  return dates[dates.length - 1]
})

const totalUnread = computed(() =>
  feedsStore.feeds.reduce((sum, f) => sum + (f.unread_count || 0), 0)
)
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 p-4 overflow-y-auto">
      <div class="px-3 mb-3 flex items-center justify-between">
        <span class="flex items-center gap-2 font-bold tracking-[0.2em] text-red-600 select-none">
          🌹Rosso
        </span>
        <ThemeToggle />
      </div>
      <router-link to="/unread" class="block px-3 py-2 rounded-md text-sm font-medium"
        :class="route.path === '/unread' ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'">
        <span class="flex items-center justify-between">
          New
          <span v-if="totalUnread > 0"
            class="inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
            {{ totalUnread }}
          </span>
        </span>
      </router-link>
      <router-link to="/starred" class="block px-3 py-2 rounded-md text-sm font-medium mt-1"
        :class="route.path === '/starred' ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'">
        Starred
      </router-link>

      <router-link to="/settings"
        class="block px-3 py-2 rounded-md text-sm font-medium mt-1"
        :class="route.path === '/settings' ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'">
        Settings
      </router-link>


      <div class="mt-6">
        <div class="px-3 flex items-center">
          <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Feeds</h3>
          <button @click="showAddFeed = true"
            class="w-5 h-5 flex items-center font-bold justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-700 transition-colors">
            +
          </button>
        </div>
        <div class="mt-2 space-y-1">
          <button v-for="feed in feedsStore.feeds" :key="feed.id" @click="selectFeed(feed.id)"
            class="w-full text-left px-3 py-2 rounded-md text-sm"
            :class="route.params.id === String(feed.id) ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300' : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'">
            <div class="flex items-center gap-2">
              <img v-if="feed.icon_url" :src="feed.icon_url" class="w-4 h-4 shrink-0 rounded" alt="" loading="lazy" />
              <div class="flex items-center justify-between flex-1 min-w-0">
                <span class="truncate">{{ feed.title || feed.url }}</span>
                <span v-if="feed.unread_count && feed.unread_count > 0"
                  class="ml-2 shrink-0 inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                  {{ feed.unread_count }}
                </span>
              </div>
            </div>
          </button>
        </div>
      </div>
    </div>

    <div v-if="lastUpdate"
      class="shrink-0 px-4 py-3 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-400 dark:text-gray-500">
      Last updated: {{ formatDate(lastUpdate) }}
    </div>
    <AddFeedDialog v-if="showAddFeed" @close="showAddFeed = false" />
  </div>
</template>
