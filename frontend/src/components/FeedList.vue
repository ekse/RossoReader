<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useFeedsStore } from "@/stores/feeds";
import { useAuthStore } from "@/stores/auth";
import ThemeToggle from "./ThemeToggle.vue";
import { useAddFeed } from "@/composables/useAddFeed";
import { useSearch } from "@/composables/useSearch";
import { useTheme } from "@/composables/useTheme";
import whiteLogo from "@/assets/rosso_reader_white_112px.png";
import transparentLogo from "@/assets/rosso_reader_transparent_112px.png";

const feedsStore = useFeedsStore();
const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

onMounted(async () => {
  await feedsStore.loadGroupedFeeds();
  feedsStore.loadFeedsLimit();
  feedsStore.startUnreadPolling();
});

onBeforeUnmount(() => {
  feedsStore.stopUnreadPolling();
});

function selectFeed(id: number) {
  router.push(`/feed/${id}`);
}

async function logout() {
  await auth.logout();
  router.push("/login");
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "";
  return new Date(dateStr).toLocaleString();
}

const lastUpdate = computed(() => {
  const allFeeds = [
    ...feedsStore.unlabeledFeeds,
    ...feedsStore.labelGroups.flatMap((g) => g.feeds),
  ];
  const dates = allFeeds.map((f) => f.last_fetched_at).filter(Boolean) as string[];
  if (dates.length === 0) return null;
  dates.sort();
  return dates[dates.length - 1];
});

const totalUnread = computed(() => feedsStore.totalUnread);

const { open: openAddFeed } = useAddFeed();
const { openSearch } = useSearch();
const { isDark } = useTheme();
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 p-4 overflow-y-auto">
      <div class="px-3 mb-3 flex items-start justify-between">
        <span class="flex items-center select-none">
          <img :src="isDark ? transparentLogo : whiteLogo" alt="Rosso" class="w-[112px] h-auto" />
        </span>
        <ThemeToggle />
      </div>
      <router-link
        to="/unread"
        class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium"
        :class="
          route.path === '/unread'
            ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
            : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'
        "
      >
        <svg class="w-4 h-4 shrink-0">
          <use href="#icon-inbox" />
        </svg>
        <span class="flex-1">New</span>
        <span
          v-if="totalUnread > 0"
          class="inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
        >
          {{ totalUnread }}
        </span>
      </router-link>
      <router-link
        to="/starred"
        class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium mt-1"
        :class="
          route.path === '/starred'
            ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
            : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'
        "
      >
        <svg class="w-4 h-4 shrink-0">
          <use href="#icon-star" />
        </svg>
        Starred
      </router-link>

      <button
        @click="openSearch"
        class="w-full text-left px-3 py-1.5 rounded-md text-sm font-medium mt-1 text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50 flex items-center gap-2"
      >
        <svg class="w-4 h-4">
          <use href="#icon-search" />
        </svg>
        Search
      </button>

      <router-link
        to="/settings"
        class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium mt-1"
        :class="
          route.path === '/settings'
            ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
            : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'
        "
      >
        <svg class="w-4 h-4 shrink-0">
          <use href="#icon-gear" />
        </svg>
        Settings
      </router-link>

      <router-link
        v-if="auth.isAdmin"
        to="/admin"
        class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium mt-1"
        :class="
          route.path === '/admin'
            ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
            : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'
        "
      >
        <svg class="w-4 h-4 shrink-0">
          <use href="#icon-shield" />
        </svg>
        Administration
      </router-link>

      <div class="mt-3">
        <div class="px-3 flex items-center justify-between">
          <div class="flex items-center gap-1">
            <h3
              class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider"
            >
              Feeds
            </h3>
            <button
              @click="openAddFeed"
              class="w-5 h-5 flex items-center font-bold justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-700 transition-colors"
            >
              +
            </button>
          </div>
          <button
            @click="feedsStore.filterUnreadOnly = !feedsStore.filterUnreadOnly"
            :title="
              feedsStore.filterUnreadOnly ? 'Show all feeds' : 'Show only feeds with unread items'
            "
            class="w-5 h-5 flex items-center justify-center rounded transition-colors"
            :class="
              feedsStore.filterUnreadOnly
                ? 'text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-900/50'
                : 'text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-700'
            "
          >
            <svg class="w-3.5 h-3.5">
              <use href="#icon-filter" />
            </svg>
          </button>
        </div>
        <div class="mt-2 space-y-1">
          <!-- Label groups -->
          <div v-for="group in feedsStore.visibleLabelGroups" :key="group.label.id">
            <button
              @click="feedsStore.toggleCollapseLabel(group.label.id)"
              class="w-full text-left px-3 py-1.5 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50 flex items-center gap-1.5"
            >
              <svg
                class="w-3 h-3 shrink-0 transition-transform"
                :class="feedsStore.collapsedLabelIds.has(group.label.id) ? '' : 'rotate-90'"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path d="M6 4l8 6-8 6V4z" />
              </svg>
              <span class="truncate">{{ group.label.name }}</span>
              <span
                v-if="group.feeds.reduce((s, f) => s + (f.unread_count || 0), 0) > 0"
                class="ml-auto shrink-0 inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
              >
                {{ group.feeds.reduce((s, f) => s + (f.unread_count || 0), 0) }}
              </span>
            </button>
            <div v-if="!feedsStore.collapsedLabelIds.has(group.label.id)" class="ml-3 space-y-1">
              <button
                v-for="feed in group.feeds"
                :key="feed.id"
                @click="selectFeed(feed.id)"
                class="w-full text-left px-3 py-2 rounded-md text-sm"
                :class="
                  route.params.id === String(feed.id)
                    ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
                    : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'
                "
              >
                <div class="flex items-center gap-2">
                  <img
                    v-if="feed.icon_url"
                    :src="feed.icon_url"
                    class="w-4 h-4 shrink-0 rounded"
                    alt=""
                    loading="lazy"
                  />
                  <div class="flex items-center justify-between flex-1 min-w-0">
                    <span class="truncate">{{ feed.title || feed.url }}</span>
                    <span
                      v-if="feed.unread_count && feed.unread_count > 0"
                      class="ml-2 shrink-0 inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                    >
                      {{ feed.unread_count }}
                    </span>
                  </div>
                </div>
              </button>
            </div>
          </div>

          <!-- Unlabeled feeds -->
          <button
            v-for="feed in feedsStore.visibleUnlabeledFeeds"
            :key="'ul-' + feed.id"
            @click="selectFeed(feed.id)"
            class="w-full text-left px-3 py-2 rounded-md text-sm"
            :class="
              route.params.id === String(feed.id)
                ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300'
                : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700/50'
            "
          >
            <div class="flex items-center gap-2">
              <img
                v-if="feed.icon_url"
                :src="feed.icon_url"
                class="w-4 h-4 shrink-0 rounded"
                alt=""
                loading="lazy"
              />
              <div class="flex items-center justify-between flex-1 min-w-0">
                <span class="truncate">{{ feed.title || feed.url }}</span>
                <span
                  v-if="feed.unread_count && feed.unread_count > 0"
                  class="ml-2 shrink-0 inline-flex items-center justify-center px-2 py-0.5 text-xs font-medium rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                >
                  {{ feed.unread_count }}
                </span>
              </div>
            </div>
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="lastUpdate"
      class="shrink-0 px-4 py-3 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-400 dark:text-gray-500"
    >
      Last updated: {{ formatDate(lastUpdate) }}
    </div>

    <div
      v-if="auth.user"
      class="shrink-0 px-4 py-3 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between"
    >
      <router-link
        to="/settings"
        class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-gray-100"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
          />
        </svg>
        <span>{{ auth.user.username }}</span>
        <span
          v-if="auth.isAdmin"
          class="px-1.5 py-0.5 text-[10px] font-medium rounded bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
          >admin</span
        >
      </router-link>
      <button
        @click="logout"
        title="Sign out"
        class="text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
          />
        </svg>
      </button>
    </div>
  </div>
</template>
