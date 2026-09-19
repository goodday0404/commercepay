-- +goose Up
ALTER TABLE products
RENAME CONSTRAINT products_sku_key TO products_sku_unique;

-- +goose Down
ALTER TABLE products
RENAME CONSTRAINT products_sku_unique TO products_sku_key;
