ALTER TABLE feeds ADD COLUMN user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE settings ADD COLUMN user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;

CREATE TABLE user_item_states (
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id  INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    read     BOOLEAN NOT NULL DEFAULT FALSE,
    starred  BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, item_id)
);

CREATE INDEX idx_user_item_states_user_id ON user_item_states(user_id);
CREATE INDEX idx_user_item_states_item_id ON user_item_states(item_id);