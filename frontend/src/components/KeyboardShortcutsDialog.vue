<script setup lang="ts">
defineEmits<{ close: [] }>();
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    @click.self="$emit('close')"
  >
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-lg mx-4">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Keyboard Shortcuts</h2>
        <button
          @click="$emit('close')"
          class="p-1 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Next item</span>
          <kbd
            class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
            >j</kbd
          >
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Previous item</span>
          <kbd
            class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
            >k</kbd
          >
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Toggle open/close</span>
          <kbd
            class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
            >&#8629;</kbd
          >
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Toggle read</span>
          <kbd
            class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
            >r</kbd
          >
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Mark all as read</span>
          <div class="flex items-center gap-1">
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >⇧</kbd
            >
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >r</kbd
            >
          </div>
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Toggle starred</span>
          <kbd
            class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
            >s</kbd
          >
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Next feed</span>
          <div class="flex items-center gap-1">
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >⇧</kbd
            >
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >j</kbd
            >
          </div>
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Previous feed</span>
          <div class="flex items-center gap-1">
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >⇧</kbd
            >
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >k</kbd
            >
          </div>
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Go to New</span>
          <div class="flex items-center gap-1">
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >⇧</kbd
            >
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >n</kbd
            >
          </div>
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Go to Starred</span>
          <div class="flex items-center gap-1">
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >⇧</kbd
            >
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >s</kbd
            >
          </div>
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Search</span>
          <div class="flex items-center gap-1">
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >⇧</kbd
            >
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >f</kbd
            >
            <span class="text-xs text-gray-400 dark:text-gray-500 mx-1">or</span>
            <kbd
              class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
              >/</kbd
            >
          </div>
        </div>
        <div
          class="flex items-center justify-between py-2 px-3 rounded-md bg-gray-50 dark:bg-gray-700/50"
        >
          <span class="text-sm text-gray-700 dark:text-gray-300">Show shortcuts</span>
          <kbd
            class="px-2 py-0.5 text-xs font-mono font-semibold bg-white dark:bg-gray-600 border border-gray-300 dark:border-gray-500 rounded shadow-sm"
            >h</kbd
          >
        </div>
      </div>

      <div class="mt-6 flex justify-end">
        <button
          @click="$emit('close')"
          class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-md hover:bg-gray-200 dark:hover:bg-gray-600"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>
