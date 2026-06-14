<script setup lang="ts">
import { ref } from 'vue'
import { useFeedsStore } from '@/stores/feeds'

const emits = defineEmits<{
  close: []
}>()

const feedsStore = useFeedsStore()
const url = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  if (!url.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    await feedsStore.addFeed(url.value.trim())
    url.value = ''
    emits('close')
  } catch (e: any) {
    error.value = e.response?.data || 'Failed to add feed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
      <h2 class="text-lg font-semibold text-gray-900">Add Feed</h2>
      <form @submit.prevent="submit" class="mt-4">
        <input
          v-model="url"
          type="url"
          placeholder="https://example.com/rss"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm focus:ring-blue-500 focus:border-blue-500"
          autofocus
        />
        <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>
        <div class="mt-4 flex justify-end gap-3">
          <button
            type="button"
            @click="emits('close')"
            class="px-4 py-2 text-sm text-gray-700 hover:text-gray-900"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="loading || !url.trim()"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {{ loading ? 'Adding...' : 'Add Feed' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
