import { defineStore } from "pinia";
import { computed, ref } from "vue";
import type { Feed } from "@/types";
import * as api from "@/api/client";

export const useFeedsStore = defineStore("feeds", () => {
  const feeds = ref<Feed[]>([]);
  const loading = ref(false);

  const totalUnread = computed(() =>
    feeds.value.reduce((sum, f) => sum + (f.unread_count || 0), 0),
  );

  const feedNames = computed(() => {
    const map: Record<number, string> = {};
    for (const f of feeds.value) {
      map[f.id] = f.title || f.url;
    }
    return map;
  });

  async function loadFeeds() {
    loading.value = true;
    try {
      feeds.value = await api.fetchFeeds();
    } finally {
      loading.value = false;
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
          const axiosErr = err as { response?: { status?: number } };
          if (axiosErr.response?.status === 409) {
            skipped++;
            continue;
          }
        }
      }
    }
    await loadFeeds();
    return { imported, skipped };
  }

  return { feeds, loading, totalUnread, feedNames, loadFeeds, addFeed, removeFeed, refreshFeed, importFeeds };
});
