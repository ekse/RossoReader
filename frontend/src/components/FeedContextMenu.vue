<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useFeedsStore } from "@/stores/feeds";
import * as api from "@/api/client";
import type { Label } from "@/types";

const props = defineProps<{
  feedId: number;
  x: number;
  y: number;
}>();

const emit = defineEmits<{ close: [] }>();

const feedsStore = useFeedsStore();
const router = useRouter();

const menu = ref<HTMLElement | null>(null);
const menuStyle = ref<Record<string, string>>({});
const showLabelPicker = ref(false);
const allLabels = ref<Label[]>([]);
const feedLabelIds = ref(new Set<number>());
const newLabelName = ref("");
const creating = ref(false);

function onClickOutside(e: MouseEvent) {
  if (menu.value && !menu.value.contains(e.target as Node)) {
    emit("close");
  }
}

function onKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    emit("close");
  }
}

onMounted(() => {
  document.addEventListener("mousedown", onClickOutside);
  document.addEventListener("keydown", onKeyDown);

  const style: Record<string, string> = {};
  const padding = 8;
  const approxWidth = 256;
  const approxHeight = 160;

  if (props.x + approxWidth > window.innerWidth) {
    style.left = `${Math.max(padding, window.innerWidth - approxWidth - padding)}px`;
  } else {
    style.left = `${props.x}px`;
  }

  if (props.y + approxHeight > window.innerHeight) {
    style.top = `${Math.max(padding, window.innerHeight - approxHeight - padding)}px`;
  } else {
    style.top = `${props.y}px`;
  }

  menuStyle.value = style;
});

onUnmounted(() => {
  document.removeEventListener("mousedown", onClickOutside);
  document.removeEventListener("keydown", onKeyDown);
});

async function markAllRead() {
  await api.markFeedRead(props.feedId);
  await feedsStore.loadGroupedFeeds();
  emit("close");
}

async function unsubscribe() {
  await feedsStore.removeFeed(props.feedId);
  if (router.currentRoute.value.params.id === String(props.feedId)) {
    router.push("/unread");
  }
  emit("close");
}

async function openLabelPicker() {
  showLabelPicker.value = true;
  const [labels, feedLabels] = await Promise.all([
    api.fetchLabels(),
    api.fetchFeedLabels(props.feedId),
  ]);
  allLabels.value = labels;
  feedLabelIds.value = new Set(feedLabels.map((l) => l.id));
}

async function toggleLabel(labelId: number) {
  const s = new Set(feedLabelIds.value);
  if (s.has(labelId)) {
    s.delete(labelId);
    await api.removeFeedLabel(props.feedId, labelId);
  } else {
    s.add(labelId);
    await api.addFeedLabel(props.feedId, labelId);
  }
  feedLabelIds.value = s;
}

async function createLabel() {
  const name = newLabelName.value.trim();
  if (!name) return;
  creating.value = true;
  try {
    const label = await api.createLabel(name);
    allLabels.value.push(label);
    feedLabelIds.value = new Set([...feedLabelIds.value, label.id]);
    newLabelName.value = "";
  } finally {
    creating.value = false;
  }
}

function backToMenu() {
  showLabelPicker.value = false;
}
</script>

<template>
  <Teleport to="body">
  <div
    ref="menu"
    class="fixed z-50 w-64 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1"
    :style="menuStyle"
  >
    <template v-if="!showLabelPicker">
      <button
        @click="markAllRead"
        class="w-full text-left px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
      >
        <svg class="w-4 h-4 shrink-0"><use href="#icon-check" /></svg>
        Mark all as read
      </button>
      <button
        @click="openLabelPicker"
        class="w-full text-left px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
      >
        <svg class="w-4 h-4 shrink-0"><use href="#icon-tag" /></svg>
        Edit labels
      </button>
      <hr class="my-1 border-gray-200 dark:border-gray-700" />
      <button
        @click="unsubscribe"
        class="w-full text-left px-3 py-1.5 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 flex items-center gap-2"
      >
        <svg class="w-4 h-4 shrink-0"><use href="#icon-trash" /></svg>
        Unsubscribe
      </button>
    </template>
    <template v-else>
      <div class="px-3 py-2">
        <div class="flex items-center justify-between mb-2">
          <button
            @click="backToMenu"
            class="text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
          >
            &larr; Back
          </button>
          <span class="text-xs font-semibold text-gray-500 dark:text-gray-400">Labels</span>
          <div class="w-4" />
        </div>
        <div class="max-h-32 overflow-y-auto space-y-1 mb-2">
          <label
            v-for="label in allLabels"
            :key="label.id"
            class="flex items-center gap-2 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer text-sm"
            @click.prevent="toggleLabel(label.id)"
          >
            <input
              type="checkbox"
              :checked="feedLabelIds.has(label.id)"
              class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
            />
            <span class="text-gray-700 dark:text-gray-300 truncate">{{ label.name }}</span>
          </label>
          <p
            v-if="allLabels.length === 0"
            class="text-xs text-gray-500 dark:text-gray-400 px-2 py-1"
          >
            No labels yet.
          </p>
        </div>
        <form @submit.prevent="createLabel" class="flex gap-1">
          <input
            v-model="newLabelName"
            placeholder="New label"
            class="flex-1 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <button
            type="submit"
            :disabled="creating || !newLabelName.trim()"
            class="px-2 py-1 text-xs font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {{ creating ? "..." : "Add" }}
          </button>
        </form>
      </div>
    </template>
  </div>
  </Teleport>
</template>
