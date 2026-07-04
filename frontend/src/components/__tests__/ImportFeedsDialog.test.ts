import { describe, it, expect, beforeEach, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { setActivePinia, createPinia } from "pinia";
import ImportFeedsDialog from "../ImportFeedsDialog.vue";
import { useFeedsStore } from "@/stores/feeds";
import type { DiscoveredFeed } from "@/types";

const mockPreviewOpmlImport = vi.fn();
vi.mock("@/api/client", () => ({
  previewOpmlImport: (...args: any[]) => mockPreviewOpmlImport(...args),
}));

function createMockFile(content: string, name = "feeds.opml"): File {
  return new File([content], name, { type: "text/xml" });
}

async function selectFile(wrapper: any, feeds: DiscoveredFeed[]) {
  mockPreviewOpmlImport.mockResolvedValue(feeds);
  const input = wrapper.find('input[type="file"]');
  Object.defineProperty(input.element, "files", {
    value: [createMockFile("<opml></opml>")],
  });
  await input.trigger("change");
  await new Promise((resolve) => setTimeout(resolve, 0));
}

describe("ImportFeedsDialog", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it("renders file input initially", () => {
    const wrapper = mount(ImportFeedsDialog);
    expect(wrapper.text()).toContain("Import Feeds");
    expect(wrapper.find('input[type="file"]').exists()).toBe(true);
  });

  it("emits close when cancel is clicked", async () => {
    const wrapper = mount(ImportFeedsDialog);
    await wrapper
      .findAll("button")
      .find((b: any) => b.text() === "Cancel")!
      .trigger("click");
    expect(wrapper.emitted("close")).toBeTruthy();
  });

  it("shows preview list with checkboxes after file selection", async () => {
    const wrapper = mount(ImportFeedsDialog);
    await selectFile(wrapper, [
      { title: "Feed A", url: "https://a.com/rss" },
      { title: "Feed B", url: "https://b.com/rss" },
    ]);

    expect(wrapper.text()).toContain("Feed A");
    expect(wrapper.text()).toContain("Feed B");
    expect(wrapper.text()).toContain("Import Selected (2)");
  });

  it("checkboxes toggle selection count", async () => {
    const wrapper = mount(ImportFeedsDialog);
    await selectFile(wrapper, [
      { title: "Feed A", url: "https://a.com/rss" },
      { title: "Feed B", url: "https://b.com/rss" },
    ]);

    const checkboxes = wrapper.findAll('input[type="checkbox"]');
    expect(checkboxes).toHaveLength(2);

    await checkboxes[0].setValue(false);
    expect(wrapper.text()).toContain("Import Selected (1)");

    await checkboxes[1].setValue(false);
    expect(wrapper.text()).toContain("Import Selected (0)");
  });

  it("import button disabled when none selected", async () => {
    const wrapper = mount(ImportFeedsDialog);
    await selectFile(wrapper, [{ title: "Feed A", url: "https://a.com/rss" }]);

    await wrapper.find('input[type="checkbox"]').setValue(false);

    const btn = wrapper.findAll("button").find((b: any) => b.text().startsWith("Import Selected"));
    expect(btn?.attributes("disabled")).toBeDefined();
  });

  it("shows success message after import", async () => {
    const store = useFeedsStore();
    vi.spyOn(store, "importFeeds").mockResolvedValue({ imported: 1, skipped: 1 });

    const wrapper = mount(ImportFeedsDialog);
    await selectFile(wrapper, [
      { title: "Feed A", url: "https://a.com/rss" },
      { title: "Feed B", url: "https://b.com/rss" },
    ]);

    const importBtn = wrapper
      .findAll("button")
      .find((b: any) => b.text().startsWith("Import Selected"));
    await importBtn!.trigger("click");
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(wrapper.text()).toContain("1 feed(s) imported");
    expect(wrapper.text()).toContain("1 skipped");
  });

  it("back button resets to upload state", async () => {
    const wrapper = mount(ImportFeedsDialog);
    await selectFile(wrapper, [{ title: "Feed A", url: "https://a.com/rss" }]);

    const backBtn = wrapper.findAll("button").find((b: any) => b.text() === "Back");
    await backBtn!.trigger("click");

    expect(wrapper.find('input[type="file"]').exists()).toBe(true);
  });

  it("done button closes dialog after import", async () => {
    const store = useFeedsStore();
    vi.spyOn(store, "importFeeds").mockResolvedValue({ imported: 1, skipped: 0 });

    const wrapper = mount(ImportFeedsDialog);
    await selectFile(wrapper, [{ title: "Feed A", url: "https://a.com/rss" }]);

    const importBtn = wrapper
      .findAll("button")
      .find((b: any) => b.text().startsWith("Import Selected"));
    await importBtn!.trigger("click");
    await new Promise((resolve) => setTimeout(resolve, 0));

    const doneBtn = wrapper.findAll("button").find((b: any) => b.text() === "Done");
    await doneBtn!.trigger("click");

    expect(wrapper.emitted("close")).toBeTruthy();
  });
});
