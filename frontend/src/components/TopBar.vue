<script setup lang="ts">
import { useSidebar } from "@/composables/useSidebar";
import { useHeader } from "@/composables/useHeader";

defineProps<{ title?: string; showMarkAllRead?: boolean }>();
defineEmits<{ markAllRead: [] }>();

const { toggle } = useSidebar();
const { isHeaderVisible } = useHeader();
</script>

<template>
  <div
    class="sticky top-0 z-10 px-4 py-2 md:px-6 md:py-2 border-b border-gray-200 dark:border-gray-700 bg-white/95 dark:bg-gray-800/30 backdrop-blur-sm flex items-center justify-between transition-transform duration-300 ease-in-out md:translate-y-0"
    :class="isHeaderVisible ? 'translate-y-0' : '-translate-y-full'"
  >
    <div class="flex items-center gap-3">
      <button
        @click="toggle"
        class="p-1.5 -ml-1 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-700 md:hidden transition-colors"
        title="Toggle Sidebar"
      >
        <svg class="w-5 h-5"><use href="#icon-menu" /></svg>
      </button>
      <h2 v-if="title" class="text-sm font-semibold text-gray-900 dark:text-gray-100">
        {{ title }}
      </h2>
      <slot name="left-actions" />
    </div>
    <div class="flex items-center gap-2">
      <slot name="actions" />
      <button
        v-if="showMarkAllRead"
        @click="$emit('markAllRead')"
        class="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 dark:bg-blue-800 dark:hover:bg-blue-900"
      >
        Mark all as read
      </button>
    </div>
  </div>
</template>
