import { describe, it, expect, beforeEach, vi } from "vitest";
import { ITEM_GROUPING_FEED, ITEM_GROUPING_NONE } from "@/types";

describe("useItemGrouping", () => {
  beforeEach(() => {
    vi.resetModules();
    localStorage.clear();
  });

  it("defaults to none when nothing is stored", async () => {
    const { useItemGrouping } = await import("../useItemGrouping");
    const { grouping } = useItemGrouping();
    expect(grouping.value).toBe(ITEM_GROUPING_NONE);
  });

  it("reads a stored feed grouping on init", async () => {
    localStorage.setItem("itemGrouping", ITEM_GROUPING_FEED);
    const { useItemGrouping } = await import("../useItemGrouping");
    const { grouping } = useItemGrouping();
    expect(grouping.value).toBe(ITEM_GROUPING_FEED);
  });

  it("persists the selection made via setGrouping", async () => {
    const { useItemGrouping } = await import("../useItemGrouping");
    const { grouping, setGrouping } = useItemGrouping();
    setGrouping(ITEM_GROUPING_FEED);
    expect(grouping.value).toBe(ITEM_GROUPING_FEED);
    expect(localStorage.getItem("itemGrouping")).toBe(ITEM_GROUPING_FEED);
  });

  it("persists the selection made via the writable computed", async () => {
    const { useItemGrouping } = await import("../useItemGrouping");
    const { grouping } = useItemGrouping();
    grouping.value = ITEM_GROUPING_NONE;
    expect(grouping.value).toBe(ITEM_GROUPING_NONE);
    expect(localStorage.getItem("itemGrouping")).toBe(ITEM_GROUPING_NONE);
  });
});
