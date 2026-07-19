import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";

const mockLoadFeeds = vi.fn();
const mockLoadGroupedFeeds = vi.fn();
const mockRemoveFeed = vi.fn();
const mockRefreshFeed = vi.fn();
vi.mock("@/stores/feeds", () => ({
  useFeedsStore: () => ({
    feeds: [],
    totalUnread: 0,
    visibleFeeds: [],
    filterUnreadOnly: false,
    loadFeeds: mockLoadFeeds,
    loadGroupedFeeds: mockLoadGroupedFeeds,
    loadFeedsLimit: vi.fn(),
    labelGroups: [
      {
        label: { id: 1, user_id: 1, name: "Blogs", created_at: "" },
        feeds: [
          {
            id: 1,
            url: "https://example.com/rss",
            title: "Example Blog",
            created_at: "",
            updated_at: "",
            unread_count: 3,
          },
        ],
      },
      {
        label: { id: 2, user_id: 1, name: "News", created_at: "" },
        feeds: [
          {
            id: 2,
            url: "https://news.example.com/rss",
            title: "News Site",
            created_at: "",
            updated_at: "",
            unread_count: 0,
          },
        ],
      },
    ],
    unlabeledFeeds: [
      {
        id: 3,
        url: "https://other.example.com/rss",
        title: "Other Feed",
        created_at: "",
        updated_at: "",
        unread_count: 0,
      },
    ],
    collapsedLabelIds: new Set(),
    visibleLabelGroups: [
      {
        label: { id: 1, user_id: 1, name: "Blogs", created_at: "" },
        feeds: [
          {
            id: 1,
            url: "https://example.com/rss",
            title: "Example Blog",
            created_at: "",
            updated_at: "",
            unread_count: 3,
          },
        ],
      },
      {
        label: { id: 2, user_id: 1, name: "News", created_at: "" },
        feeds: [
          {
            id: 2,
            url: "https://news.example.com/rss",
            title: "News Site",
            created_at: "",
            updated_at: "",
            unread_count: 0,
          },
        ],
      },
    ],
    visibleUnlabeledFeeds: [
      {
        id: 3,
        url: "https://other.example.com/rss",
        title: "Other Feed",
        created_at: "",
        updated_at: "",
        unread_count: 0,
      },
    ],
    toggleCollapseLabel: vi.fn(),
    startUnreadPolling: vi.fn(),
    stopUnreadPolling: vi.fn(),
    feedNames: {},
    removeFeed: mockRemoveFeed,
    refreshFeed: mockRefreshFeed,
  }),
}));

let mockIsAdmin = false;
let mockUsername = "user";
const mockLogout = vi.fn();
vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({
    user: { id: 1, username: mockUsername, is_admin: mockIsAdmin },
    isAuthenticated: true,
    isAdmin: mockIsAdmin,
    logout: mockLogout,
  }),
}));

vi.mock("@/composables/useAddFeed", () => ({
  useAddFeed: () => ({
    open: vi.fn(),
  }),
}));

vi.mock("@/composables/useTheme", () => ({
  useTheme: () => ({
    isDark: false,
  }),
}));

import FeedList from "@/components/FeedList.vue";

import { createRouter, createWebHistory } from "vue-router";
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/unread" },
    { path: "/unread", name: "unread", component: { template: "<div>unread</div>" } },
    { path: "/starred", name: "starred", component: { template: "<div>starred</div>" } },
    { path: "/admin", name: "admin", component: { template: "<div>admin</div>" } },
    { path: "/settings", name: "settings", component: { template: "<div>settings</div>" } },
  ],
});

describe("FeedList admin link", () => {
  beforeEach(async () => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    mockIsAdmin = false;
    mockUsername = "user";
    router.push("/unread");
    await router.isReady();
  });

  it("shows Administration link when user is admin", async () => {
    mockIsAdmin = true;
    mockUsername = "admin";

    const wrapper = mount(FeedList, {
      global: {
        plugins: [router],
        stubs: {
          ThemeToggle: true,
        },
      },
    });
    await router.isReady();

    const links = wrapper.findAll("a");
    const adminLink = links.find((l) => l.text().trim() === "Administration");
    expect(adminLink).toBeTruthy();
  });

  it("hides Administration link when user is not admin", async () => {
    mockIsAdmin = false;
    mockUsername = "user";

    const wrapper = mount(FeedList, {
      global: {
        plugins: [router],
        stubs: {
          ThemeToggle: true,
        },
      },
    });
    await router.isReady();

    const links = wrapper.findAll("a");
    const adminLink = links.find((l) => l.text().trim() === "Administration");
    expect(adminLink).toBeUndefined();
  });

  it("admin link has correct href", async () => {
    mockIsAdmin = true;
    mockUsername = "admin";

    const wrapper = mount(FeedList, {
      global: {
        plugins: [router],
        stubs: {
          ThemeToggle: true,
        },
      },
    });
    await router.isReady();

    const links = wrapper.findAll("a");
    const adminLink = links.find((l) => l.text().trim() === "Administration");
    expect(adminLink).toBeTruthy();
    expect(adminLink!.attributes("href")).toBe("/admin");
  });
});

describe("FeedList context menu", () => {
  beforeEach(async () => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    mockIsAdmin = false;
    mockUsername = "user";
    router.push("/unread");
    await router.isReady();
  });

  it("shows FeedContextMenu on right-click on a feed", async () => {
    const wrapper = mount(FeedList, {
      global: {
        plugins: [router],
        stubs: {
          ThemeToggle: true,
          FeedContextMenu: true,
        },
      },
    });
    await router.isReady();

    const feedButton = wrapper.find('[data-feed-id="1"]');
    expect(feedButton.exists()).toBe(true);

    await feedButton.trigger("contextmenu", {
      clientX: 100,
      clientY: 200,
      preventDefault: vi.fn(),
    });

    const contextMenu = wrapper.findComponent({ name: "FeedContextMenu" });
    expect(contextMenu.exists()).toBe(true);
    expect(contextMenu.props("feedId")).toBe(1);
  });
});
