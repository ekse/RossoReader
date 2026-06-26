# Scroll-to-Hide Mobile Header & Sticky Desktop Header

## Overview

Improve header accessibility and screen utilization by introducing sticky actions on desktop and dynamic scroll-to-hide / scroll-up-to-reveal navigation (the "Quick Return" pattern) on mobile viewports. The implementation is lightweight, leveraging reactive global Vue state.

## Decisions (confirmed)

- **Desktop Header**: Sticky at the top (`sticky top-0 z-10`) with a glassmorphism blur effect (`bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm`).
- **Mobile Header**: Hides on scroll-down (after scrolling past `50px` threshold) and immediately reveals on scroll-up.
- **Scope**: Applied to the main feed lists (`FeedItems.vue`, `UnreadItems.vue`, and `StarredItems.vue`).
- **Reset Behavior**: Reset header visibility to `true` on every route transition so the user starts with visible headers when changing views.
- **Testing**: Dedicated unit tests for the composable, validating state sharing and lifecycle resets.

## Phase 1 — State & Composable

Create the state container in [useHeader.ts](file:///Users/pmiron/repos/RossoReader/frontend/src/composables/useHeader.ts) to manage visibility globally:
- `isHeaderVisible`: A boolean ref default-initialized to `true`.
- `lastScrollTop`: A number ref default-initialized to `0`.

## Phase 2 — Main Container

Modify [App.vue](file:///Users/pmiron/repos/RossoReader/frontend/src/App.vue) to bind scroll listener:
- Bind `@scroll="handleScroll"` to the `<main>` scroll container.
- Implement `handleScroll()` to:
  1. Return early and force `isHeaderVisible.value = true` on desktop viewports (`window.innerWidth >= 768`).
  2. Ignore negative scrolls (to handle iOS bounce scrolling).
  3. Toggle `isHeaderVisible.value = false` when scrolling down past a `50px` threshold.
  4. Toggle `isHeaderVisible.value = true` when scrolling up.
- Watch `route.path` to reset `isHeaderVisible.value = true` on navigation.

## Phase 3 — Views & Layouts

Apply the sticky and transition styles in the view files:
- **Modified files**: [FeedItems.vue](file:///Users/pmiron/repos/RossoReader/frontend/src/views/FeedItems.vue), [UnreadItems.vue](file:///Users/pmiron/repos/RossoReader/frontend/src/views/UnreadItems.vue), and [StarredItems.vue](file:///Users/pmiron/repos/RossoReader/frontend/src/views/StarredItems.vue).
- Bind transition classes to the header tag:
  - Transition duration: `transition-transform duration-300 ease-in-out`
  - Dynamic translate classes: `:class="isHeaderVisible ? 'translate-y-0' : '-translate-y-full'"`
  - Responsive override: `md:translate-y-0` (ensures desktop views remain fixed).

## Phase 4 — Testing & Validation

### Automated Tests
- Unit tests written in [useHeader.test.ts](file:///Users/pmiron/repos/RossoReader/frontend/src/composables/__tests__/useHeader.test.ts) to test:
  1. Default states.
  2. State changes.
  3. Shared global state instances.
- Run tests inside the frontend container:
  ```bash
  docker compose exec web-dev pnpm test
  ```

### Manual Verification
- **Desktop**: Scroll down and verify the glassmorphism header stays sticky at `top-0`.
- **Mobile**: Toggle mobile emulation, scroll down to hide header, scroll up to reveal. Navigate between feeds and verify the header resets to visible.
