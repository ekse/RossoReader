<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useFeedsStore } from '@/stores/feeds'
import AddFeedDialog from '@/components/AddFeedDialog.vue'

const feedsStore = useFeedsStore()
const showAddFeed = ref(false)

onMounted(() => {
  feedsStore.loadFeeds()
})
</script>

<template>
  <div class="max-w-3xl mx-auto px-6 py-8">
    <h1 class="text-2xl font-bold text-gray-900">Settings</h1>

    <section class="mt-8">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Feed Subscriptions</h2>
        <button
          @click="showAddFeed = true"
          class="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
        >
          + Add Feed
        </button>
      </div>

      <div class="space-y-2">
        <div
          v-for="feed in feedsStore.feeds"
          :key="feed.id"
          class="flex items-center justify-between px-4 py-3 bg-white rounded-lg border border-gray-200"
        >
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-gray-900 truncate">{{ feed.title || feed.url }}</p>
            <p class="text-xs text-gray-500 truncate">{{ feed.url }}</p>
          </div>
          <button
            @click="feedsStore.removeFeed(feed.id)"
            class="ml-4 text-sm text-red-600 hover:text-red-800"
          >
            Remove
          </button>
        </div>
      </div>

      <p v-if="feedsStore.feeds.length === 0" class="text-sm text-gray-500 mt-4">
        No feeds subscribed yet. Add one to get started!
      </p>
    </section>

    <AddFeedDialog v-if="showAddFeed" @close="showAddFeed = false" />
  </div>
</template>
