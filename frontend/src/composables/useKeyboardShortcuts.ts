import { onMounted, onUnmounted, ref, computed } from "vue";
import { useItemsStore } from "@/stores/items";
import { useFeedsStore } from "@/stores/feeds";
import { useCurrentItem } from "./useCurrentItem";
import * as api from "@/api/client";

const showHelp = ref(false);

export function useKeyboardShortcuts() {
  const itemsStore = useItemsStore();
  const feedsStore = useFeedsStore();
  const { currentItemId } = useCurrentItem();

  const currentItem = computed(() =>
    currentItemId.value !== null
      ? itemsStore.items.find((i) => i.id === currentItemId.value) ?? null
      : null,
  );

  function handler(e: KeyboardEvent) {
    const tag = (e.target as HTMLElement).tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

    switch (e.key) {
      case "Escape": {
        showHelp.value = false;
        e.preventDefault();
        break;
      }
      case "h": {
        showHelp.value = !showHelp.value;
        e.preventDefault();
        break;
      }
      case "r":
      case "R": {
        if (showHelp.value) break;
        if (e.shiftKey) {
          markAllRead(itemsStore, feedsStore);
          e.preventDefault();
        } else if (currentItem.value) {
          itemsStore.toggleRead(currentItem.value);
          e.preventDefault();
        }
        break;
      }
      case "s": {
        if (showHelp.value) break;
        if (currentItem.value) {
          itemsStore.toggleStarred(currentItem.value);
          e.preventDefault();
        }
        break;
      }
    }
  }

  onMounted(() => window.addEventListener("keydown", handler));
  onUnmounted(() => window.removeEventListener("keydown", handler));

  function closeHelp() {
    showHelp.value = false;
  }

  return { showHelp, closeHelp };
}

async function markAllRead(
  itemsStore: ReturnType<typeof useItemsStore>,
  feedsStore: ReturnType<typeof useFeedsStore>,
) {
  if (itemsStore.filterFeedId !== undefined) {
    await api.markFeedRead(itemsStore.filterFeedId);
    itemsStore.loadItems();
    const feed = feedsStore.feeds.find((f) => f.id === itemsStore.filterFeedId);
    if (feed) feed.unread_count = 0;
  } else if (itemsStore.filterRead === false) {
    await api.markAllItemsRead();
    itemsStore.loadItems();
    for (const feed of feedsStore.feeds) {
      feed.unread_count = 0;
    }
  }
}
