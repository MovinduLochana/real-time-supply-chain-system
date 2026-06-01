-- Database schema initialization for Analytics Service

-- Orders fact table
CREATE TABLE IF NOT EXISTS orders_fact (
    order_id VARCHAR PRIMARY KEY,
    customer_id VARCHAR NOT NULL,
    total_amount DOUBLE NOT NULL,
    item_count INTEGER NOT NULL,
    status VARCHAR DEFAULT 'created',
    created_at TIMESTAMP NOT NULL,
    cancelled_at TIMESTAMP,
    sku_list VARCHAR[],
    event_id VARCHAR UNIQUE,
    ingested_at TIMESTAMP DEFAULT current_timestamp,
    INDEX idx_customer_id (customer_id),
    INDEX idx_created_at (created_at),
    INDEX idx_status (status)
);

-- Inventory fact table
CREATE TABLE IF NOT EXISTS inventory_fact (
    sku VARCHAR NOT NULL,
    warehouse_id VARCHAR,
    previous_quantity INTEGER,
    new_quantity INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    alert_threshold INTEGER,
    is_low_stock BOOLEAN DEFAULT false,
    event_id VARCHAR UNIQUE,
    ingested_at TIMESTAMP DEFAULT current_timestamp,
    PRIMARY KEY (sku, warehouse_id, updated_at),
    INDEX idx_sku (sku),
    INDEX idx_updated_at (updated_at),
    INDEX idx_low_stock (is_low_stock)
);

-- Customers dimension table
CREATE TABLE IF NOT EXISTS customers_dim (
    customer_id VARCHAR PRIMARY KEY,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    total_orders INTEGER DEFAULT 0,
    total_spent DOUBLE DEFAULT 0.0,
    is_repeat_customer BOOLEAN DEFAULT false,
    updated_at TIMESTAMP DEFAULT current_timestamp,
    INDEX idx_total_spent (total_spent),
    INDEX idx_is_repeat (is_repeat_customer)
);

-- Forecasts table
CREATE TABLE IF NOT EXISTS forecasts (
    forecast_id VARCHAR PRIMARY KEY,
    sku VARCHAR NOT NULL,
    forecast_days INTEGER NOT NULL,
    model_type VARCHAR NOT NULL,
    forecast_data VARCHAR NOT NULL,
    confidence_level DOUBLE,
    generated_at TIMESTAMP NOT NULL,
    valid_until TIMESTAMP,
    event_id VARCHAR UNIQUE,
    INDEX idx_sku_generated (sku, generated_at),
    INDEX idx_valid_until (valid_until)
);

-- Metrics cache table
CREATE TABLE IF NOT EXISTS metrics_cache (
    metric_key VARCHAR PRIMARY KEY,
    metric_type VARCHAR NOT NULL,
    metric_value VARCHAR NOT NULL,
    cached_at TIMESTAMP DEFAULT current_timestamp,
    expires_at TIMESTAMP,
    INDEX idx_metric_type (metric_type),
    INDEX idx_expires_at (expires_at)
);
