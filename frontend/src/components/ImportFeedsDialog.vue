<script setup lang="ts">
import { ref, computed } from "vue";
import { useFeedsStore } from "@/stores/feeds";
import * as api from "@/api/client";
import type { DiscoveredFeed } from "@/types";

const emits = defineEmits<{
  close: [];
}>();

const feedsStore = useFeedsStore();
const file = ref<File | null>(null);
const loading = ref(false);
const error = ref("");
const previewFeeds = ref<DiscoveredFeed[]>([]);
const selected = ref<Set<number>>(new Set());
const importing = ref(false);
const importProgress = ref({ current: 0, total: 0 });
const importResult = ref<{ imported: number; skipped: number } | null>(null);

const allSelected = computed(
  () => previewFeeds.value.length > 0 && selected.value.size === previewFeeds.value.length,
);
const anySelected = computed(() => selected.value.size > 0);
const phase = computed(() => {
  if (importResult.value) return "done";
  if (importing.value) return "importing";
  if (previewFeeds.value.length > 0) return "preview";
  if (loading.value) return "parsing";
  return "upload";
});

function toggleAll() {
  if (allSelected.value) {
    selected.value = new Set();
  } else {
    selected.value = new Set(previewFeeds.value.map((_, i) => i));
  }
}

function toggleIndex(index: number) {
  const next = new Set(selected.value);
  if (next.has(index)) {
    next.delete(index);
  } else {
    next.add(index);
  }
  selected.value = next;
}

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  if (!input.files?.length) return;
  file.value = input.files[0];
  error.value = "";
  loading.value = true;
  previewFeeds.value = [];
  selected.value = new Set();
  try {
    const feeds = await api.previewOpmlImport(file.value);
    if (feeds.length === 0) {
      error.value = "No feeds found in the OPML file.";
    } else {
      previewFeeds.value = feeds;
      selected.value = new Set(feeds.map((_, i) => i));
    }
  } catch (e: any) {
    error.value = e.response?.data || "Failed to parse OPML file.";
  } finally {
    loading.value = false;
  }
}

async function handleImport() {
  const urls = Array.from(selected.value).map((i) => previewFeeds.value[i].url);
  importing.value = true;
  importProgress.value = { current: 0, total: urls.length };
  error.value = "";

  try {
    const result = await feedsStore.importFeeds(urls);
    importResult.value = result;
  } catch {
    error.value = "Import failed.";
  } finally {
    importing.value = false;
  }
}

function reset() {
  file.value = null;
  loading.value = false;
  error.value = "";
  previewFeeds.value = [];
  selected.value = new Set();
  importing.value = false;
  importProgress.value = { current: 0, total: 0 };
  importResult.value = null;
}

function handleClose() {
  reset();
  emits("close");
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-3xl mx-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Import Feeds</h2>

      <!-- Upload -->
      <div v-if="phase === 'upload'" class="mt-4">
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
          Select an OPML file to import feed subscriptions.
        </p>
        <input
          type="file"
          accept=".opml,.xml"
          @change="handleFileChange"
          class="block w-full text-sm text-gray-700 dark:text-gray-300 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-medium file:bg-blue-50 dark:file:bg-blue-900 file:text-blue-700 dark:file:text-blue-300 hover:file:bg-blue-100 dark:hover:file:bg-blue-800"
        />
        <p v-if="error" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        <div class="mt-4 flex justify-end">
          <button
            type="button"
            @click="handleClose"
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
          >
            Cancel
          </button>
        </div>
      </div>

      <!-- Parsing -->
      <div v-if="phase === 'parsing'" class="mt-4">
        <p class="text-sm text-gray-500 dark:text-gray-400">Parsing OPML file...</p>
      </div>

      <!-- Preview -->
      <div v-if="phase === 'preview'" class="mt-4">
        <div class="flex items-center justify-between mb-3">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Found {{ previewFeeds.length }} feed(s):
          </p>
          <button
            type="button"
            @click="toggleAll"
            class="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
          >
            {{ allSelected ? "Deselect All" : "Select All" }}
          </button>
        </div>
        <div class="space-y-2 max-h-60 overflow-y-auto">
          <div
            v-for="(feed, index) in previewFeeds"
            :key="feed.url"
            class="flex items-center px-3 py-2 rounded-md border border-gray-200 dark:border-gray-700"
          >
            <input
              type="checkbox"
              :checked="selected.has(index)"
              @change="toggleIndex(index)"
              class="mr-3 h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                {{ feed.title || feed.url }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ feed.url }}</p>
            </div>
          </div>
        </div>
        <p v-if="error" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        <div class="mt-4 flex justify-end gap-3">
          <button
            type="button"
            @click="reset"
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
          >
            Back
          </button>
          <button
            type="button"
            @click="handleClose"
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:hover:text-gray-100"
          >
            Cancel
          </button>
          <button
            type="button"
            :disabled="!anySelected"
            @click="handleImport"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            Import Selected ({{ selected.size }})
          </button>
        </div>
      </div>

      <!-- Importing -->
      <div v-if="phase === 'importing'" class="mt-4">
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Importing feeds... ({{ importProgress.current }} / {{ importProgress.total }})
        </p>
      </div>

      <!-- Done -->
      <div v-if="phase === 'done'" class="mt-4">
        <p class="text-sm text-green-600 dark:text-green-400">
          {{ importResult?.imported }} feed(s) imported
          <span v-if="importResult?.skipped"
            >({{ importResult?.skipped }} skipped — already subscribed)</span
          >.
        </p>
        <p v-if="error" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
        <div class="mt-4 flex justify-end gap-3">
          <button
            type="button"
            @click="handleClose"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
