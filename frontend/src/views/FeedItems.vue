<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
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

const feedName = computed(() => (feedId.value ? feedsStore.feedNames[feedId.value] : undefined));
const feedIcon = computed(() => (feedId.value ? feedsStore.feedIcons[feedId.value] : undefined));

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
        <div v-if="feedId" class="flex items-center gap-2">
          <img v-if="feedIcon" :src="feedIcon" class="w-4 h-4 rounded shrink-0" alt="" />
          <span v-if="feedName" class="text-sm font-bold text-gray-700 dark:text-gray-100">{{
            feedName
          }}</span>
        </div>
        <div v-if="feedId" class="relative">
          <button
            @click="showLabelPicker = !showLabelPicker"
            class="p-1.5 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600"
            title="Edit labels"
            aria-label="Edit labels"
          >
            <svg class="w-5 h-5"><use href="#icon-tag" /></svg>
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
      :feed-icons="feedsStore.feedIcons"
      :show-feed="false"
      @toggle-read="itemsStore.toggleRead"
      @toggle-starred="itemsStore.toggleStarred"
      @load-more="itemsStore.loadMore"
    />
  </div>
</template>
