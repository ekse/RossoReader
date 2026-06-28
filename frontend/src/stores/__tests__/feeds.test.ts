import { describe, it, expect, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useFeedsStore } from "../feeds";

describe("useFeedsStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("starts with an empty feed list", () => {
    const store = useFeedsStore();
    expect(store.feeds).toEqual([]);
    expect(store.loading).toBe(false);
  });
});
