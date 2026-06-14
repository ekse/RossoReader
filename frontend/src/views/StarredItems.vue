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
  <ItemList
    :items="itemsStore.items"
    :loading="itemsStore.loading"
    @toggle-read="itemsStore.toggleRead"
    @toggle-starred="itemsStore.toggleStarred"
    @load-more="itemsStore.loadMore"
  />
</template>
