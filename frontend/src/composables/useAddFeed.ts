import { ref } from "vue";

const showAddFeed = ref(false);

export function useAddFeed() {
  function open() {
    showAddFeed.value = true;
  }
  function close() {
    showAddFeed.value = false;
  }
  return { showAddFeed, open, close };
}
