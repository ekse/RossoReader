<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useItemsStore } from '@/stores/items'
import { useFeedsStore } from '@/stores/feeds'
import ItemList from '@/components/ItemList.vue'
import TopBar from '@/components/TopBar.vue'
import * as api from '@/api/client'

const route = useRoute()
const itemsStore = useItemsStore()
const feedsStore = useFeedsStore()
const feedId = ref<number | null>(null)
const feedName = computed(() => feedId.value ? feedsStore.feedNames[feedId.value] : undefined)

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
    <TopBar show-mark-all-read @mark-all-read="markAllRead" />

    <ItemList
      :items="itemsStore.items"
      :loading="itemsStore.loading"
      :feed-names="feedsStore.feedNames"
      @toggle-read="itemsStore.toggleRead"
      @toggle-starred="itemsStore.toggleStarred"
      @load-more="itemsStore.loadMore"
    />
  </div>
</template>

