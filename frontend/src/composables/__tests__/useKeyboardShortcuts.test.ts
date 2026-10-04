import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import { setActivePinia, createPinia } from "pinia";

const mockPush = vi.hoisted(() => vi.fn());

vi.mock("@/stores/items", () => ({
  useItemsStore: () => ({
    items: [],
    filterFeedId: undefined,
    filterRead: undefined,
    hasMore: false,
    loading: false,
    loadMore: vi.fn(),
    loadItems: vi.fn(),
    toggleRead: vi.fn(),
    toggleStarred: vi.fn(),
  }),
}));

vi.mock("@/stores/feeds", () => ({
  useFeedsStore: () => ({
    feeds: [],
    orderedVisibleFeeds: [],
  }),
}));

vi.mock("vue-router", () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: {} }),
}));

vi.mock("@/api/client", () => ({}));

import { useKeyboardShortcuts } from "../useKeyboardShortcuts";

const Harness = defineComponent({
  setup() {
    useKeyboardShortcuts();
    return () => h("div");
  },
});

function press(init: KeyboardEventInit) {
  window.dispatchEvent(new KeyboardEvent("keydown", { ...init, cancelable: true }));
}

describe("useKeyboardShortcuts", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("triggers shortcut for a plain key", () => {
    mount(Harness);
    press({ key: "S", shiftKey: true });
    expect(mockPush).toHaveBeenCalledWith("/starred");
  });

  it("does not trigger shortcut when Ctrl is held", () => {
    mount(Harness);
    press({ key: "S", shiftKey: true, ctrlKey: true });
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("does not trigger shortcut when Meta is held", () => {
    mount(Harness);
    press({ key: "S", shiftKey: true, metaKey: true });
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("does not trigger shortcut when Alt is held", () => {
    mount(Harness);
    press({ key: "S", shiftKey: true, altKey: true });
    expect(mockPush).not.toHaveBeenCalled();
  });
});
