import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import { setActivePinia, createPinia } from "pinia";

vi.mock("@/stores/items", () => ({
  useItemsStore: () => ({
    items: [],
    loading: false,
    hasMore: false,
    setFilterFeedId: vi.fn(),
    loadItems: vi.fn(),
    loadMore: vi.fn(),
    toggleRead: vi.fn(),
    toggleStarred: vi.fn(),
  }),
}));

vi.mock("@/stores/feeds", () => ({
  useFeedsStore: () => ({
    feeds: [],
    feedNames: { 1: "Example Feed" },
    feedIcons: { 1: "https://ex.com/icon.png" },
  }),
}));

vi.mock("@/api/client", () => ({
  markFeedRead: vi.fn(),
}));

vi.mock("@/composables/useSidebar", () => ({
  useSidebar: () => ({ toggle: vi.fn() }),
}));

vi.mock("@/composables/useHeader", () => ({
  useHeader: () => ({ isHeaderVisible: { value: true } }),
}));

vi.mock("vue-router", () => ({
  useRoute: () => ({ params: { id: "1" } }),
}));

import FeedItems from "@/views/FeedItems.vue";

describe("FeedItems", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("shows the feed icon and name before the label button", async () => {
    const wrapper = mount(FeedItems, {
      global: {
        stubs: {
          TopBar: { template: '<div><slot name="left-actions" /></div>' },
          ItemList: true,
          LabelPicker: true,
        },
      },
    });
    await nextTick();

    const identity = wrapper.find(".flex.items-center.gap-2");
    expect(identity.text()).toContain("Example Feed");
    expect(identity.find('img[src="https://ex.com/icon.png"]').exists()).toBe(true);
    expect(wrapper.find('button[aria-label="Edit labels"]').exists()).toBe(true);
  });

  it("shows the edit labels button as a tag icon", async () => {
    const wrapper = mount(FeedItems, {
      global: {
        stubs: {
          TopBar: { template: '<div><slot name="left-actions" /></div>' },
          ItemList: true,
          LabelPicker: true,
        },
      },
    });
    await nextTick();

    const button = wrapper.find('button[aria-label="Edit labels"]');
    expect(button.exists()).toBe(true);
    expect(button.find('use[href="#icon-tag"]').exists()).toBe(true);
  });
});
