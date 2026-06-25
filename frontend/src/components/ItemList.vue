<script setup lang="ts">
import type { Item } from '@/types'

defineProps<{
  items: Item[]
  loading?: boolean
  feedNames?: Record<number, string>
}>()

const emit = defineEmits<{
  toggleRead: [item: Item]
  toggleStarred: [item: Item]
  loadMore: []
}>()

function formatDate(dateStr?: string): string {
	if (!dateStr) return ''
	return new Date(dateStr).toLocaleDateString('en-US', {
		month: 'short', day: 'numeric', year: 'numeric',
	})
}

function stripHtml(s?: string): string {
	if (!s) return ''
	return s.replace(/<[^>]*>/g, '')
}
</script>

<template>
  <div>
    <div v-if="loading && items.length === 0" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 dark:border-blue-400" />
    </div>

    <div v-else-if="items.length === 0" class="text-center py-12 text-gray-500 dark:text-gray-400">
      No items to show.
    </div>

    <div v-else class="divide-y divide-gray-200 dark:divide-gray-700">
      <div
        v-for="item in items"
        :key="item.id"
        class="px-6 py-4 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
        :class="{ 'bg-white dark:bg-gray-800': !item.read, 'bg-gray-50 dark:bg-gray-800/30': item.read }"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="flex-1 min-w-0">
            <h3 class="text-sm font-medium">
              <a :href="item.url" target="_blank" rel="noopener noreferrer" :class="[item.read ? 'text-gray-500 dark:text-gray-400' : 'text-gray-900 dark:text-gray-100 hover:text-blue-600 dark:hover:text-blue-400', 'hover:underline']">
                {{ item.title }}
              </a>
            </h3>
            <p v-if="item.description" class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
              {{ stripHtml(item.description) }}
            </p>
            <div class="mt-1 flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
              <span v-if="feedNames?.[item.feed_id]">{{ feedNames[item.feed_id] }}</span>
              <span>{{ formatDate(item.published_at) }}</span>
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button
              @click="emit('toggleRead', item)"
              class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              :title="item.read ? 'Mark as unread' : 'Mark as read'"
            >
              <svg v-if="item.read" class="w-4 h-4 text-gray-400 dark:text-gray-500"><use href="#icon-envelope-open" /></svg>
              <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500"><use href="#icon-envelope" /></svg>
            </button>
            <button
              @click="emit('toggleStarred', item)"
              class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              :title="item.starred ? 'Unstar' : 'Star'"
            >
              <svg v-if="item.starred" class="w-4 h-4 text-yellow-500"><use href="#icon-star-filled" /></svg>
              <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500"><use href="#icon-star" /></svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="!loading && items.length > 0"
      class="flex justify-center py-4"
    >
      <button
        @click="emit('loadMore')"
        class="px-4 py-2 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
      >
        Load more
      </button>
    </div>
  </div>
</template>
