-- ใช้กับฐานที่มีอยู่แล้ว รันซ้ำได้
-- สินค้าเดิมทั้งร้านเป็นของ somchai@seller.com (id 2)

ALTER TABLE products ADD COLUMN IF NOT EXISTS seller_id BIGINT REFERENCES members (id);
UPDATE products SET seller_id = 2 WHERE seller_id IS NULL;

CREATE TABLE IF NOT EXISTS wishlists (
    member_id   BIGINT      NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    product_id  BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, product_id)
);

CREATE TABLE IF NOT EXISTS reviews (
    id          BIGSERIAL    PRIMARY KEY,
    product_id  BIGINT       NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    member_id   BIGINT       NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    rating      SMALLINT     NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment     TEXT         NOT NULL CHECK (char_length(btrim(comment)) >= 1),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (member_id, product_id)
);
CREATE INDEX IF NOT EXISTS idx_reviews_product ON reviews (product_id, created_at DESC);

CREATE TABLE IF NOT EXISTS carts (
    member_id   BIGINT      NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    product_id  BIGINT      NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    size        VARCHAR(20) NOT NULL,
    quantity    INTEGER     NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (member_id, product_id, size)
);

CREATE TABLE IF NOT EXISTS orders (
    id         BIGSERIAL    PRIMARY KEY,
    buyer_id   BIGINT       NOT NULL REFERENCES members (id),
    total      INTEGER      NOT NULL CHECK (total >= 0),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
    id           BIGSERIAL    PRIMARY KEY,
    order_id     BIGINT       NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id   BIGINT       REFERENCES products (id) ON DELETE SET NULL,
    seller_id    BIGINT       NOT NULL REFERENCES members (id),
    product_name VARCHAR(255) NOT NULL,
    size         VARCHAR(20)  NOT NULL,
    quantity     INTEGER      NOT NULL CHECK (quantity > 0),
    unit_price   INTEGER      NOT NULL CHECK (unit_price >= 0),
    status       VARCHAR(30)  NOT NULL DEFAULT 'รอการจัดส่ง',
    tracking_no  VARCHAR(50)  NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_order_items_seller ON order_items (seller_id);

CREATE TABLE IF NOT EXISTS disputes (
    id          BIGSERIAL    PRIMARY KEY,
    product_id  BIGINT       REFERENCES products (id) ON DELETE SET NULL,
    buyer_id    BIGINT       NOT NULL REFERENCES members (id),
    reason      TEXT         NOT NULL CHECK (char_length(btrim(reason)) >= 1),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    resolved    BOOLEAN      NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id          BIGSERIAL    PRIMARY KEY,
    product_id  BIGINT       NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    buyer_id    BIGINT       NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    sender      VARCHAR(10)  NOT NULL CHECK (sender IN ('user', 'seller')),
    body        TEXT         NOT NULL CHECK (char_length(btrim(body)) >= 1),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_chat_thread ON chat_messages (product_id, buyer_id, created_at);

CREATE TABLE IF NOT EXISTS seller_applications (
    id         BIGSERIAL    PRIMARY KEY,
    member_id  BIGINT       NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    shop_name  VARCHAR(80)  NOT NULL CHECK (char_length(btrim(shop_name)) >= 1),
    status     VARCHAR(20)  NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_seller_app_pending
    ON seller_applications (member_id)
    WHERE status = 'pending';
