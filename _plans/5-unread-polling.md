# Poll unread counts in sidebar

Periodically refresh `unread_count` on every feed in the sidebar so the user sees when new items arrive without any action.

## Backend

**New endpoint** `GET /api/feeds/unread-counts` in `backend/internal/handlers/feeds.go`:
- Calls `h.Store.GetUnreadCountByFeed(r.Context(), userID)` (already exists)
- Returns `{"counts": {"1": 5, "2": 0}}` (map[int64]int)
- Route registered in `handlers.go` as `r.Get("/api/feeds/unread-counts", h.UnreadCounts)`

**Tests** in `feeds_test.go`:
- Returns correct counts for feeds with unread items
- All 0 when no items exist
- Only returns counts for authenticated user

## Frontend

1. **Constant** in `frontend/src/types.ts`:
   `UNREAD_POLL_INTERVAL_MS = 10 * 60 * 1000`

2. **API client** in `frontend/src/api/client.ts`:
   `fetchUnreadCounts()` → `GET /api/feeds/unread-counts` → `Record<number, number>`

3. **Feeds store** in `frontend/src/stores/feeds.ts`:
   - `startUnreadPolling()`: `setInterval` every `UNREAD_POLL_INTERVAL_MS`, fetches counts, updates `unread_count` on each feed in-place (same object references propagate to label groups)
   - `stopUnreadPolling()`: `clearInterval`
   - Silent failure on network errors

4. **FeedList** in `frontend/src/components/FeedList.vue`:
   - `onMounted`: `feedsStore.startUnreadPolling()`
   - `onBeforeUnmount`: `feedsStore.stopUnreadPolling()`

   No template changes — already renders `feed.unread_count` reactively.

5. **Store tests** in `frontend/src/stores/__tests__/feeds.test.ts`:
   - Polling updates unread counts on feeds
   - Stop clears interval
   - Handles empty feeds list
