<script setup lang="ts">
import { ref, watch, nextTick } from "vue";
import type { Item } from "@/types";
import { useSearchHighlight, highlightTextNodes } from "@/composables/useSearchHighlight";
import ItemActions from "./ItemActions.vue";

const props = defineProps<{
  item: Item | null;
}>();

const emit = defineEmits<{
  close: [];
  toggleRead: [item: Item];
  toggleStarred: [item: Item];
}>();

const { highlightQuery } = useSearchHighlight();

const articleRef = ref<HTMLElement | null>(null);

watch(
  () => [props.item, highlightQuery.value] as const,
  async () => {
    await nextTick();
    const q = highlightQuery.value;
    if (!q || !articleRef.value) return;
    highlightTextNodes(articleRef.value, q);
  },
  { immediate: true },
);

function formatDate(dateStr?: string): string {
  if (!dateStr) return "";
  return new Date(dateStr).toLocaleString();
}
</script>

<template>
  <div
    v-if="!item"
    class="flex items-center justify-center h-full text-gray-400 dark:text-gray-500"
  >
    Select an item to read
  </div>
  <article v-else ref="articleRef" class="max-w-3xl mx-auto px-6 py-2">
    <div
      data-close-region
      class="h-6 -mt-2 cursor-pointer"
      title="Close article"
      @click="emit('close')"
    />
    <div class="flex items-start justify-between gap-4">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ item.title }}</h1>
      <button
        data-close-button
        class="shrink-0 p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-200 dark:text-gray-500 dark:hover:text-gray-300 dark:hover:bg-gray-700 transition-colors"
        title="Close article"
        @click="emit('close')"
      >
        <svg class="w-5 h-5">
          <use href="#icon-x" />
        </svg>
      </button>
    </div>
    <div class="mt-2 flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
      <span v-if="item.author">By {{ item.author }}</span>
      <span>{{ formatDate(item.published_at) }}</span>
      <a
        v-if="item.url"
        :href="item.url"
        target="_blank"
        class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
      >
        View original
      </a>
    </div>
    <div class="mt-2 flex items-center gap-2">
      <ItemActions
        :item="item"
        @toggle-read="emit('toggleRead', item)"
        @toggle-starred="emit('toggleStarred', item)"
      />
    </div>
    <div
      v-if="item.content"
      class="mt-6 prose prose-sm max-w-none text-gray-700 dark:text-gray-300 dark:prose-invert"
      v-html="item.content"
    />
    <div
      v-else-if="item.description"
      class="mt-6 prose prose-sm max-w-none text-gray-700 dark:text-gray-300 dark:prose-invert"
      v-html="item.description"
    />
  </article>
</template>
