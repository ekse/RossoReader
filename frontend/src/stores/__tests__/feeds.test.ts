import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useFeedsStore } from "../feeds";
import type { Feed } from "@/types";

vi.mock("@/api/client", () => ({
  addFeed: vi.fn(),
  fetchFeeds: vi.fn(),
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
});
