<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useFeedsStore } from '@/stores/feeds'

const feedsStore = useFeedsStore()
const router = useRouter()
const route = useRoute()

onMounted(() => {
  feedsStore.loadFeeds()
})

function selectFeed(id: number) {
  router.push(`/feed/${id}`)
}
</script>

<template>
  <div class="p-4">
    <router-link
      to="/starred"
      class="block px-3 py-2 rounded-md text-sm font-medium mt-1"
      :class="route.path === '/starred' ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-100'"
    >
      Starred
    </router-link>

    <div class="mt-6">
      <h3 class="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">Feeds</h3>
      <div class="mt-2 space-y-1">
        <button
          v-for="feed in feedsStore.feeds"
          :key="feed.id"
          @click="selectFeed(feed.id)"
          class="w-full text-left px-3 py-2 rounded-md text-sm"
          :class="route.params.id === String(feed.id) ? 'bg-blue-50 text-blue-700' : 'text-gray-700 hover:bg-gray-100'"
        >
          <div class="flex items-center justify-between">
            <span class="truncate">{{ feed.title || feed.url }}</span>
            <span
              v-if="feed.unread_count && feed.unread_count > 0"
              class="ml-2 inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800"
            >
              {{ feed.unread_count }}
            </span>
          </div>
        </button>
      </div>
    </div>

    <div class="mt-6 px-3">
      <router-link
        to="/settings"
        class="block text-sm text-gray-500 hover:text-gray-700"
        :class="{ 'text-blue-700': route.path === '/settings' }"
      >
        Settings
      </router-link>
    </div>
  </div>
</template>
