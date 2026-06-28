<script setup lang="ts">
import { onMounted } from "vue";
import { useItemsStore } from "@/stores/items";
import { useFeedsStore } from "@/stores/feeds";
import ItemList from "@/components/ItemList.vue";
import TopBar from "@/components/TopBar.vue";
import * as api from "@/api/client";

const itemsStore = useItemsStore();
const feedsStore = useFeedsStore();
onMounted(() => {
  itemsStore.setFilterRead(false);
});

async function markAllRead() {
  await api.markAllItemsRead();
  itemsStore.loadItems();
  for (const feed of feedsStore.feeds) {
    feed.unread_count = 0;
  }
}
</script>

<template>
  <div>
    <TopBar title="New" show-mark-all-read @mark-all-read="markAllRead" />
    <ItemList
      :items="itemsStore.items"
      :loading="itemsStore.loading"
      :has-more="itemsStore.hasMore"
      :feed-names="feedsStore.feedNames"
      @toggle-read="itemsStore.toggleRead"
      @toggle-starred="itemsStore.toggleStarred"
      @load-more="itemsStore.loadMore"
    />
  </div>
</template>
