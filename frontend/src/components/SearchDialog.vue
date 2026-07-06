<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useFeedsStore } from "@/stores/feeds";
import { useSearch } from "@/composables/useSearch";
import { useCurrentItem } from "@/composables/useCurrentItem";
import * as api from "@/api/client";
import type { Item } from "@/types";

const emits = defineEmits<{ close: [] }>();

const feedsStore = useFeedsStore();
const { closeSearch } = useSearch();
const { pendingFocusItemId } = useCurrentItem();
const router = useRouter();

const searchInput = ref<HTMLInputElement | null>(null);
const query = ref("");

onMounted(() => {
  searchInput.value?.focus();
});
const results = ref<Item[]>([]);
const total = ref(0);
const page = ref(1);
const perPage = 20;
const loading = ref(false);
const searched = ref(false);
const error = ref("");

const selectedFeedIds = ref(new Set<number>());
const selectedLabelIds = ref(new Set<number>());

const showFilters = ref(false);

function toggleFeed(feedId: number) {
  const s = new Set(selectedFeedIds.value);
  if (s.has(feedId)) {
    s.delete(feedId);
  } else {
    s.add(feedId);
  }
  selectedFeedIds.value = s;
  search();
}

function toggleLabel(labelId: number, feedIds: number[]) {
  const ls = new Set(selectedLabelIds.value);
  const fs = new Set(selectedFeedIds.value);
  if (ls.has(labelId)) {
    ls.delete(labelId);
    for (const fid of feedIds) {
      fs.delete(fid);
    }
  } else {
    ls.add(labelId);
    for (const fid of feedIds) {
      fs.add(fid);
    }
  }
  selectedLabelIds.value = ls;
  selectedFeedIds.value = fs;
  search();
}

function clearFilters() {
  selectedFeedIds.value = new Set();
  selectedLabelIds.value = new Set();
  if (query.value.trim()) {
    search();
  }
}

const hasActiveFilters = computed(
  () => selectedFeedIds.value.size > 0 || selectedLabelIds.value.size > 0,
);

const hasMore = computed(() => results.value.length < total.value);

async function search() {
  const q = query.value.trim();
  if (!q) return;
  loading.value = true;
  error.value = "";
  searched.value = true;
  page.value = 1;
  try {
    const resp = await api.searchItems({
      q,
      page: 1,
      per_page: perPage,
      feed_ids: selectedFeedIds.value.size > 0 ? [...selectedFeedIds.value].join(",") : undefined,
      label_ids:
        selectedLabelIds.value.size > 0 ? [...selectedLabelIds.value].join(",") : undefined,
    });
    results.value = resp.items;
    total.value = resp.total;
  } catch (e: any) {
    error.value = e.response?.data || "Search failed";
    results.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

async function loadMore() {
  if (loading.value || !hasMore.value) return;
  loading.value = true;
  const nextPage = page.value + 1;
  try {
    const resp = await api.searchItems({
      q: query.value.trim(),
      page: nextPage,
      per_page: perPage,
      feed_ids: selectedFeedIds.value.size > 0 ? [...selectedFeedIds.value].join(",") : undefined,
      label_ids:
        selectedLabelIds.value.size > 0 ? [...selectedLabelIds.value].join(",") : undefined,
    });
    results.value.push(...resp.items);
    page.value = nextPage;
  } catch (e: any) {
    error.value = e.response?.data || "Search failed";
  } finally {
    loading.value = false;
  }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null;
function onQueryInput() {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    if (query.value.trim()) {
      search();
    } else {
      results.value = [];
      total.value = 0;
      searched.value = false;
    }
  }, 300);
}

function highlightText(text: string, q: string): string {
  if (!text || !q) return text || "";
  const div = document.createElement("div");
  div.innerHTML = text;
  const clean = div.textContent || div.innerText || "";
  if (!clean) return text;
  const escapedQ = q.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const regex = new RegExp(`(${escapedQ})`, "gi");
  return clean.replace(
    regex,
    '<mark class="bg-yellow-200 dark:bg-yellow-700/70 rounded px-0.5">$1</mark>',
  );
}

function navigateToItem(item: Item) {
  pendingFocusItemId.value = item.id;
  closeSearch();
  emits("close");
  const q = query.value.trim();
  const hash = q ? `#highlight=${encodeURIComponent(q)}` : "";
  router.push(`/feed/${item.feed_id}${hash}`);
}

const feedList = computed(() => {
  const list: { feedId: number; title: string; labelId?: number }[] = [];
  for (const group of feedsStore.labelGroups) {
    for (const feed of group.feeds) {
      list.push({ feedId: feed.id, title: feed.title || feed.url, labelId: group.label.id });
    }
  }
  for (const feed of feedsStore.unlabeledFeeds) {
    list.push({ feedId: feed.id, title: feed.title || feed.url });
  }
  return list;
});

const labelFeedMap = computed(() => {
  const map: Record<number, number[]> = {};
  for (const group of feedsStore.labelGroups) {
    map[group.label.id] = group.feeds.map((f) => f.id);
  }
  return map;
});
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-start justify-center pt-16 bg-black/50"
    @click.self="emits('close')"
    @keydown.escape.window="emits('close')"
  >
    <div
      class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl mx-4 flex flex-col max-h-[calc(100vh-8rem)]"
    >
      <div class="p-4 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center gap-3">
          <svg class="w-5 h-5 text-gray-400 shrink-0">
            <use href="#icon-search" />
          </svg>
          <input
            ref="searchInput"
            v-model="query"
            @input="onQueryInput"
            type="text"
            placeholder="Search title and content..."
            class="flex-1 bg-transparent border-none outline-none text-sm text-gray-900 dark:text-gray-100 placeholder-gray-400"
          />
          <button
            @click="showFilters = !showFilters"
            class="text-xs px-2 py-1 rounded font-medium"
            :class="
              hasActiveFilters
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
                : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
            "
          >
            Filters
          </button>
          <button
            @click="emits('close')"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div v-if="showFilters" class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between mb-2">
            <span
              class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              >Filter by feeds &amp; labels</span
            >
            <button
              v-if="hasActiveFilters"
              @click="clearFilters"
              class="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              Clear filters
            </button>
          </div>
          <div class="max-h-48 overflow-y-auto space-y-1 text-sm">
            <div v-for="group in feedsStore.labelGroups" :key="group.label.id">
              <label
                class="flex items-center gap-2 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700/50 cursor-pointer"
              >
                <input
                  type="checkbox"
                  :checked="selectedLabelIds.has(group.label.id)"
                  @change="toggleLabel(group.label.id, labelFeedMap[group.label.id] || [])"
                  class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                />
                <span class="font-medium text-gray-700 dark:text-gray-300">{{
                  group.label.name
                }}</span>
              </label>
              <div class="ml-5 space-y-0.5">
                <label
                  v-for="feed in group.feeds"
                  :key="feed.id"
                  class="flex items-center gap-2 px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700/50 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    :checked="selectedFeedIds.has(feed.id)"
                    @change="toggleFeed(feed.id)"
                    class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-gray-600 dark:text-gray-400 truncate">{{
                    feed.title || feed.url
                  }}</span>
                </label>
              </div>
            </div>
            <div v-if="feedsStore.unlabeledFeeds.length > 0">
              <div class="ml-5 space-y-0.5">
                <label
                  v-for="feed in feedsStore.unlabeledFeeds"
                  :key="feed.id"
                  class="flex items-center gap-2 px-2 py-0.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700/50 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    :checked="selectedFeedIds.has(feed.id)"
                    @change="toggleFeed(feed.id)"
                    class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                  />
                  <span class="text-gray-600 dark:text-gray-400 truncate">{{
                    feed.title || feed.url
                  }}</span>
                </label>
              </div>
            </div>
            <p v-if="feedList.length === 0" class="px-2 py-1 text-gray-400 text-xs">
              No feeds yet.
            </p>
          </div>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto p-4">
        <div v-if="loading && results.length === 0" class="flex items-center justify-center py-12">
          <svg class="animate-spin h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24">
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </div>

        <div v-else-if="error" class="text-sm text-red-600 dark:text-red-400 text-center py-12">
          {{ error }}
        </div>

        <div v-else-if="!searched" class="text-sm text-gray-400 text-center py-12">
          Type to search items
        </div>

        <div v-else-if="results.length === 0" class="text-sm text-gray-400 text-center py-12">
          No results found for "{{ query }}"
        </div>

        <div v-else class="space-y-2">
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">
            {{ total }} result{{ total !== 1 ? "s" : "" }} for "{{ query }}"
          </p>
          <div
            v-for="item in results"
            :key="item.id"
            @click="navigateToItem(item)"
            class="px-3 py-2 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/30 cursor-pointer transition-colors"
          >
            <div class="flex items-start justify-between gap-2">
              <p
                class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate"
                v-html="highlightText(item.title, query)"
              />
              <span class="shrink-0 text-xs text-gray-500 dark:text-gray-400">
                {{ feedsStore.feedNames[item.feed_id] || "Unknown feed" }}
              </span>
            </div>
            <p
              v-if="item.description"
              class="mt-1 text-xs text-gray-500 dark:text-gray-400 line-clamp-2"
              v-html="highlightText(item.description, query)"
            />
          </div>

          <button
            v-if="hasMore"
            @click="loadMore"
            :disabled="loading"
            class="w-full py-2 text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 disabled:opacity-50"
          >
            {{ loading ? "Loading..." : "Load more" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
