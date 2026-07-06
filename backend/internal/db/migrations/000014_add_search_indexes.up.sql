CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);
CREATE INDEX idx_items_description_trgm ON items USING GIN (description gin_trgm_ops);
CREATE INDEX idx_items_content_trgm ON items USING GIN (content gin_trgm_ops);
