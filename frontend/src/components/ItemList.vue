<script setup lang="ts">
import { ref } from "vue";
import type { Item } from "@/types";
import ItemDetail from "./ItemDetail.vue";

defineProps<{
  items: Item[];
  loading?: boolean;
  hasMore?: boolean;
  feedNames?: Record<number, string>;
}>();

const emit = defineEmits<{
  toggleRead: [item: Item];
  toggleStarred: [item: Item];
  loadMore: [];
}>();

const expandedItems = ref<Record<number, boolean>>({});

function toggleExpand(itemId: number) {
  expandedItems.value[itemId] = !expandedItems.value[itemId];
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "";
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function stripHtml(s?: string): string {
  if (!s) return "";
  return s.replace(/<[^>]*>/g, "");
}
</script>

<template>
  <div>
    <div v-if="loading && items.length === 0" class="flex justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 dark:border-blue-400"
      />
    </div>

    <div v-else-if="items.length === 0" class="text-center py-12 text-gray-500 dark:text-gray-400">
      No items to show.
    </div>

    <div v-else class="divide-y divide-gray-200 dark:divide-gray-700">
      <div
        v-for="item in items"
        :key="item.id"
        class="px-6 py-4 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
        :class="{
          'bg-white dark:bg-gray-800': !item.read && !expandedItems[item.id],
          'bg-gray-50 dark:bg-gray-800/30': item.read || expandedItems[item.id],
        }"
        @click="toggleExpand(item.id)"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="flex-1 min-w-0">
            <div
              class="flex flex-wrap items-baseline gap-x-2 text-xs text-gray-400 dark:text-gray-500"
            >
              <span v-if="feedNames?.[item.feed_id]" class="hidden md:inline">{{
                feedNames[item.feed_id]
              }}</span>
              <span class="hidden md:inline">{{ formatDate(item.published_at) }}</span>
              <h3 class="text-sm font-medium">
                <span
                  :class="[
                    item.read
                      ? 'text-gray-500 dark:text-gray-400'
                      : 'text-gray-900 dark:text-gray-100 hover:text-blue-600 dark:hover:text-blue-400',
                    'hover:underline',
                  ]"
                >
                  {{ item.title }}
                </span>
              </h3>
            </div>
            <span
              v-if="!expandedItems[item.id] && feedNames?.[item.feed_id]"
              class="md:hidden mt-0.5 text-xs text-gray-400 dark:text-gray-500"
              >{{ feedNames[item.feed_id] }}</span
            >
            <span
              v-if="item.description && !expandedItems[item.id]"
              class="text-sm text-gray-500 dark:text-gray-400 line-clamp-3 md:line-clamp-1"
            >
              {{ stripHtml(item.description) }}
            </span>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button
              @click.stop="emit('toggleRead', item)"
              class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              :title="item.read ? 'Mark as unread' : 'Mark as read'"
            >
              <svg v-if="item.read" class="w-4 h-4 text-gray-400 dark:text-gray-500">
                <use href="#icon-envelope-open" />
              </svg>
              <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500">
                <use href="#icon-envelope" />
              </svg>
            </button>
            <button
              @click.stop="emit('toggleStarred', item)"
              class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              :title="item.starred ? 'Unstar' : 'Star'"
            >
              <svg v-if="item.starred" class="w-4 h-4 text-yellow-500">
                <use href="#icon-star-filled" />
              </svg>
              <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500">
                <use href="#icon-star" />
              </svg>
            </button>
          </div>
        </div>
        <div
          v-if="expandedItems[item.id]"
          class="mt-2 border-t border-gray-200 dark:border-gray-700 pt-2 cursor-default"
          @click.stop
        >
          <ItemDetail :item="item" />
        </div>
      </div>
    </div>

    <div v-if="!loading && items.length > 0 && hasMore" class="flex justify-center py-4">
      <button
        @click="emit('loadMore')"
        class="px-4 py-2 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
      >
        Load more
      </button>
    </div>
  </div>
</template>
