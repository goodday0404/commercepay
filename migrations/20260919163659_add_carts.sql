-- +goose Up

CREATE TABLE carts (
    id UUID PRIMARY KEY
);

CREATE TABLE cart_items (
    id UUID PRIMARY KEY,
    cart_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,

    CONSTRAINT cart_items_cart_id_fkey
        FOREIGN KEY (cart_id)
        REFERENCES carts(id)
        ON DELETE CASCADE,

    CONSTRAINT cart_items_product_id_fkey
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT,

    CONSTRAINT cart_items_quantity_positive
        CHECK (quantity > 0),

    CONSTRAINT cart_items_unit_price_minor_nonnegative
        CHECK (unit_price_minor >= 0),

    CONSTRAINT cart_items_currency_nonempty
        CHECK (btrim(currency) <> ''),

    CONSTRAINT cart_items_cart_product_unique
        UNIQUE (cart_id, product_id)
);

-- +goose Down

DROP TABLE cart_items;
DROP TABLE carts;
