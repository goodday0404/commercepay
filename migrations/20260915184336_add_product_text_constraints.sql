-- +goose Up
ALTER TABLE products
ADD CONSTRAINT products_sku_not_blank
CHECK (btrim(sku) <> '');

ALTER TABLE products
ADD CONSTRAINT products_name_not_blank
CHECK (btrim(name) <> '');

ALTER TABLE products
ADD CONSTRAINT products_currency_not_blank
CHECK (btrim(currency) <> '');

-- +goose Down
ALTER TABLE products
DROP CONSTRAINT products_currency_not_blank;

ALTER TABLE products
DROP CONSTRAINT products_name_not_blank;

ALTER TABLE products
DROP CONSTRAINT products_sku_not_blank;