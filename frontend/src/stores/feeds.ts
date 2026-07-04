import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { DEFAULT_FEEDS_LIMIT, type Feed, type Label, type LabelGroup } from "@/types";
import * as api from "@/api/client";

export const useFeedsStore = defineStore("feeds", () => {
  const feeds = ref<Feed[]>([]);
  const loading = ref(false);
  const filterUnreadOnly = ref(false);
  const feedsLimit = ref(DEFAULT_FEEDS_LIMIT);
  const labels = ref<Label[]>([]);
  const labelGroups = ref<LabelGroup[]>([]);
  const unlabeledFeeds = ref<Feed[]>([]);
  const collapsedLabelIds = ref(new Set<number>());

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

  const visibleLabelGroups = computed(() => {
    if (!filterUnreadOnly.value) return labelGroups.value;
    return labelGroups.value.filter((g) => g.feeds.some((f) => (f.unread_count || 0) > 0));
  });

  const visibleUnlabeledFeeds = computed(() => {
    if (!filterUnreadOnly.value) return unlabeledFeeds.value;
    return unlabeledFeeds.value.filter((f) => (f.unread_count || 0) > 0);
  });

  const orderedVisibleFeeds = computed(() => {
    const result: Feed[] = [];
    for (const g of visibleLabelGroups.value) {
      if (!collapsedLabelIds.value.has(g.label.id)) {
        result.push(...g.feeds);
      }
    }
    result.push(...visibleUnlabeledFeeds.value);
    return result;
  });

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

  async function loadGroupedFeeds() {
    const data = await api.fetchGroupedFeeds();
    labelGroups.value = data.label_groups;
    unlabeledFeeds.value = data.unlabeled_feeds;
    feeds.value = [...data.unlabeled_feeds, ...data.label_groups.flatMap((g) => g.feeds)];
    return data;
  }

  async function loadLabels() {
    labels.value = await api.fetchLabels();
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

  function toggleCollapseLabel(id: number) {
    const s = new Set(collapsedLabelIds.value);
    if (s.has(id)) {
      s.delete(id);
    } else {
      s.add(id);
    }
    collapsedLabelIds.value = s;
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
    labels,
    labelGroups,
    unlabeledFeeds,
    collapsedLabelIds,
    visibleLabelGroups,
    visibleUnlabeledFeeds,
    orderedVisibleFeeds,
    loadFeeds,
    loadFeedsLimit,
    loadGroupedFeeds,
    loadLabels,
    addFeed,
    removeFeed,
    refreshFeed,
    importFeeds,
    toggleCollapseLabel,
  };
});
