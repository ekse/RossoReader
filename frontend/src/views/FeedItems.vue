<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useItemsStore } from "@/stores/items";
import { useFeedsStore } from "@/stores/feeds";
import ItemList from "@/components/ItemList.vue";
import TopBar from "@/components/TopBar.vue";
import LabelPicker from "@/components/LabelPicker.vue";
import * as api from "@/api/client";

const route = useRoute();
const itemsStore = useItemsStore();
const feedsStore = useFeedsStore();
const feedId = ref<number | null>(null);
const showLabelPicker = ref(false);

onMounted(() => {
  loadFeed();
});

watch(
  () => route.params.id,
  () => {
    loadFeed();
  },
);

async function markAllRead() {
  if (!feedId.value) return;
  await api.markFeedRead(feedId.value);
  itemsStore.loadItems();
  const feed = feedsStore.feeds.find((f) => f.id === feedId.value);
  if (feed) feed.unread_count = 0;
}

function loadFeed() {
  const id = Number(route.params.id);
  if (id) {
    feedId.value = id;
    itemsStore.setFilterFeedId(id);
    showLabelPicker.value = false;
  }
}
</script>

<template>
  <div>
    <TopBar show-mark-all-read @mark-all-read="markAllRead">
      <template #left-actions>
        <div v-if="feedId" class="relative">
          <button
            @click="showLabelPicker = !showLabelPicker"
            class="px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            <span class="hidden md:inline">Edit Labels</span>
            <span class="inline md:hidden">Labels</span>
          </button>
          <LabelPicker v-if="showLabelPicker" :feed-id="feedId" @close="showLabelPicker = false" />
        </div>
      </template>
    </TopBar>

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
