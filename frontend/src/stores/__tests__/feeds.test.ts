import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useFeedsStore } from "../feeds";
import type { Feed } from "@/types";

vi.mock("@/api/client", () => ({
  addFeed: vi.fn(),
  fetchFeeds: vi.fn(),
  fetchUnreadCounts: vi.fn(),
}));

describe("useFeedsStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("starts with an empty feed list", () => {
    const store = useFeedsStore();
    expect(store.feeds).toEqual([]);
    expect(store.loading).toBe(false);
  });

  describe("importFeeds", () => {
    it("imports all feeds successfully", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.addFeed).mockResolvedValue({ id: 1 } as Feed);
      vi.mocked(api.fetchFeeds).mockResolvedValue([]);

      const store = useFeedsStore();
      const result = await store.importFeeds(["https://a.com/rss", "https://b.com/rss"]);

      expect(result).toEqual({ imported: 2, skipped: 0 });
    });

    it("skips duplicate feeds that return 409", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.addFeed)
        .mockResolvedValueOnce({ id: 1 } as Feed)
        .mockRejectedValueOnce({ response: { status: 409 } })
        .mockResolvedValueOnce({ id: 3 } as Feed);
      vi.mocked(api.fetchFeeds).mockResolvedValue([]);

      const store = useFeedsStore();
      const result = await store.importFeeds([
        "https://a.com/rss",
        "https://dup.com/rss",
        "https://c.com/rss",
      ]);

      expect(result).toEqual({ imported: 2, skipped: 1 });
    });

    it("continues on other errors", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.addFeed)
        .mockResolvedValueOnce({ id: 1 } as Feed)
        .mockRejectedValueOnce(new Error("network error"))
        .mockResolvedValueOnce({ id: 3 } as Feed);
      vi.mocked(api.fetchFeeds).mockResolvedValue([]);

      const store = useFeedsStore();
      const result = await store.importFeeds([
        "https://a.com/rss",
        "https://broken.com/rss",
        "https://c.com/rss",
      ]);

      expect(result).toEqual({ imported: 2, skipped: 0 });
    });

    it("refreshes the feed list after importing", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.addFeed).mockResolvedValue({ id: 1 } as Feed);
      vi.mocked(api.fetchFeeds).mockResolvedValue([
        { id: 1, url: "https://a.com/rss", title: "A" },
      ] as Feed[]);

      const store = useFeedsStore();
      expect(store.feeds).toHaveLength(0);

      await store.importFeeds(["https://a.com/rss"]);

      expect(store.feeds).toHaveLength(1);
      expect(store.feeds[0].url).toBe("https://a.com/rss");
    });
  });

  describe("unread polling", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("updates unread counts on feeds", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.fetchUnreadCounts).mockResolvedValue({ 1: 5, 2: 3 });

      const store = useFeedsStore();
      store.feeds = [
        { id: 1, url: "https://a.com/rss" } as Feed,
        { id: 2, url: "https://b.com/rss" } as Feed,
        { id: 3, url: "https://c.com/rss" } as Feed,
      ];

      store.startUnreadPolling();
      await vi.advanceTimersByTimeAsync(10 * 60 * 1000);

      expect(store.feeds[0].unread_count).toBe(5);
      expect(store.feeds[1].unread_count).toBe(3);
      expect(store.feeds[2].unread_count).toBe(0);
    });

    it("stops polling after stopUnreadPolling", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.fetchUnreadCounts).mockResolvedValue({ 1: 5 });

      const store = useFeedsStore();
      store.feeds = [{ id: 1, url: "https://a.com/rss" } as Feed];

      store.startUnreadPolling();
      store.stopUnreadPolling();
      await vi.advanceTimersByTimeAsync(10 * 60 * 1000);

      expect(api.fetchUnreadCounts).toHaveBeenCalledTimes(0);
    });

    it("handles empty feeds list without error", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.fetchUnreadCounts).mockResolvedValue({});

      const store = useFeedsStore();
      store.feeds = [];

      store.startUnreadPolling();
      await vi.advanceTimersByTimeAsync(10 * 60 * 1000);

      expect(api.fetchUnreadCounts).toHaveBeenCalledTimes(1);
    });

    it("silently handles network errors", async () => {
      const api = await import("@/api/client");
      vi.mocked(api.fetchUnreadCounts).mockRejectedValue(new Error("network error"));

      const store = useFeedsStore();
      store.feeds = [{ id: 1, url: "https://a.com/rss" } as Feed];

      store.startUnreadPolling();
      await vi.advanceTimersByTimeAsync(10 * 60 * 1000);

      expect(api.fetchUnreadCounts).toHaveBeenCalledTimes(1);
      expect(store.feeds[0].unread_count).toBeUndefined();
    });
  });
});
