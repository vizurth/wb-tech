-- +goose Up
CREATE TABLE orders (
    order_uid         VARCHAR(64) PRIMARY KEY,
    track_number      VARCHAR(64) NOT NULL,
    entry             VARCHAR(16),
    locale            VARCHAR(8),
    internal_signature TEXT,
    customer_id       VARCHAR(64),
    delivery_service  VARCHAR(64),
    shardkey          VARCHAR(8),
    sm_id             INT,
    date_created      TIMESTAMP NOT NULL,
    oof_shard         VARCHAR(8)
);

CREATE TABLE deliveries (
    order_uid    VARCHAR(64) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name         VARCHAR(255),
    phone        VARCHAR(32),
    zip          VARCHAR(16),
    city         VARCHAR(128),
    address      TEXT,
    region       VARCHAR(128),
    email        VARCHAR(255)
);

CREATE TABLE payments (
    transaction    VARCHAR(64) PRIMARY KEY,
    order_uid      VARCHAR(64) REFERENCES orders(order_uid) ON DELETE CASCADE,
    request_id     VARCHAR(64),
    currency       VARCHAR(8),
    provider       VARCHAR(64),
    amount         INT,
    payment_dt     BIGINT,
    bank           VARCHAR(64),
    delivery_cost  INT,
    goods_total    INT,
    custom_fee     INT
);

CREATE TABLE items (
    id            SERIAL PRIMARY KEY,
    order_uid     VARCHAR(64) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id       BIGINT,
    track_number  VARCHAR(64),
    price         INT,
    rid           VARCHAR(64),
    name          VARCHAR(255),
    sale          INT,
    size          VARCHAR(16),
    total_price   INT,
    nm_id         BIGINT,
    brand         VARCHAR(128),
    status        INT
);

-- +goose Down
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS orders;
