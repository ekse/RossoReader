<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useItemsStore } from '@/stores/items'
import { useFeedsStore } from '@/stores/feeds'
import ItemList from '@/components/ItemList.vue'
import AddFeedDialog from '@/components/AddFeedDialog.vue'
import { ref } from 'vue'
import * as api from '@/api/client'

const route = useRoute()
const itemsStore = useItemsStore()
const feedsStore = useFeedsStore()
const showAddFeed = ref(false)

const feedId = ref<number | null>(null)

onMounted(() => {
  loadFeed()
})

watch(() => route.params.id, () => {
  loadFeed()
})

async function markAllRead() {
  if (!feedId.value) return
  await api.markFeedRead(feedId.value)
  itemsStore.loadItems()
  const feed = feedsStore.feeds.find(f => f.id === feedId.value)
  if (feed) feed.unread_count = 0
}

function loadFeed() {
  const id = Number(route.params.id)
  if (id) {
    feedId.value = id
    itemsStore.setFilterFeedId(id)
  }
}
</script>

<template>
  <div>
    <div class="px-6 py-4 border-b border-gray-200 bg-white flex items-center justify-between">
      <button
        @click="showAddFeed = true"
        class="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
      >
        + Add Feed
      </button>
      <button
        @click="markAllRead"
        class="px-3 py-1.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
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

    <AddFeedDialog v-if="showAddFeed" @close="showAddFeed = false" />
  </div>
</template>
