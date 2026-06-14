<script setup lang="ts">
import type { Item } from '@/types'

defineProps<{
  items: Item[]
  loading?: boolean
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
</script>

<template>
  <div>
    <div v-if="loading && items.length === 0" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
    </div>

    <div v-else-if="items.length === 0" class="text-center py-12 text-gray-500">
      No items to show.
    </div>

    <div v-else class="divide-y divide-gray-200">
      <div
        v-for="item in items"
        :key="item.id"
        class="px-6 py-4 hover:bg-gray-50 transition-colors"
        :class="{ 'bg-white': !item.read, 'bg-gray-50': item.read }"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="flex-1 min-w-0">
            <h3 class="text-sm font-medium">
              <a :href="item.url" target="_blank" rel="noopener noreferrer" :class="[item.read ? 'text-gray-500' : 'text-gray-900 hover:text-blue-600', 'hover:underline']">
                {{ item.title }}
              </a>
            </h3>
            <p v-if="item.description" class="mt-1 text-sm text-gray-500 line-clamp-2">
              {{ item.description }}
            </p>
            <div class="mt-1 flex items-center gap-3 text-xs text-gray-400">
              <span v-if="item.author">{{ item.author }}</span>
              <span>{{ formatDate(item.published_at) }}</span>
            </div>
          </div>
          <div class="flex items-center gap-2 shrink-0">
            <button
              @click="emit('toggleRead', item)"
              class="p-1 rounded hover:bg-gray-200 transition-colors"
              :title="item.read ? 'Mark as unread' : 'Mark as read'"
            >
              <svg v-if="item.read" class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255V19a2 2 0 01-2 2H5a2 2 0 01-2-2v-5.745m16 0A2 2 0 0019.586 12.55l-5.586-2.548-5.586 2.548A2 2 0 005 13.255m16 0l-4.5-2.625M5 13.255l4.5-2.625" />
              </svg>
              <svg v-else class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
            </button>
            <button
              @click="emit('toggleStarred', item)"
              class="p-1 rounded hover:bg-gray-200 transition-colors"
              :title="item.starred ? 'Unstar' : 'Star'"
            >
              <svg v-if="item.starred" class="w-4 h-4 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
              </svg>
              <svg v-else class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
              </svg>
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
        class="px-4 py-2 text-sm text-blue-600 hover:text-blue-800"
      >
        Load more
      </button>
    </div>
  </div>
</template>
