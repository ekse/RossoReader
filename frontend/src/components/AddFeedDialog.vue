<script setup lang="ts">
import { ref } from "vue";
import { useFeedsStore } from "@/stores/feeds";
import * as api from "@/api/client";
import type { DiscoveredFeed } from "@/types";

const emits = defineEmits<{
  close: [];
}>();

const feedsStore = useFeedsStore();
const url = ref("");
const loading = ref(false);
const adding = ref<number | null>(null);
const error = ref("");
const discovered = ref<DiscoveredFeed[]>([]);

async function submit() {
  if (!url.value.trim()) return;
  loading.value = true;
  error.value = "";
  discovered.value = [];
  try {
    const feeds = await api.discoverFeeds(url.value.trim());
    if (feeds.length === 0) {
      error.value = "No RSS feeds found at this URL.";
    } else {
      discovered.value = feeds;
    }
  } catch (e: any) {
    error.value = e.response?.data || "Failed to discover feeds";
  } finally {
    loading.value = false;
  }
}

async function addSelected(feed: DiscoveredFeed, index: number) {
  adding.value = index;
  try {
    await feedsStore.addFeed(feed.url);
    discovered.value.splice(index, 1);
    if (discovered.value.length === 0) {
      url.value = "";
      emits("close");
    }
  } catch (e: any) {
    error.value = e.response?.data || "Failed to add feed";
  } finally {
    adding.value = null;
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-3xl mx-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Add Feed</h2>

      <form v-if="discovered.length === 0" @submit.prevent="submit" class="mt-4">
        <input
          v-model="url"
          type="url"
          placeholder="https://example.com or https://example.com/rss"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          autofocus
        />
        <p v-if="error" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        <div class="mt-4 flex justify-end gap-3">
          <button
            type="button"
            @click="emits('close')"
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="loading || !url.trim()"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {{ loading ? "Discovering..." : "Discover" }}
          </button>
        </div>
      </form>

      <div v-else class="mt-4">
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
          Found {{ discovered.length }} feed(s):
        </p>
        <div class="space-y-2 max-h-60 overflow-y-auto">
          <div
            v-for="(feed, index) in discovered"
            :key="feed.url"
            class="flex items-center justify-between px-3 py-2 rounded-md border border-gray-200 dark:border-gray-700"
          >
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                {{ feed.title || feed.url }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ feed.url }}</p>
            </div>
            <button
              @click="addSelected(feed, index)"
              :disabled="adding !== null"
              class="ml-3 shrink-0 px-3 py-1 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {{ adding === index ? "Adding..." : "Add" }}
            </button>
          </div>
        </div>
        <p v-if="error" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        <div class="mt-4 flex justify-end gap-3">
          <button
            type="button"
            @click="
              discovered = [];
              error = '';
            "
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
          >
            Back
          </button>
          <button
            type="button"
            @click="emits('close')"
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
