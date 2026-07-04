<script setup lang="ts">
import { onMounted, onUnmounted, watch } from "vue";
import { useRoute } from "vue-router";
import FeedList from "./components/FeedList.vue";
import AddFeedDialog from "./components/AddFeedDialog.vue";
import KeyboardShortcutsDialog from "./components/KeyboardShortcutsDialog.vue";
import SvgSprites from "./components/SvgSprites.vue";
import { useSidebar, DEFAULT_WIDTH } from "./composables/useSidebar";
import { useHeader } from "./composables/useHeader";
import { useAddFeed } from "./composables/useAddFeed";
import { useKeyboardShortcuts } from "./composables/useKeyboardShortcuts";
import { useFeedsStore } from "./stores/feeds";

const route = useRoute();
const { isOpen, close, sidebarWidth, startResize } = useSidebar();
const { isHeaderVisible, handleScroll } = useHeader();
const { showAddFeed, close: closeAddFeed } = useAddFeed();
const { showHelp, closeHelp } = useKeyboardShortcuts();
const feedsStore = useFeedsStore();

watch(
  () => feedsStore.totalUnread,
  (n) => {
    document.title = n > 0 ? `(${n}) Rosso Reader` : "Rosso Reader";
  },
  { immediate: true },
);

let wasMobile = false;

function handleResize() {
  const isMobile = window.innerWidth < 768;
  if (wasMobile && !isMobile) {
    isOpen.value = true;
  }
  wasMobile = isMobile;
}

onMounted(() => {
  wasMobile = window.innerWidth < 768;
  if (wasMobile) {
    isOpen.value = false;
  }
  window.addEventListener("resize", handleResize);
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
});

// Close sidebar on mobile when navigating
watch(
  () => route.path,
  () => {
    if (window.innerWidth < 768) {
      close();
    }
    isHeaderVisible.value = true;
  },
);
</script>

<template>
  <div class="flex h-screen overflow-hidden">
    <SvgSprites />
    <!-- Backdrop overlay for mobile screen when sidebar is open -->
    <Transition name="fade">
      <div
        v-if="isOpen && route.meta.public !== true"
        @click="close"
        class="fixed inset-0 z-30 bg-gray-900/50 backdrop-blur-sm md:hidden"
      ></div>
    </Transition>

    <aside
      v-if="route.meta.public !== true"
      :style="isOpen && sidebarWidth !== DEFAULT_WIDTH ? { width: sidebarWidth + 'px' } : {}"
      class="fixed inset-y-0 left-0 z-40 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 overflow-y-auto overflow-x-hidden transform transition-transform duration-300 ease-in-out md:static md:translate-x-0"
      :class="[
        isOpen
          ? 'w-72 translate-x-0'
          : 'w-72 -translate-x-full md:!w-0 md:border-r-0 md:translate-x-0',
      ]"
    >
      <FeedList />
      <div
        v-if="isOpen"
        @mousedown="startResize"
        class="absolute right-0 top-0 bottom-0 w-1.5 cursor-col-resize z-50 group"
      >
        <div
          class="absolute inset-y-0 right-0 w-0.5 group-hover:bg-gray-400 dark:group-hover:bg-gray-500 group-active:bg-blue-500/60 transition-colors"
        ></div>
      </div>
    </aside>
    <main class="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-900" @scroll="handleScroll">
      <router-view />
    </main>

    <AddFeedDialog v-if="showAddFeed" @close="closeAddFeed" />
    <KeyboardShortcutsDialog v-if="showHelp" @close="closeHelp" />
  </div>
</template>
