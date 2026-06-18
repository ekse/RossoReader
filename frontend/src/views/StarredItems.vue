<script setup lang="ts">
import { onMounted } from 'vue'
import { useItemsStore } from '@/stores/items'
import ItemList from '@/components/ItemList.vue'

const itemsStore = useItemsStore()

onMounted(() => {
  itemsStore.setFilterStarred(true)
})
</script>

<template>
  <div class="px-6 py-4 border-b border-gray-200 bg-white flex items-center justify-between">
    <h2 class="text-sm font-semibold text-gray-900">Starred</h2>
  </div>
  <ItemList :items="itemsStore.items" :loading="itemsStore.loading" @toggle-read="itemsStore.toggleRead"
    @toggle-starred="itemsStore.toggleStarred" @load-more="itemsStore.loadMore" />
</template>
