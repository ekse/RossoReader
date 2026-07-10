import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";

const mockLoadFeeds = vi.fn();
const mockLoadGroupedFeeds = vi.fn();
vi.mock("@/stores/feeds", () => ({
  useFeedsStore: () => ({
    feeds: [],
    totalUnread: 0,
    visibleFeeds: [],
    filterUnreadOnly: false,
    loadFeeds: mockLoadFeeds,
    loadGroupedFeeds: mockLoadGroupedFeeds,
    loadFeedsLimit: vi.fn(),
    labelGroups: [],
    unlabeledFeeds: [],
    collapsedLabelIds: new Set(),
    visibleLabelGroups: [],
    visibleUnlabeledFeeds: [],
    toggleCollapseLabel: vi.fn(),
    startUnreadPolling: vi.fn(),
    stopUnreadPolling: vi.fn(),
    feedNames: {},
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
