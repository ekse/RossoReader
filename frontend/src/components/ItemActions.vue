<script setup lang="ts">
import type { Item } from "@/types";

defineProps<{
  item: Item;
}>();

const emit = defineEmits<{
  toggleRead: [item: Item];
  toggleStarred: [item: Item];
}>();
</script>

<template>
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
    :title="item.starred ? 'Unsave' : 'Save'"
  >
    <svg v-if="item.starred" class="w-4 h-4 text-yellow-500">
      <use href="#icon-bookmark-filled" />
    </svg>
    <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500">
      <use href="#icon-bookmark" />
    </svg>
  </button>
</template>
