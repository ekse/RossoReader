<script setup lang="ts">
import { onMounted } from 'vue'
import { useItemsStore } from '@/stores/items'
import { useFeedsStore } from '@/stores/feeds'
import ItemList from '@/components/ItemList.vue'

const itemsStore = useItemsStore()
const feedsStore = useFeedsStore()

onMounted(() => {
  itemsStore.setFilterStarred(true)
})
</script>

<template>
  <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex items-center justify-between">
    <h2 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Starred</h2>
  </div>
  <ItemList :items="itemsStore.items" :loading="itemsStore.loading" :feed-names="feedsStore.feedNames"
    @toggle-read="itemsStore.toggleRead" @toggle-starred="itemsStore.toggleStarred" @load-more="itemsStore.loadMore" />
</template>
