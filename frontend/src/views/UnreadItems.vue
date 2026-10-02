<script setup lang="ts">
import { onMounted } from "vue";
import { useItemsStore } from "@/stores/items";
import { useFeedsStore } from "@/stores/feeds";
import { useItemGrouping } from "@/composables/useItemGrouping";
import { ITEM_GROUPING_FEED, ITEM_GROUPING_NONE } from "@/types";
import ItemList from "@/components/ItemList.vue";
import TopBar from "@/components/TopBar.vue";
import * as api from "@/api/client";

const itemsStore = useItemsStore();
const feedsStore = useFeedsStore();
const { grouping } = useItemGrouping();
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
    <TopBar title="New" show-mark-all-read @mark-all-read="markAllRead">
      <template #actions>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <span class="hidden md:inline">Group by</span>
          <select
            v-model="grouping"
            class="px-2 py-1.5 text-sm rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300"
          >
            <option :value="ITEM_GROUPING_NONE">None</option>
            <option :value="ITEM_GROUPING_FEED">By Feed</option>
          </select>
        </label>
      </template>
    </TopBar>
    <ItemList
      :items="itemsStore.items"
      :loading="itemsStore.loading"
      :has-more="itemsStore.hasMore"
      :feed-names="feedsStore.feedNames"
      :group-by="grouping"
      @toggle-read="itemsStore.toggleRead"
      @toggle-starred="itemsStore.toggleStarred"
      @load-more="itemsStore.loadMore"
    />
  </div>
</template>
