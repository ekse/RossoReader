import { ref, watch } from "vue";
import { useRoute } from "vue-router";

const _highlightQuery = ref("");

const HIGHLIGHT_CLASS = "bg-yellow-200 dark:bg-yellow-700/70 rounded px-0.5";

function highlightText(text: string): string {
  const q = _highlightQuery.value;
  if (!text || !q) return text || "";
  const div = document.createElement("div");
  div.innerHTML = text;
  const clean = div.textContent || div.innerText || "";
  if (!clean) return text;
  const escapedQ = q.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const regex = new RegExp(`(${escapedQ})`, "gi");
  return clean.replace(regex, `<mark class="${HIGHLIGHT_CLASS}">$1</mark>`);
}

export function useSearchHighlight() {
  const route = useRoute();

  if (!route) {
    return { highlightQuery: _highlightQuery, highlightText };
  }

  function onRouteChange() {
    const hash = route.hash;
    const match = hash.match(/^#highlight=(.+)$/);
    _highlightQuery.value = match ? decodeURIComponent(match[1]) : "";
  }

  onRouteChange();

  watch(() => route.hash, onRouteChange);

  return { highlightQuery: _highlightQuery, highlightText };
}

export function highlightTextNodes(root: HTMLElement, query: string) {
  if (!query) return;

  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
    acceptNode(node) {
      const tag = node.parentElement?.tagName;
      if (tag === "MARK" || tag === "SCRIPT" || tag === "STYLE") {
        return NodeFilter.FILTER_REJECT;
      }
      return NodeFilter.FILTER_ACCEPT;
    },
  });

  const textNodes: Text[] = [];
  while (walker.nextNode()) {
    textNodes.push(walker.currentNode as Text);
  }

  const escapedQ = query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const regex = new RegExp(`(${escapedQ})`, "gi");

  for (const textNode of textNodes) {
    const text = textNode.nodeValue || "";
    if (!text.match(regex)) continue;

    const fragment = document.createDocumentFragment();
    let lastIndex = 0;
    let m: RegExpExecArray | null;
    const r = new RegExp(`(${escapedQ})`, "gi");

    while ((m = r.exec(text)) !== null) {
      if (m.index > lastIndex) {
        fragment.appendChild(document.createTextNode(text.slice(lastIndex, m.index)));
      }
      const mark = document.createElement("mark");
      mark.className = HIGHLIGHT_CLASS;
      mark.textContent = m[1];
      fragment.appendChild(mark);
      lastIndex = r.lastIndex;
    }

    if (lastIndex < text.length) {
      fragment.appendChild(document.createTextNode(text.slice(lastIndex)));
    }

    textNode.parentNode?.replaceChild(fragment, textNode);
  }
}
