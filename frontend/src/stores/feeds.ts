import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { DEFAULT_FEEDS_LIMIT, type Feed } from "@/types";
import * as api from "@/api/client";

export const useFeedsStore = defineStore("feeds", () => {
  const feeds = ref<Feed[]>([]);
  const loading = ref(false);
  const filterUnreadOnly = ref(false);
  const feedsLimit = ref(DEFAULT_FEEDS_LIMIT);

  const totalUnread = computed(() =>
    feeds.value.reduce((sum, f) => sum + (f.unread_count || 0), 0),
  );

  const visibleFeeds = computed(() => {
    if (!filterUnreadOnly.value) return feeds.value;
    return feeds.value.filter((f) => (f.unread_count || 0) > 0);
  });

  const feedNames = computed(() => {
    const map: Record<number, string> = {};
    for (const f of feeds.value) {
      map[f.id] = f.title || f.url;
    }
    return map;
  });

  const hasReachedLimit = computed(() => feeds.value.length >= feedsLimit.value);

  async function loadFeeds() {
    loading.value = true;
    try {
      feeds.value = await api.fetchFeeds();
    } finally {
      loading.value = false;
    }
  }

  async function loadFeedsLimit() {
    try {
      const settings = await api.fetchAdminSettings();
      feedsLimit.value = settings.feeds_limit;
    } catch {
      // defaults to DEFAULT_FEEDS_LIMIT
    }
  }

  async function addFeed(url: string) {
    const feed = await api.addFeed(url);
    feeds.value.push(feed);
    return feed;
  }

  async function removeFeed(id: number) {
    await api.deleteFeed(id);
    feeds.value = feeds.value.filter((f) => f.id !== id);
  }

  async function refreshFeed(id: number) {
    await api.refreshFeed(id);
    await loadFeeds();
  }

  async function importFeeds(urls: string[]): Promise<{ imported: number; skipped: number }> {
    let imported = 0;
    let skipped = 0;
    for (const url of urls) {
      try {
        await api.addFeed(url);
        imported++;
      } catch (err: unknown) {
        if (err && typeof err === "object" && "response" in err) {
          const apiErr = err as { response?: { status?: number } };
          if (apiErr.response?.status === 409) {
            skipped++;
            continue;
          }
        }
      }
    }
    await loadFeeds();
    return { imported, skipped };
  }

  return {
    feeds,
    loading,
    totalUnread,
    feedNames,
    visibleFeeds,
    filterUnreadOnly,
    feedsLimit,
    hasReachedLimit,
    loadFeeds,
    loadFeedsLimit,
    addFeed,
    removeFeed,
    refreshFeed,
    importFeeds,
  };
});
