import { onMounted, onUnmounted, ref, computed, nextTick } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useItemsStore } from "@/stores/items";
import { useFeedsStore } from "@/stores/feeds";
import { useCurrentItem } from "./useCurrentItem";
import { useSearch } from "./useSearch";
import * as api from "@/api/client";

const showHelp = ref(false);

export function useKeyboardShortcuts() {
  const itemsStore = useItemsStore();
  const feedsStore = useFeedsStore();
  const router = useRouter();
  const route = useRoute();
  const { currentItemId, expandedItems, collapseItem, expandItem } = useCurrentItem();

  const currentItem = computed(() =>
    currentItemId.value !== null
      ? (itemsStore.items.find((i) => i.id === currentItemId.value) ?? null)
      : null,
  );

  async function scrollToItem(id: number) {
    await nextTick();
    const el = document.querySelector(`[data-item-id="${id}"]`);
    if (el) {
      el.scrollIntoView({ block: "start", behavior: "smooth" });
    }
  }

  function navigate(delta: number) {
    if (itemsStore.items.length === 0) return;

    const prevId = currentItemId.value;
    const wasExpanded = prevId !== null && expandedItems.value[prevId];
    if (wasExpanded) {
      collapseItem(prevId);
    }

    let idx = prevId !== null ? itemsStore.items.findIndex((i) => i.id === prevId) : -1;

    let newIdx: number;
    if (idx === -1) {
      newIdx = delta > 0 ? 0 : itemsStore.items.length - 1;
    } else {
      newIdx = idx + delta;
      if (newIdx < 0) newIdx = 0;
      if (newIdx >= itemsStore.items.length) {
        if (delta > 0 && itemsStore.hasMore) {
          itemsStore.loadMore().then(() => {
            if (newIdx < itemsStore.items.length) {
              currentItemId.value = itemsStore.items[newIdx].id;
              if (wasExpanded) {
                const target = itemsStore.items[newIdx];
                expandItem(target.id);
                if (!target.read) {
                  itemsStore.toggleRead(target);
                }
              }
              scrollToItem(currentItemId.value);
            }
          });
          return;
        }
        newIdx = itemsStore.items.length - 1;
      }
    }

    currentItemId.value = itemsStore.items[newIdx].id;
    if (wasExpanded) {
      const target = itemsStore.items[newIdx];
      expandItem(target.id);
      if (!target.read) {
        itemsStore.toggleRead(target);
      }
    }
    scrollToItem(currentItemId.value);

    if (newIdx >= itemsStore.items.length - 3 && itemsStore.hasMore && !itemsStore.loading) {
      itemsStore.loadMore();
    }
  }

  function navigateFeed(delta: number) {
    const feeds = feedsStore.orderedVisibleFeeds;
    if (feeds.length === 0) return;

    const currentId = route.params.id ? Number(route.params.id) : null;
    let idx = currentId !== null ? feeds.findIndex((f) => f.id === currentId) : -1;

    let newIdx: number;
    if (idx === -1) {
      newIdx = delta > 0 ? 0 : feeds.length - 1;
    } else {
      newIdx = idx + delta;
      if (newIdx < 0) newIdx = 0;
      if (newIdx >= feeds.length) newIdx = feeds.length - 1;
    }

    if (newIdx !== idx) {
      router.push(`/feed/${feeds[newIdx].id}`);
    }
  }

  function handler(e: KeyboardEvent) {
    const tag = (e.target as HTMLElement).tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

    switch (e.key) {
      case "F": {
        if (!showHelp.value && e.shiftKey) {
          useSearch().openSearch();
          e.preventDefault();
        }
        break;
      }
      case "/": {
        if (!showHelp.value) {
          useSearch().openSearch();
          e.preventDefault();
        }
        break;
      }
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
      case "j": {
        if (showHelp.value) break;
        navigate(1);
        e.preventDefault();
        break;
      }
      case "k": {
        if (showHelp.value) break;
        navigate(-1);
        e.preventDefault();
        break;
      }
      case "J": {
        if (showHelp.value) break;
        navigateFeed(1);
        e.preventDefault();
        break;
      }
      case "K": {
        if (showHelp.value) break;
        navigateFeed(-1);
        e.preventDefault();
        break;
      }
      case "N": {
        if (showHelp.value) break;
        router.push("/unread");
        e.preventDefault();
        break;
      }
      case "S": {
        if (showHelp.value) break;
        router.push("/starred");
        e.preventDefault();
        break;
      }
      case "Enter": {
        if (showHelp.value) break;
        if (currentItemId.value !== null && currentItem.value) {
          if (expandedItems.value[currentItemId.value]) {
            expandedItems.value[currentItemId.value] = false;
          } else {
            expandItem(currentItemId.value);
            if (!currentItem.value.read) {
              itemsStore.toggleRead(currentItem.value);
            }
          }
          scrollToItem(currentItemId.value);
          e.preventDefault();
        }
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
