import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";

vi.mock("@/api/client", () => ({
  fetchItems: vi.fn(),
  updateItem: vi.fn(),
}));

import { useItemsStore } from "../items";
import type { Item } from "@/types";

const item: Item = {
  id: 1,
  feed_id: 1,
  guid: "1",
  title: "Test Post",
  url: "https://ex.com/1",
  fetched_at: "2024-01-01T00:00:00Z",
  read: true,
  starred: false,
};

describe("useItemsStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("clears stale items immediately when the read filter changes", async () => {
    const api = await import("@/api/client");
    vi.mocked(api.fetchItems).mockResolvedValue({ items: [], total: 0, page: 1 });

    const store = useItemsStore();
    store.items = [item];
    store.total = 1;

    store.setFilterRead(false);

    expect(store.items).toEqual([]);
    expect(store.total).toBe(0);
    expect(store.page).toBe(1);

    await Promise.resolve();
    expect(api.fetchItems).toHaveBeenCalledWith(expect.objectContaining({ read: false, page: 1 }));
  });
});
