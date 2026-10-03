<script setup lang="ts">
import { computed, watch } from "vue";
import { ITEM_GROUPING_FEED, DEFAULT_ITEM_GROUPING, type Item, type ItemGrouping } from "@/types";
import ItemDetail from "./ItemDetail.vue";
import { useCurrentItem } from "@/composables/useCurrentItem";
import { useSearchHighlight } from "@/composables/useSearchHighlight";

const props = withDefaults(
  defineProps<{
    items: Item[];
    loading?: boolean;
    hasMore?: boolean;
    feedNames?: Record<number, string>;
    feedIcons?: Record<number, string>;
    groupBy?: ItemGrouping;
  }>(),
  { groupBy: DEFAULT_ITEM_GROUPING },
);

interface DisplayGroup {
  key: string;
  feedId?: number;
  icon?: string;
  name: string;
  showHeader: boolean;
  items: Item[];
}

const displayGroups = computed<DisplayGroup[]>(() => {
  if (props.groupBy !== ITEM_GROUPING_FEED) {
    return [{ key: "all", name: "", showHeader: false, items: props.items }];
  }

  const groups: DisplayGroup[] = [];
  const byFeed = new Map<number, DisplayGroup>();
  for (const item of props.items) {
    let group = byFeed.get(item.feed_id);
    if (!group) {
      group = {
        key: `feed-${item.feed_id}`,
        feedId: item.feed_id,
        icon: props.feedIcons?.[item.feed_id],
        name: props.feedNames?.[item.feed_id] || "Unknown feed",
        showHeader: true,
        items: [],
      };
      byFeed.set(item.feed_id, group);
      groups.push(group);
    }
    group.items.push(item);
  }
  return groups;
});

const showFeedName = computed(() => props.groupBy !== ITEM_GROUPING_FEED);

const emit = defineEmits<{
  toggleRead: [item: Item];
  toggleStarred: [item: Item];
  loadMore: [];
}>();

const { currentItemId, expandedItems, clearExpanded, isExpanded, pendingFocusItemId, expandItem } =
  useCurrentItem();

const { highlightText } = useSearchHighlight();

watch(
  () => props.items,
  () => {
    clearExpanded();
    const focusId = pendingFocusItemId.value;
    if (focusId !== null) {
      pendingFocusItemId.value = null;
      const item = props.items.find((i) => i.id === focusId);
      if (item) {
        expandItem(item.id);
        if (!item.read) {
          emit("toggleRead", item);
        }
      }
    }
  },
);

function toggleExpand(item: Item) {
  const itemId = item.id;
  const wasExpanded = isExpanded(itemId);

  if (wasExpanded) {
    expandedItems.value[itemId] = false;
    currentItemId.value = null;
  } else {
    expandedItems.value[itemId] = true;
    currentItemId.value = itemId;
    if (!item.read) {
      emit("toggleRead", item);
    }
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "";
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function stripHtml(s?: string): string {
  if (!s) return "";
  return s.replace(/<[^>]*>/g, "");
}
</script>

<template>
  <div>
    <div v-if="loading && items.length === 0" class="flex justify-center py-12">
      <div
        class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 dark:border-blue-400"
      />
    </div>

    <div v-else-if="items.length === 0" class="text-center py-12 text-gray-500 dark:text-gray-400">
      No items to show.
    </div>

    <div v-else>
      <template v-for="group in displayGroups" :key="group.key">
        <div
          v-if="group.showHeader"
          :data-feed-group="group.feedId"
          class="flex items-center gap-2 px-6 pt-2.5 pb-2 text-sm font-bold text-gray-700 dark:text-gray-100 bg-gray-100 dark:bg-gray-700"
        >
          <img
            v-if="group.icon"
            :src="group.icon"
            class="w-4 h-4 rounded shrink-0"
            alt=""
            loading="lazy"
          />
          {{ group.name }}
        </div>
        <div
          v-for="item in group.items"
          :key="item.id"
          :data-item-id="item.id"
          class="px-6 py-3 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
          :class="{
            'bg-white dark:bg-gray-800': !item.read && !isExpanded(item.id),
            'bg-gray-50 dark:bg-gray-800/30': item.read || isExpanded(item.id),
            'ring-2 ring-blue-400 dark:ring-blue-500 ring-inset':
              currentItemId === item.id && !isExpanded(item.id),
          }"
          @click="toggleExpand(item)"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="flex-1 min-w-0">
              <div
                class="flex flex-wrap items-baseline gap-x-2 text-xs text-gray-400 dark:text-gray-500"
              >
                <img
                  v-if="showFeedName && feedIcons?.[item.feed_id]"
                  :src="feedIcons[item.feed_id]"
                  class="hidden md:inline w-4 h-4 rounded shrink-0 self-center"
                  alt=""
                  loading="lazy"
                />
                <span v-if="showFeedName && feedNames?.[item.feed_id]" class="hidden md:inline">{{
                  feedNames[item.feed_id]
                }}</span>
                <h3 class="text-sm font-medium">
                  <span
                    v-html="highlightText(item.title)"
                    :class="[
                      item.read
                        ? 'text-gray-500 dark:text-gray-400'
                        : 'text-gray-900 dark:text-gray-100 hover:text-blue-600 dark:hover:text-blue-400',
                      'hover:underline',
                    ]"
                  />
                </h3>
              </div>
              <span
                v-if="showFeedName && !isExpanded(item.id) && feedNames?.[item.feed_id]"
                class="md:hidden mt-0.5 flex items-center gap-1 text-xs text-gray-400 dark:text-gray-500"
              >
                <img
                  v-if="feedIcons?.[item.feed_id]"
                  :src="feedIcons[item.feed_id]"
                  class="w-3.5 h-3.5 rounded shrink-0"
                  alt=""
                  loading="lazy"
                />
                {{ feedNames[item.feed_id] }}
              </span>
              <span
                v-if="item.description && !isExpanded(item.id)"
                class="text-sm text-gray-500 dark:text-gray-400 line-clamp-3 md:line-clamp-1"
                v-html="highlightText(stripHtml(item.description))"
              />
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="hidden md:inline text-xs text-gray-400 dark:text-gray-500">{{
                formatDate(item.published_at)
              }}</span>
              <button
                @click.stop="emit('toggleRead', item)"
                class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                :title="item.read ? 'Mark as unread' : 'Mark as read'"
              >
                <svg v-if="item.read" class="w-4 h-4 text-gray-400 dark:text-gray-500">
                  <use href="#icon-envelope-open" />
                </svg>
                <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500">
                  <use href="#icon-envelope" />
                </svg>
              </button>
              <button
                @click.stop="emit('toggleStarred', item)"
                class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                :title="item.starred ? 'Unsave' : 'Save'"
              >
                <svg v-if="item.starred" class="w-4 h-4 text-yellow-500">
                  <use href="#icon-bookmark-filled" />
                </svg>
                <svg v-else class="w-4 h-4 text-gray-400 dark:text-gray-500">
                  <use href="#icon-bookmark" />
                </svg>
              </button>
            </div>
          </div>
          <div
            v-if="isExpanded(item.id)"
            class="mt-2 border-t border-gray-200 dark:border-gray-700 pt-2 cursor-default"
            @click.stop
          >
            <ItemDetail :item="item" />
          </div>
        </div>
      </template>
    </div>

    <div v-if="!loading && items.length > 0 && hasMore" class="flex justify-center py-4">
      <button
        @click="emit('loadMore')"
        class="px-4 py-2 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
      >
        Load more
      </button>
    </div>
  </div>
</template>
