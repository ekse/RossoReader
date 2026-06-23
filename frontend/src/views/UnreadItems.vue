<script setup lang="ts">
import { onMounted } from 'vue'
import { useItemsStore } from '@/stores/items'
import { useFeedsStore } from '@/stores/feeds'
import ItemList from '@/components/ItemList.vue'
import { useSidebar } from '@/composables/useSidebar'
import * as api from '@/api/client'

const itemsStore = useItemsStore()
const feedsStore = useFeedsStore()
const { toggle } = useSidebar()

onMounted(() => {
  itemsStore.setFilterRead(false)
})

async function markAllRead() {
  await api.markAllItemsRead()
  itemsStore.loadItems()
  for (const feed of feedsStore.feeds) {
    feed.unread_count = 0
  }
}
</script>

<template>
  <div>
    <div class="px-4 py-3 md:px-6 md:py-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <button
          @click="toggle"
          class="p-1.5 -ml-1 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-700 md:hidden transition-colors"
          title="Toggle Sidebar"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <h2 class="text-sm font-semibold text-gray-900 dark:text-gray-100">New content</h2>
      </div>
      <button
        @click="markAllRead"
        class="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
      >
        Mark all as read
      </button>
    </div>
    <ItemList
      :items="itemsStore.items"
      :loading="itemsStore.loading"
      @toggle-read="itemsStore.toggleRead"
      @toggle-starred="itemsStore.toggleStarred"
      @load-more="itemsStore.loadMore"
    />
  </div>
</template>

