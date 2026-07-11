package bootstrap

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ekse/rossoreader/internal/auth"
	"github.com/ekse/rossoreader/internal/store"
)

// Run performs one-time setup of multi-user ownership on the database.
// It is idempotent: each step guards against being re-applied.
//
//   - If there are no users, an admin user is created with a random password,
//     and the credentials are logged.
//   - Existing feeds/settings rows with no user_id are assigned to the first user.
//   - user_item_states is backfilled from items.read/starred.
//   - feeds and settings user_id columns are made NOT NULL.
//   - The feeds.url uniqueness constraint is replaced with UNIQUE(user_id, url).
//   - The settings primary key is replaced with PRIMARY KEY(user_id, key).
//   - The items.read and items.starred columns and their indexes are dropped.
//
// The schema changes here are intentionally AFTER the migrations:
// migrations add the new tables/columns and create new unique indexes,
// this step fills the new columns and drops the now obsolete constraints/columns.
func Run(ctx context.Context, pool *pgxpool.Pool, st store.Store) error {
	usingLegacyItemColumns, err := columnExists(ctx, pool, "items", "read")
	if err != nil {
		return fmt.Errorf("check legacy items.read column: %w", err)
	}
	if !usingLegacyItemColumns {
		// Nothing to do; bootstrap has already been completed in a prior run.
		return nil
	}

	// Step 1: Create the initial admin user if no users exist.
	count, err := st.CountUsers(ctx)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}

	var adminID int64
	if count == 0 {
		password, err := randomPassword(20)
		if err != nil {
			return fmt.Errorf("generate admin password: %w", err)
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hash admin password: %w", err)
		}

		u, err := st.CreateUser(ctx, "admin", hash, true)
		if err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
		adminID = u.ID
		log.Printf("Created admin user — username: admin password: %s", password)
	} else {
		users, err := st.ListUsers(ctx)
		if err != nil {
			return fmt.Errorf("list users: %w", err)
		}
		if len(users) == 0 {
			return fmt.Errorf("inconsistent state: count > 0 but list empty")
		}
		adminID = users[0].ID
	}

	// Step 2: Assign existing rows to the first user.
	if _, err := pool.Exec(ctx,
		"UPDATE feeds SET user_id = $1 WHERE user_id IS NULL", adminID); err != nil {
		return fmt.Errorf("assign feeds user_id: %w", err)
	}
	if _, err := pool.Exec(ctx,
		"UPDATE settings SET user_id = $1 WHERE user_id IS NULL", adminID); err != nil {
		return fmt.Errorf("assign settings user_id: %w", err)
	}

	// Step 3: Backfill user_item_states from existing read/starred item columns.
	if _, err := pool.Exec(ctx, `
		INSERT INTO user_item_states (user_id, item_id, read, starred)
		SELECT $1, i.id, i.read, i.starred
		FROM items i
		WHERE i.read = TRUE OR i.starred = TRUE
		ON CONFLICT (user_id, item_id) DO NOTHING
	`, adminID); err != nil {
		return fmt.Errorf("backfill user_item_states: %w", err)
	}

	// Step 4: Promote user_id columns to NOT NULL.
	if _, err := pool.Exec(ctx, "ALTER TABLE feeds ALTER COLUMN user_id SET NOT NULL"); err != nil {
		return fmt.Errorf("set feeds.user_id not null: %w", err)
	}
	if _, err := pool.Exec(ctx, "ALTER TABLE settings ALTER COLUMN user_id SET NOT NULL"); err != nil {
		return fmt.Errorf("set settings.user_id not null: %w", err)
	}

	// Step 5: Replace feeds.url unique constraint with (user_id, url).
	if _, err := pool.Exec(ctx, "ALTER TABLE feeds DROP CONSTRAINT IF EXISTS feeds_url_key"); err != nil {
		return fmt.Errorf("drop feeds_url_key: %w", err)
	}
	if _, err := pool.Exec(ctx,
		"ALTER TABLE feeds ADD CONSTRAINT feeds_user_url_key UNIQUE (user_id, url)"); err != nil {
		return fmt.Errorf("add feeds_user_url_key: %w", err)
	}

	// Step 6: Replace settings primary key with (user_id, key).
	if _, err := pool.Exec(ctx, "ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_pkey"); err != nil {
		return fmt.Errorf("drop settings_pkey: %w", err)
	}
	if _, err := pool.Exec(ctx,
		"ALTER TABLE settings ADD CONSTRAINT settings_pkey_new PRIMARY KEY (user_id, key)"); err != nil {
		return fmt.Errorf("add settings primary key: %w", err)
	}

	// Step 7: Drop legacy items.read/starred columns (and their indexes).
	if _, err := pool.Exec(ctx, "ALTER TABLE items DROP COLUMN read"); err != nil {
		return fmt.Errorf("drop items.read: %w", err)
	}
	if _, err := pool.Exec(ctx, "ALTER TABLE items DROP COLUMN starred"); err != nil {
		return fmt.Errorf("drop items.starred: %w", err)
	}

	if count == 0 {
		log.Printf("bootstrap complete: created admin user %d", adminID)
	} else {
		log.Printf("bootstrap complete: ownership migrated to user %d", adminID)
	}
	return nil
}

func columnExists(ctx context.Context, pool *pgxpool.Pool, table, column string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = $1 AND column_name = $2
		)
	`
	var exists bool
	if err := pool.QueryRow(ctx, q, table, column).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// randomPassword generates a URL-safe random password of the given length.
func randomPassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:length], nil
}
