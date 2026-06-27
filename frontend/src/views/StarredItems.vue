<script setup lang="ts">
import { onMounted } from 'vue'
import { useItemsStore } from '@/stores/items'
import { useFeedsStore } from '@/stores/feeds'
import ItemList from '@/components/ItemList.vue'
import TopBar from '@/components/TopBar.vue'

const itemsStore = useItemsStore()
const feedsStore = useFeedsStore()
onMounted(() => {
  itemsStore.setFilterStarred(true)
})
</script>

<template>
  <TopBar title="Starred" />
  <ItemList :items="itemsStore.items" :loading="itemsStore.loading" :has-more="itemsStore.hasMore" :feed-names="feedsStore.feedNames"
    @toggle-read="itemsStore.toggleRead" @toggle-starred="itemsStore.toggleStarred" @load-more="itemsStore.loadMore" />
</template>

