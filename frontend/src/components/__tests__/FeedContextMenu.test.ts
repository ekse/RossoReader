import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";

const { mockRemoveFeed, mockLoadGroupedFeeds, mockLoadLabels, mockRenameFeed } = vi.hoisted(() => ({
  mockRemoveFeed: vi.fn(),
  mockLoadGroupedFeeds: vi.fn(),
  mockLoadLabels: vi.fn(),
  mockRenameFeed: vi.fn(),
}));

vi.mock("@/stores/feeds", () => ({
  useFeedsStore: () => ({
    feeds: [{ id: 1, title: "Test Feed", url: "https://example.com/rss" }],
    removeFeed: mockRemoveFeed,
    loadGroupedFeeds: mockLoadGroupedFeeds,
    loadLabels: mockLoadLabels,
    renameFeed: mockRenameFeed,
  }),
}));

const {
  mockMarkFeedRead,
  mockFetchLabels,
  mockFetchFeedLabels,
  mockAddFeedLabel,
  mockRemoveFeedLabel,
  mockCreateLabel,
} = vi.hoisted(() => ({
  mockMarkFeedRead: vi.fn(),
  mockFetchLabels: vi.fn(),
  mockFetchFeedLabels: vi.fn(),
  mockAddFeedLabel: vi.fn(),
  mockRemoveFeedLabel: vi.fn(),
  mockCreateLabel: vi.fn(),
}));

vi.mock("@/api/client", () => ({
  markFeedRead: mockMarkFeedRead,
  fetchLabels: mockFetchLabels,
  fetchFeedLabels: mockFetchFeedLabels,
  addFeedLabel: mockAddFeedLabel,
  removeFeedLabel: mockRemoveFeedLabel,
  createLabel: mockCreateLabel,
}));

vi.mock("vue-router", () => ({
  useRouter: () => ({
    push: vi.fn(),
    currentRoute: { value: { params: {} } },
  }),
}));

import FeedContextMenu from "@/components/FeedContextMenu.vue";

function bodyText() {
  return document.body.innerText || document.body.textContent || "";
}

function findButtonInBody(text: string) {
  const buttons = document.body.querySelectorAll("button");
  return Array.from(buttons).find((b) => b.textContent?.trim() === text);
}

describe("FeedContextMenu", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  function mountMenu() {
    return mount(FeedContextMenu, {
      props: { feedId: 1, x: 100, y: 200 },
      attachTo: document.body,
    });
  }

  it("renders main menu actions", () => {
    mountMenu();
    expect(bodyText()).toContain("Mark all as read");
    expect(bodyText()).toContain("Edit labels");
    expect(bodyText()).toContain("Unsubscribe");
  });

  it("calls markFeedRead and reloads on Mark all as read click", async () => {
    mountMenu();
    const btn = findButtonInBody("Mark all as read");
    expect(btn).toBeTruthy();
    btn!.click();

    await flushPromises();
    expect(mockMarkFeedRead).toHaveBeenCalledWith(1);
    expect(mockLoadGroupedFeeds).toHaveBeenCalled();
  });

  it("calls removeFeed and closes on Unsubscribe click", async () => {
    mountMenu();
    const btn = findButtonInBody("Unsubscribe");
    expect(btn).toBeTruthy();
    btn!.click();

    expect(mockRemoveFeed).toHaveBeenCalledWith(1);
  });

  it("shows label picker on Edit labels click", async () => {
    mockFetchLabels.mockResolvedValue([{ id: 1, user_id: 1, name: "News", created_at: "" }]);
    mockFetchFeedLabels.mockResolvedValue([]);

    mountMenu();
    const btn = findButtonInBody("Edit labels");
    expect(btn).toBeTruthy();
    btn!.click();

    await flushPromises();
    expect(bodyText()).toContain("Labels");
  });

  it("closes on Escape key", () => {
    const wrapper = mountMenu();
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(wrapper.emitted("close")).toBeTruthy();
  });

  it("closes on click outside", () => {
    const wrapper = mountMenu();
    document.dispatchEvent(new MouseEvent("mousedown", { bubbles: true }));
    expect(wrapper.emitted("close")).toBeTruthy();
  });

  it("shows rename input when clicking Rename", async () => {
    mountMenu();
    const btn = findButtonInBody("Rename");
    expect(btn).toBeTruthy();
    btn!.click();

    await flushPromises();
    expect(bodyText()).toContain("Rename feed");
  });

  it("calls renameFeed on save", async () => {
    mountMenu();
    const btn = findButtonInBody("Rename");
    btn!.click();
    await flushPromises();

    const input = document.body.querySelector("input") as HTMLInputElement;
    input.value = "New Name";
    input.dispatchEvent(new Event("input"));

    const saveBtn = findButtonInBody("Save");
    saveBtn!.click();
    await flushPromises();

    expect(mockRenameFeed).toHaveBeenCalledWith(1, "New Name");
  });
});
