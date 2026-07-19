<script setup lang="ts">
import { useTheme } from "@/composables/useTheme";

const { isDark, toggle } = useTheme();
</script>

<template>
  <button
    @click="toggle"
    class="p-2 rounded-md transition-colors hover:bg-gray-100 dark:hover:bg-gray-700"
    :title="isDark ? 'Switch to light theme' : 'Switch to dark theme'"
  >
    <!-- sun -->
    <svg v-if="!isDark" class="w-5 h-5 text-yellow-500"><use href="#icon-sun" /></svg>
    <!-- moon -->
    <svg v-else class="w-5 h-5 text-yellow-400"><use href="#icon-moon" /></svg>
  </button>
</template>
