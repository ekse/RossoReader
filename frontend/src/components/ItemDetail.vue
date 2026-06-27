<script setup lang="ts">
import type { Item } from '@/types'

defineProps<{
  item: Item | null
}>()

function formatDate(dateStr?: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString()
}
</script>

<template>
  <div v-if="!item" class="flex items-center justify-center h-full text-gray-400 dark:text-gray-500">
    Select an item to read
  </div>
  <article v-else class="max-w-3xl mx-auto px-6 py-8">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ item.title }}</h1>
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
    <div
      v-if="item.content"
      class="mt-6 prose prose-sm max-w-none text-gray-700 dark:text-gray-300 dark:prose-invert"
      v-html="item.content"
    />
    <div
      v-else-if="item.description"
      class="mt-6 prose prose-sm max-w-none text-gray-700 dark:text-gray-300 dark:prose-invert"
    >
      {{ item.description }}
    </div>
  </article>
</template>
