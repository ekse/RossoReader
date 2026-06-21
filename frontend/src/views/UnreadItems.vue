<script setup lang="ts">
import { onMounted } from 'vue'
import { useItemsStore } from '@/stores/items'
import { useFeedsStore } from '@/stores/feeds'
import ItemList from '@/components/ItemList.vue'
import * as api from '@/api/client'

const itemsStore = useItemsStore()
const feedsStore = useFeedsStore()

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
    <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex items-center justify-between">
      <h2 class="text-sm font-semibold text-gray-900 dark:text-gray-100">New content</h2>
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
