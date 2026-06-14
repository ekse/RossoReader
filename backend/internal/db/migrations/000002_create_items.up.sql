CREATE TABLE items (
    id           SERIAL PRIMARY KEY,
    feed_id      INTEGER NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    guid         TEXT NOT NULL,
    title        TEXT NOT NULL,
    url          TEXT NOT NULL,
    content      TEXT,
    description  TEXT,
    author       TEXT,
    published_at TIMESTAMPTZ,
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read         BOOLEAN NOT NULL DEFAULT FALSE,
    starred      BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(feed_id, guid)
);

CREATE INDEX idx_items_feed_id ON items(feed_id);
CREATE INDEX idx_items_published_at ON items(published_at DESC);
CREATE INDEX idx_items_read ON items(read);
CREATE INDEX idx_items_starred ON items(starred);
