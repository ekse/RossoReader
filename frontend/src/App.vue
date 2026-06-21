<script setup lang="ts">
import FeedList from './components/FeedList.vue'
</script>

<template>
  <div class="flex h-screen">
    <aside class="w-72 shrink-0 border-r border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800 overflow-y-auto">
      <FeedList />
    </aside>
    <main class="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-900">
      <router-view />
    </main>
  </div>
</template>
