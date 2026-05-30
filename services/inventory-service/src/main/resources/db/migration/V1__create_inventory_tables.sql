CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(1000),
    price DECIMAL(19, 2) NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_sku ON items(sku);
CREATE INDEX idx_category ON items(category);

CREATE TABLE stock_levels (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    warehouse_location VARCHAR(50) NOT NULL,
    quantity_on_hand INTEGER NOT NULL DEFAULT 0,
    quantity_reserved INTEGER NOT NULL DEFAULT 0,
    last_counted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(item_id, warehouse_location)
);

CREATE INDEX idx_item_location ON stock_levels(item_id, warehouse_location);
CREATE INDEX idx_warehouse ON stock_levels(warehouse_location);

CREATE TABLE reservations (
    id UUID PRIMARY KEY,
    order_id VARCHAR(100) NOT NULL,
    item_id BIGINT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity_reserved INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    released_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_order_id ON reservations(order_id);
CREATE INDEX idx_status ON reservations(status);
CREATE INDEX idx_expires_at ON reservations(expires_at);
