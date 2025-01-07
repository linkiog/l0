-- Таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR PRIMARY KEY,
    track_number VARCHAR,
    entry VARCHAR,
    locale VARCHAR,
    internal_signature VARCHAR,
    customer_id VARCHAR,
    delivery_service VARCHAR,
    shardkey VARCHAR,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR
);

-- Таблица с данными о доставке
CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR PRIMARY KEY,
    name VARCHAR,
    phone VARCHAR,
    zip VARCHAR,
    city VARCHAR,
    address VARCHAR,
    region VARCHAR,
    email VARCHAR,
    CONSTRAINT fk_delivery_order
        FOREIGN KEY (order_uid)
        REFERENCES orders (order_uid)
        ON DELETE CASCADE
);

-- Таблица с данными об оплате
CREATE TABLE IF NOT EXISTS payment (
    order_uid VARCHAR PRIMARY KEY,
    transaction VARCHAR,
    request_id VARCHAR,
    currency VARCHAR,
    provider VARCHAR,
    amount INT,
    payment_dt BIGINT,
    bank VARCHAR,
    delivery_cost INT,
    goods_total INT,
    custom_fee INT,
    CONSTRAINT fk_payment_order
        FOREIGN KEY (order_uid)
        REFERENCES orders (order_uid)
        ON DELETE CASCADE
);

-- Таблица с данными о товарах (items)
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR NOT NULL,
    chrt_id BIGINT,
    track_number VARCHAR,
    price INT,
    rid VARCHAR,
    name VARCHAR,
    sale INT,
    size VARCHAR,
    total_price INT,
    nm_id BIGINT,
    brand VARCHAR,
    status INT,
    CONSTRAINT fk_items_order
        FOREIGN KEY (order_uid)
        REFERENCES orders (order_uid)
        ON DELETE CASCADE
);
