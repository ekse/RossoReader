<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import * as api from "@/api/client";
import type { Label } from "@/types";

const props = defineProps<{ feedId: number }>();
const emit = defineEmits<{ close: [] }>();

const dropdown = ref<HTMLElement | null>(null);
const allLabels = ref<Label[]>([]);
const feedLabelIds = ref(new Set<number>());
const newLabelName = ref("");
const creating = ref(false);

function onClickOutside(e: MouseEvent) {
  if (dropdown.value && !dropdown.value.contains(e.target as Node)) {
    emit("close");
  }
}

onMounted(async () => {
  document.addEventListener("mousedown", onClickOutside);
  const [labels, feedLabels] = await Promise.all([
    api.fetchLabels(),
    api.fetchFeedLabels(props.feedId),
  ]);
  allLabels.value = labels;
  feedLabelIds.value = new Set(feedLabels.map((l) => l.id));
});

onUnmounted(() => {
  document.removeEventListener("mousedown", onClickOutside);
});

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
</script>

<template>
  <div
    ref="dropdown"
    class="absolute left-0 top-full mt-1 z-20 w-64 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700"
  >
    <div class="p-3">
      <h4 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">Labels</h4>
      <div class="max-h-48 overflow-y-auto space-y-1">
        <label
          v-for="label in allLabels"
          :key="label.id"
          class="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer text-sm"
          @click.prevent="toggleLabel(label.id)"
        >
          <input
            type="checkbox"
            :checked="feedLabelIds.has(label.id)"
            class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
          />
          <span class="text-gray-700 dark:text-gray-300">{{ label.name }}</span>
        </label>
        <p v-if="allLabels.length === 0" class="text-xs text-gray-500 dark:text-gray-400 px-2 py-1">
          No labels yet. Create one below.
        </p>
      </div>
      <div class="mt-2 pt-2 border-t border-gray-200 dark:border-gray-700">
        <form @submit.prevent="createLabel" class="flex gap-1">
          <input
            v-model="newLabelName"
            placeholder="New label name"
            class="flex-1 px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-blue-500 focus:border-blue-500"
          />
          <button
            type="submit"
            :disabled="creating || !newLabelName.trim()"
            class="px-2 py-1 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {{ creating ? "..." : "Add" }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
