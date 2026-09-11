-- +goose Up
CREATE INDEX products_created_at_id_idx
ON products (created_at DESC, id DESC);

-- +goose Down
DROP INDEX products_created_at_id_idx;