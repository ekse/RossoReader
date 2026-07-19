package pgstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ekse/rossoreader/internal/db/generated"
	"github.com/ekse/rossoreader/internal/domain"
	"github.com/ekse/rossoreader/internal/store"
)

type PGStore struct {
	q *generated.Queries
}

func New(pool *pgxpool.Pool) *PGStore {
	return &PGStore{q: generated.New(pool)}
}

// Feeds

func (s *PGStore) GetFeeds(ctx context.Context, userID int64) ([]domain.Feed, error) {
	rows, err := s.q.GetFeeds(ctx, &userID)
	if err != nil {
		return nil, fmt.Errorf("get feeds: %w", err)
	}
	feeds := make([]domain.Feed, 0, len(rows))
	for _, r := range rows {
		feeds = append(feeds, toDomainFeed(r))
	}
	return feeds, nil
}

func (s *PGStore) GetAllFeeds(ctx context.Context) ([]domain.Feed, error) {
	rows, err := s.q.GetAllFeeds(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all feeds: %w", err)
	}
	feeds := make([]domain.Feed, 0, len(rows))
	for _, r := range rows {
		feeds = append(feeds, toDomainFeed(r))
	}
	return feeds, nil
}

func (s *PGStore) GetFeed(ctx context.Context, userID, id int64) (domain.Feed, error) {
	r, err := s.q.GetFeedByID(ctx, generated.GetFeedByIDParams{ID: int32(id), UserID: &userID})
	if err != nil {
		return domain.Feed{}, fmt.Errorf("get feed %d: %w", id, err)
	}
	return toDomainFeed(r), nil
}

func (s *PGStore) GetFeedByIDAny(ctx context.Context, id int64) (domain.Feed, error) {
	r, err := s.q.GetFeedByIDAny(ctx, int32(id))
	if err != nil {
		return domain.Feed{}, fmt.Errorf("get feed %d: %w", id, err)
	}
	return toDomainFeed(r), nil
}

func (s *PGStore) CreateFeed(ctx context.Context, userID int64, url, title, description, siteLink, iconURL, etag string) (domain.Feed, error) {
	r, err := s.q.CreateFeed(ctx, generated.CreateFeedParams{
		UserID:      &userID,
		Url:         url,
		Title:       title,
		Description: nullableString(description),
		SiteLink:    nullableString(siteLink),
		IconUrl:     nullableString(iconURL),
		Etag:        nullableString(etag),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "feeds_user_url_key" {
			return domain.Feed{}, domain.ErrFeedAlreadyExists
		}
		return domain.Feed{}, fmt.Errorf("create feed: %w", err)
	}
	return toDomainFeed(r), nil
}

func (s *PGStore) DeleteFeed(ctx context.Context, userID, id int64) error {
	err := s.q.DeleteFeed(ctx, generated.DeleteFeedParams{ID: int32(id), UserID: &userID})
	if err != nil {
		return fmt.Errorf("delete feed %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) UpdateFeedLastFetched(ctx context.Context, id int64) error {
	err := s.q.UpdateFeedLastFetched(ctx, int32(id))
	if err != nil {
		return fmt.Errorf("update feed last fetched %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) SetFeedFetchError(ctx context.Context, id int64, fetchError string) error {
	err := s.q.SetFeedFetchError(ctx, generated.SetFeedFetchErrorParams{
		ID:             int32(id),
		LastFetchError: nullableString(fetchError),
	})
	if err != nil {
		return fmt.Errorf("set feed fetch error %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) ClearFeedFetchError(ctx context.Context, id int64) error {
	err := s.q.ClearFeedFetchError(ctx, int32(id))
	if err != nil {
		return fmt.Errorf("clear feed fetch error %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) UpdateFeedEtag(ctx context.Context, id int64, etag string) error {
	err := s.q.UpdateFeedEtag(ctx, generated.UpdateFeedEtagParams{
		ID:   int32(id),
		Etag: nullableString(etag),
	})
	if err != nil {
		return fmt.Errorf("update feed etag %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) RenameFeed(ctx context.Context, userID, feedID int64, title string) error {
	err := s.q.RenameFeed(ctx, generated.RenameFeedParams{
		ID:     int32(feedID),
		Title:  title,
		UserID: &userID,
	})
	if err != nil {
		return fmt.Errorf("rename feed %d: %w", feedID, err)
	}
	return nil
}

func (s *PGStore) UpdateFeedMetadata(ctx context.Context, id int64, title, description, siteLink, iconURL string) error {
	err := s.q.UpdateFeedMetadata(ctx, generated.UpdateFeedMetadataParams{
		ID:          int32(id),
		Title:       title,
		Description: nullableString(description),
		SiteLink:    nullableString(siteLink),
		IconUrl:     nullableString(iconURL),
	})
	if err != nil {
		return fmt.Errorf("update feed metadata %d: %w", id, err)
	}
	return nil
}

// Items

func (s *PGStore) GetItems(ctx context.Context, q domain.ItemsQuery) ([]domain.Item, int64, error) {
	limit := int32(q.PerPage)
	if limit <= 0 {
		limit = 20
	}
	offset := int32((q.Page - 1) * q.PerPage)
	if offset < 0 {
		offset = 0
	}
	var feedID *int32
	if q.FeedID != nil {
		v := int32(*q.FeedID)
		feedID = &v
	}

	params := generated.GetItemsParams{
		UserID:  q.UserID,
		Limit:   limit,
		Offset:  offset,
		FeedID:  feedID,
		Read:    q.Read,
		Starred: q.Starred,
	}

	rows, err := s.q.GetItems(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("get items: %w", err)
	}

	countParams := generated.CountItemsParams{
		UserID:  q.UserID,
		FeedID:  feedID,
		Read:    q.Read,
		Starred: q.Starred,
	}
	total, err := s.q.CountItems(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("count items: %w", err)
	}

	items := make([]domain.Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, toDomainItem(r.ID, r.FeedID, r.Guid, r.Title, r.Url,
			r.Content, r.Description, r.Author, r.PublishedAt, r.FetchedAt,
			r.IsRead, r.IsStarred))
	}
	return items, total, nil
}

func (s *PGStore) SearchItems(ctx context.Context, q domain.SearchQuery) ([]domain.Item, int64, error) {
	limit := int32(q.PerPage)
	if limit <= 0 {
		limit = 20
	}
	offset := int32((q.Page - 1) * q.PerPage)
	if offset < 0 {
		offset = 0
	}

	feedIDs := make([]int32, 0, len(q.FeedIDs))
	for _, id := range q.FeedIDs {
		feedIDs = append(feedIDs, int32(id))
	}
	labelIDs := make([]int32, 0, len(q.LabelIDs))
	for _, id := range q.LabelIDs {
		labelIDs = append(labelIDs, int32(id))
	}

	params := generated.SearchItemsParams{
		UserID:   q.UserID,
		Query:    &q.Query,
		FeedIds:  feedIDs,
		LabelIds: labelIDs,
		Offset:   offset,
		Limit:    limit,
	}

	rows, err := s.q.SearchItems(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("search items: %w", err)
	}

	countParams := generated.CountSearchItemsParams{
		UserID:   q.UserID,
		Query:    &q.Query,
		FeedIds:  feedIDs,
		LabelIds: labelIDs,
	}
	total, err := s.q.CountSearchItems(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("count search items: %w", err)
	}

	items := make([]domain.Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, toDomainItem(r.ID, r.FeedID, r.Guid, r.Title, r.Url,
			r.Content, r.Description, r.Author, r.PublishedAt, r.FetchedAt,
			r.IsRead, r.IsStarred))
	}
	return items, total, nil
}

func (s *PGStore) GetItem(ctx context.Context, userID, id int64) (domain.Item, error) {
	r, err := s.q.GetItemByID(ctx, generated.GetItemByIDParams{ID: int32(id), UserID: userID})
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item %d: %w", id, err)
	}
	return toDomainItem(r.ID, r.FeedID, r.Guid, r.Title, r.Url,
		r.Content, r.Description, r.Author, r.PublishedAt, r.FetchedAt,
		r.IsRead, r.IsStarred), nil
}

func (s *PGStore) UpsertItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	r, err := s.q.UpsertItem(ctx, generated.UpsertItemParams{
		FeedID:      int32(item.FeedID),
		Guid:        item.GUID,
		Title:       item.Title,
		Url:         item.URL,
		Content:     item.Content,
		Description: item.Description,
		Author:      item.Author,
		PublishedAt: toTimestamptz(item.PublishedAt),
	})
	if err != nil {
		return domain.Item{}, fmt.Errorf("upsert item: %w", err)
	}
	return toDomainItem(r.ID, r.FeedID, r.Guid, r.Title, r.Url,
		r.Content, r.Description, r.Author, r.PublishedAt, r.FetchedAt,
		false, false), nil
}

func (s *PGStore) MarkItemRead(ctx context.Context, userID, id int64, read bool) error {
	err := s.q.SetItemRead(ctx, generated.SetItemReadParams{
		UserID: userID,
		ID:     int32(id),
		Read:   read,
	})
	if err != nil {
		return fmt.Errorf("mark item %d read: %w", id, err)
	}
	return nil
}

func (s *PGStore) MarkItemStarred(ctx context.Context, userID, id int64, starred bool) error {
	err := s.q.SetItemStarred(ctx, generated.SetItemStarredParams{
		UserID:  userID,
		ID:      int32(id),
		Starred: starred,
	})
	if err != nil {
		return fmt.Errorf("mark item %d starred: %w", id, err)
	}
	return nil
}

func (s *PGStore) MarkAllFeedItemsRead(ctx context.Context, userID, feedID int64) error {
	err := s.q.MarkFeedItemsReadForUser(ctx, generated.MarkFeedItemsReadForUserParams{
		UserID: userID,
		FeedID: int32(feedID),
	})
	if err != nil {
		return fmt.Errorf("mark all feed items read %d: %w", feedID, err)
	}
	return nil
}

func (s *PGStore) MarkAllItemsRead(ctx context.Context, userID int64) error {
	err := s.q.MarkAllItemsReadForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("mark all items read: %w", err)
	}
	return nil
}

func (s *PGStore) GetUnreadCountByFeed(ctx context.Context, userID int64) (map[int64]int, error) {
	rows, err := s.q.GetUnreadCountByFeedForUser(ctx, &userID)
	if err != nil {
		return nil, fmt.Errorf("get unread count by feed: %w", err)
	}
	result := make(map[int64]int, len(rows))
	for _, r := range rows {
		result[int64(r.FeedID)] = int(r.Count)
	}
	return result, nil
}

// Settings

func (s *PGStore) GetSettings(ctx context.Context, userID int64) (map[string]string, error) {
	rows, err := s.q.GetSettings(ctx, &userID)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.Key] = r.Value
	}
	return result, nil
}

func (s *PGStore) GetSetting(ctx context.Context, userID int64, key string) (string, error) {
	r, err := s.q.GetSetting(ctx, generated.GetSettingParams{UserID: &userID, Key: key})
	if err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return r.Value, nil
}

func (s *PGStore) UpsertSetting(ctx context.Context, userID int64, key, value string) error {
	err := s.q.UpsertSetting(ctx, generated.UpsertSettingParams{
		UserID: &userID,
		Key:    key,
		Value:  value,
	})
	if err != nil {
		return fmt.Errorf("upsert setting %q: %w", key, err)
	}
	return nil
}

func (s *PGStore) DeleteSetting(ctx context.Context, userID int64, key string) error {
	err := s.q.DeleteSetting(ctx, generated.DeleteSettingParams{UserID: &userID, Key: key})
	if err != nil {
		return fmt.Errorf("delete setting %q: %w", key, err)
	}
	return nil
}

// Users

func (s *PGStore) CreateUser(ctx context.Context, username, passwordHash string, isAdmin bool) (domain.User, error) {
	r, err := s.q.CreateUser(ctx, generated.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
		IsAdmin:      isAdmin,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_username_key" {
			return domain.User{}, domain.ErrUserAlreadyExists
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(r, r.PasswordHash), nil
}

func (s *PGStore) GetUserByUsername(ctx context.Context, username string) (domain.User, string, error) {
	r, err := s.q.GetUserByUsername(ctx, username)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("get user by username %q: %w", username, err)
	}
	return toDomainUser(r, r.PasswordHash), r.PasswordHash, nil
}

func (s *PGStore) GetUserByID(ctx context.Context, id int64) (domain.User, string, error) {
	r, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("get user %d: %w", id, err)
	}
	return toDomainUser(r, r.PasswordHash), r.PasswordHash, nil
}

func (s *PGStore) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, r := range rows {
		users = append(users, domain.User{
			ID:        r.ID,
			Username:  r.Username,
			IsAdmin:   r.IsAdmin,
			CreatedAt: r.CreatedAt.Time,
			UpdatedAt: r.UpdatedAt.Time,
		})
	}
	return users, nil
}

func (s *PGStore) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	err := s.q.UpdateUserPassword(ctx, generated.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("update user %d password: %w", id, err)
	}
	return nil
}

func (s *PGStore) DeleteUser(ctx context.Context, id int64) error {
	err := s.q.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) CountUsers(ctx context.Context) (int64, error) {
	count, err := s.q.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// Sessions

func (s *PGStore) CreateSession(ctx context.Context, id [16]byte, userID int64, expiresAt time.Time) error {
	_, err := s.q.CreateSession(ctx, generated.CreateSessionParams{
		ID:        pgUUIDFromBytes(id),
		UserID:    userID,
		ExpiresAt: toTimestamptz(&expiresAt),
	})
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *PGStore) GetSession(ctx context.Context, id [16]byte) (domain.Session, error) {
	r, err := s.q.GetSessionWithUser(ctx, pgUUIDFromBytes(id))
	if err != nil {
		return domain.Session{}, fmt.Errorf("get session: %w", err)
	}
	if !r.ExpiresAt.Valid || r.ExpiresAt.Time.Before(time.Now()) {
		return domain.Session{}, fmt.Errorf("session expired")
	}
	return domain.Session{
		ID:        id,
		ExpiresAt: r.ExpiresAt.Time,
		User: domain.User{
			ID:        r.UserID,
			Username:  r.Username,
			IsAdmin:   r.IsAdmin,
			CreatedAt: r.CreatedAt_2.Time,
			UpdatedAt: r.UpdatedAt.Time,
		},
	}, nil
}

func (s *PGStore) DeleteSession(ctx context.Context, id [16]byte) error {
	err := s.q.DeleteSession(ctx, pgUUIDFromBytes(id))
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *PGStore) DeleteExpiredSessions(ctx context.Context) error {
	err := s.q.DeleteExpiredSessions(ctx)
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}

// Passkeys

func (s *PGStore) CreatePasskey(ctx context.Context, userID int64, name string, credentialID, publicKey []byte, attestationType string, transports []string, signCount int64, backupEligible, backupState bool, aaguid []byte) (domain.Passkey, error) {
	r, err := s.q.CreatePasskey(ctx, generated.CreatePasskeyParams{
		UserID:          userID,
		Name:            name,
		CredentialID:    credentialID,
		PublicKey:       publicKey,
		AttestationType: attestationType,
		Transports:      transports,
		SignCount:       signCount,
		BackupEligible:  backupEligible,
		BackupState:     backupState,
		Aaguid:          aaguid,
	})
	if err != nil {
		return domain.Passkey{}, fmt.Errorf("create passkey: %w", err)
	}
	return toDomainPasskey(r), nil
}

func (s *PGStore) GetPasskeysByUserID(ctx context.Context, userID int64) ([]domain.Passkey, error) {
	rows, err := s.q.GetPasskeysByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get passkeys for user %d: %w", userID, err)
	}
	passkeys := make([]domain.Passkey, 0, len(rows))
	for _, r := range rows {
		passkeys = append(passkeys, toDomainPasskey(r))
	}
	return passkeys, nil
}

func (s *PGStore) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (domain.Passkey, error) {
	r, err := s.q.GetPasskeyByCredentialID(ctx, credentialID)
	if err != nil {
		return domain.Passkey{}, fmt.Errorf("get passkey by credential: %w", err)
	}
	return toDomainPasskey(r), nil
}

func (s *PGStore) UpdatePasskeySignCount(ctx context.Context, id int64, signCount int64) error {
	err := s.q.UpdatePasskeySignCount(ctx, generated.UpdatePasskeySignCountParams{
		ID:        id,
		SignCount: signCount,
	})
	if err != nil {
		return fmt.Errorf("update passkey %d sign count: %w", id, err)
	}
	return nil
}

func (s *PGStore) DeletePasskey(ctx context.Context, userID, id int64) error {
	err := s.q.DeletePasskey(ctx, generated.DeletePasskeyParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete passkey %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) SaveAuthState(ctx context.Context, id [16]byte, stateType string, stateData []byte, expiresAt time.Time) error {
	err := s.q.SaveAuthState(ctx, generated.SaveAuthStateParams{
		ID:        pgUUIDFromBytes(id),
		StateType: stateType,
		StateData: stateData,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("save auth state: %w", err)
	}
	return nil
}

func (s *PGStore) GetAuthState(ctx context.Context, id [16]byte) (string, []byte, error) {
	r, err := s.q.GetAuthState(ctx, pgUUIDFromBytes(id))
	if err != nil {
		return "", nil, fmt.Errorf("get auth state: %w", err)
	}
	return r.StateType, r.StateData, nil
}

func (s *PGStore) DeleteAuthState(ctx context.Context, id [16]byte) error {
	err := s.q.DeleteAuthState(ctx, pgUUIDFromBytes(id))
	if err != nil {
		return fmt.Errorf("delete auth state: %w", err)
	}
	return nil
}

func (s *PGStore) DeleteExpiredAuthStates(ctx context.Context) error {
	err := s.q.DeleteExpiredAuthStates(ctx)
	if err != nil {
		return fmt.Errorf("delete expired auth states: %w", err)
	}
	return nil
}

// Helpers

func toDomainPasskey(r generated.Passkey) domain.Passkey {
	return domain.Passkey{
		ID:              r.ID,
		UserID:          r.UserID,
		Name:            r.Name,
		CredentialID:    r.CredentialID,
		PublicKey:       r.PublicKey,
		AttestationType: r.AttestationType,
		Transports:      r.Transports,
		SignCount:       r.SignCount,
		BackupEligible:  r.BackupEligible,
		BackupState:     r.BackupState,
		AAGUID:          r.Aaguid,
		CreatedAt:       r.CreatedAt.Time,
		UpdatedAt:       r.UpdatedAt.Time,
	}
}

func toDomainFeed(r generated.Feed) domain.Feed {
	return domain.Feed{
		ID:             int64(r.ID),
		UserID:         ptrInt64OrZero(r.UserID),
		URL:            r.Url,
		Title:          r.Title,
		Description:    r.Description,
		SiteLink:       r.SiteLink,
		IconURL:        r.IconUrl,
		Etag:           r.Etag,
		LastFetchError: r.LastFetchError,
		LastFetchedAt:  fromTimestamptz(r.LastFetchedAt),
		CreatedAt:      r.CreatedAt.Time,
		UpdatedAt:      r.UpdatedAt.Time,
	}
}

func toDomainItem(
	id int32, feedID int32, guid, title, url string,
	content, description, author *string,
	publishedAt, fetchedAt pgtype.Timestamptz,
	read, starred bool,
) domain.Item {
	return domain.Item{
		ID:          int64(id),
		FeedID:      int64(feedID),
		GUID:        guid,
		Title:       title,
		URL:         url,
		Content:     content,
		Description: description,
		Author:      author,
		PublishedAt: fromTimestamptz(publishedAt),
		FetchedAt:   fetchedAt.Time,
		Read:        read,
		Starred:     starred,
	}
}

func toDomainUser(r generated.User, passwordHash string) domain.User {
	return domain.User{
		ID:        r.ID,
		Username:  r.Username,
		IsAdmin:   r.IsAdmin,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

// Global Settings

func (s *PGStore) GetItemsLimit(ctx context.Context) (int, error) {
	v, err := s.q.GetGlobalSetting(ctx, "items_limit")
	if err != nil {
		return store.DefaultItemsLimit, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return store.DefaultItemsLimit, nil
	}
	return n, nil
}

func (s *PGStore) SetItemsLimit(ctx context.Context, limit int) error {
	return s.q.UpsertGlobalSetting(ctx, generated.UpsertGlobalSettingParams{
		Key:   "items_limit",
		Value: strconv.Itoa(limit),
	})
}

func (s *PGStore) GetFeedsLimit(ctx context.Context) (int, error) {
	v, err := s.q.GetGlobalSetting(ctx, "feeds_limit")
	if err != nil {
		return store.DefaultFeedsLimit, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return store.DefaultFeedsLimit, nil
	}
	return n, nil
}

func (s *PGStore) SetFeedsLimit(ctx context.Context, limit int) error {
	return s.q.UpsertGlobalSetting(ctx, generated.UpsertGlobalSettingParams{
		Key:   "feeds_limit",
		Value: strconv.Itoa(limit),
	})
}

// Purge

func (s *PGStore) CountItemsByFeed(ctx context.Context, feedID int64) (int64, error) {
	n, err := s.q.CountItemsByFeed(ctx, int32(feedID))
	return int64(n), err
}

func (s *PGStore) DeleteExcessItems(ctx context.Context, feedID int64, maxItems int) (int64, error) {
	ids, err := s.q.DeleteExcessItems(ctx, generated.DeleteExcessItemsParams{
		FeedID:   int32(feedID),
		MaxItems: int32(maxItems),
	})
	if err != nil {
		return 0, fmt.Errorf("delete excess items: %w", err)
	}
	return int64(len(ids)), nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrInt64OrZero(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func pgUUIDFromBytes(b [16]byte) pgtype.UUID {
	return pgtype.UUID{Bytes: b, Valid: true}
}

// Labels

func (s *PGStore) GetLabels(ctx context.Context, userID int64) ([]domain.Label, error) {
	rows, err := s.q.GetLabels(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}
	labels := make([]domain.Label, 0, len(rows))
	for _, r := range rows {
		labels = append(labels, toDomainLabel(r))
	}
	return labels, nil
}

func (s *PGStore) GetLabel(ctx context.Context, userID, labelID int64) (domain.Label, error) {
	r, err := s.q.GetLabelByID(ctx, generated.GetLabelByIDParams{
		ID:     int32(labelID),
		UserID: userID,
	})
	if err != nil {
		return domain.Label{}, fmt.Errorf("get label %d: %w", labelID, err)
	}
	return toDomainLabel(r), nil
}

func (s *PGStore) CreateLabel(ctx context.Context, userID int64, name string) (domain.Label, error) {
	r, err := s.q.CreateLabel(ctx, generated.CreateLabelParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Label{}, domain.ErrLabelAlreadyExists
		}
		return domain.Label{}, fmt.Errorf("create label: %w", err)
	}
	return toDomainLabel(r), nil
}

func (s *PGStore) UpdateLabel(ctx context.Context, userID, labelID int64, name string) error {
	err := s.q.UpdateLabel(ctx, generated.UpdateLabelParams{
		ID:     int32(labelID),
		Name:   name,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("update label %d: %w", labelID, err)
	}
	return nil
}

func (s *PGStore) DeleteLabel(ctx context.Context, userID, labelID int64) error {
	err := s.q.DeleteLabel(ctx, generated.DeleteLabelParams{
		ID:     int32(labelID),
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete label %d: %w", labelID, err)
	}
	return nil
}

func (s *PGStore) AddFeedLabel(ctx context.Context, userID, feedID, labelID int64) error {
	// Verify feed ownership before adding label.
	_, err := s.GetFeed(ctx, userID, feedID)
	if err != nil {
		return err
	}
	// Verify label ownership.
	_, err = s.GetLabel(ctx, userID, labelID)
	if err != nil {
		return err
	}
	err = s.q.AddFeedLabel(ctx, generated.AddFeedLabelParams{
		FeedID:  int32(feedID),
		LabelID: int32(labelID),
	})
	if err != nil {
		return fmt.Errorf("add feed label: %w", err)
	}
	return nil
}

func (s *PGStore) RemoveFeedLabel(ctx context.Context, userID, feedID, labelID int64) error {
	err := s.q.RemoveFeedLabel(ctx, generated.RemoveFeedLabelParams{
		FeedID:  int32(feedID),
		LabelID: int32(labelID),
	})
	if err != nil {
		return fmt.Errorf("remove feed label: %w", err)
	}
	return nil
}

func (s *PGStore) GetFeedLabels(ctx context.Context, userID, feedID int64) ([]domain.Label, error) {
	rows, err := s.q.GetFeedLabels(ctx, generated.GetFeedLabelsParams{
		FeedID: int32(feedID),
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get feed labels: %w", err)
	}
	labels := make([]domain.Label, 0, len(rows))
	for _, r := range rows {
		labels = append(labels, toDomainLabel(r))
	}
	return labels, nil
}

func (s *PGStore) GetFeedIDsByLabel(ctx context.Context, userID int64) (map[int64][]int64, error) {
	assignments, err := s.q.GetUserLabelAssignments(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user label assignments: %w", err)
	}
	result := make(map[int64][]int64)
	for _, a := range assignments {
		result[int64(a.LabelID)] = append(result[int64(a.LabelID)], int64(a.FeedID))
	}
	return result, nil
}

func toDomainLabel(r generated.Label) domain.Label {
	return domain.Label{
		ID:        int64(r.ID),
		UserID:    r.UserID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt.Time,
	}
}
