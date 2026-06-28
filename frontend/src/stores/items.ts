import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { Item } from "@/types";
import * as api from "@/api/client";
import { useFeedsStore } from "./feeds";

export const useItemsStore = defineStore("items", () => {
  const items = ref<Item[]>([]);
  const total = ref(0);
  const page = ref(1);
  const perPage = ref(20);
  const loading = ref(false);
  const filterFeedId = ref<number | undefined>();
  const filterRead = ref<boolean | undefined>();
  const filterStarred = ref<boolean | undefined>();

  const hasMore = computed(() => items.value.length < total.value);

  async function loadItems() {
    loading.value = true;
    try {
      const res = await api.fetchItems({
        page: page.value,
        per_page: perPage.value,
        feed_id: filterFeedId.value,
        read: filterRead.value,
        starred: filterStarred.value,
      });
      items.value = res.items;
      total.value = res.total;
    } finally {
      loading.value = false;
    }
  }

  async function loadMore() {
    if (!hasMore.value || loading.value) return;
    page.value++;
    const res = await api.fetchItems({
      page: page.value,
      per_page: perPage.value,
      feed_id: filterFeedId.value,
      read: filterRead.value,
      starred: filterStarred.value,
    });
    items.value.push(...res.items);
  }

  async function toggleRead(item: Item) {
    const wasUnread = !item.read;
    const updated = await api.updateItem(item.id, { read: !item.read });
    const idx = items.value.findIndex((i) => i.id === item.id);
    if (idx !== -1) items.value[idx] = updated;
    const feedsStore = useFeedsStore();
    const feed = feedsStore.feeds.find((f) => f.id === item.feed_id);
    if (feed && feed.unread_count !== undefined) {
      feed.unread_count += wasUnread ? -1 : 1;
    }
  }

  async function toggleStarred(item: Item) {
    const updated = await api.updateItem(item.id, { starred: !item.starred });
    const idx = items.value.findIndex((i) => i.id === item.id);
    if (idx !== -1) items.value[idx] = updated;
  }

  function resetFilters() {
    filterFeedId.value = undefined;
    filterRead.value = undefined;
    filterStarred.value = undefined;
  }

  function setFilterFeedId(id: number | undefined) {
    resetFilters();
    filterFeedId.value = id;
    page.value = 1;
    loadItems();
  }

  function setFilterRead(read: boolean | undefined) {
    resetFilters();
    filterRead.value = read;
    page.value = 1;
    loadItems();
  }

  function setFilterStarred(starred: boolean | undefined) {
    resetFilters();
    filterStarred.value = starred;
    page.value = 1;
    loadItems();
  }

  return {
    items,
    total,
    page,
    loading,
    hasMore,
    filterFeedId,
    filterRead,
    filterStarred,
    loadItems,
    loadMore,
    toggleRead,
    toggleStarred,
    setFilterFeedId,
    setFilterRead,
    setFilterStarred,
  };
});
