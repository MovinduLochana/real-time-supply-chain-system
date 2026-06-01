"""Database queries and schema definitions."""

# Table creation queries
CREATE_ORDERS_TABLE = """
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
        ingested_at TIMESTAMP DEFAULT current_timestamp
    )
"""

CREATE_INVENTORY_TABLE = """
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
        PRIMARY KEY (sku, warehouse_id, updated_at)
    )
"""

CREATE_CUSTOMERS_TABLE = """
    CREATE TABLE IF NOT EXISTS customers_dim (
        customer_id VARCHAR PRIMARY KEY,
        first_seen TIMESTAMP NOT NULL,
        last_seen TIMESTAMP NOT NULL,
        total_orders INTEGER DEFAULT 0,
        total_spent DOUBLE DEFAULT 0.0,
        is_repeat_customer BOOLEAN DEFAULT false,
        updated_at TIMESTAMP DEFAULT current_timestamp
    )
"""

CREATE_FORECASTS_TABLE = """
    CREATE TABLE IF NOT EXISTS forecasts (
        forecast_id VARCHAR PRIMARY KEY,
        sku VARCHAR NOT NULL,
        forecast_days INTEGER NOT NULL,
        model_type VARCHAR NOT NULL,
        forecast_data VARCHAR NOT NULL,
        confidence_level DOUBLE,
        generated_at TIMESTAMP NOT NULL,
        valid_until TIMESTAMP,
        event_id VARCHAR UNIQUE
    )
"""

# Aggregation queries
GET_ORDER_METRICS = """
    SELECT
        COUNT(DISTINCT order_id) as total_orders,
        SUM(total_amount) as total_revenue,
        AVG(total_amount) as avg_order_value,
        COUNT(DISTINCT order_id) / CAST(DATEDIFF('day', MIN(created_at), MAX(created_at)) + 1 AS DOUBLE) as order_rate_per_day,
        MIN(created_at) as period_start,
        MAX(created_at) as period_end,
        current_timestamp as timestamp
    FROM orders_fact
    WHERE status = 'created'
    AND created_at >= CURRENT_DATE - INTERVAL 30 DAY
"""

GET_ORDER_METRICS_BY_DATE = """
    SELECT
        DATE(created_at) as date,
        COUNT(DISTINCT order_id) as orders,
        SUM(total_amount) as revenue,
        AVG(total_amount) as avg_order_value
    FROM orders_fact
    WHERE status = 'created'
    AND created_at >= ?
    AND created_at <= ?
    GROUP BY DATE(created_at)
    ORDER BY date
"""

GET_INVENTORY_METRICS = """
    SELECT
        COUNT(DISTINCT sku) as total_skus,
        SUM(new_quantity) as total_stock_quantity,
        SUM(CASE WHEN is_low_stock THEN 1 ELSE 0 END) as low_stock_count,
        AVG(CASE WHEN new_quantity > 0 THEN new_quantity ELSE NULL END) as avg_stock_level
    FROM (
        SELECT DISTINCT ON (sku) sku, new_quantity, is_low_stock
        FROM inventory_fact
        ORDER BY sku, updated_at DESC
    ) latest_inventory
"""

GET_INVENTORY_BY_SKU = """
    SELECT
        sku,
        new_quantity as current_stock,
        SUM(CASE WHEN previous_quantity > 0 THEN new_quantity - previous_quantity ELSE 0 END) as net_change,
        COUNT(*) as update_count,
        MAX(updated_at) as last_updated
    FROM inventory_fact
    WHERE updated_at >= ? AND updated_at <= ?
    GROUP BY sku
    ORDER BY sku
"""

GET_CUSTOMER_METRICS = """
    SELECT
        COUNT(DISTINCT customer_id) as total_customers,
        SUM(CASE WHEN DATE(first_seen) = CURRENT_DATE THEN 1 ELSE 0 END) as new_customers_this_period,
        SUM(CASE WHEN is_repeat_customer THEN 1 ELSE 0 END) as repeat_customer_count,
        CAST(SUM(CASE WHEN is_repeat_customer THEN 1 ELSE 0 END) AS DOUBLE) / NULLIF(COUNT(DISTINCT customer_id), 0) as repeat_rate,
        AVG(total_spent) as avg_clv
    FROM customers_dim
"""

GET_LOW_STOCK_ITEMS = """
    SELECT
        sku,
        new_quantity as current_stock,
        alert_threshold,
        MAX(updated_at) as last_alert
    FROM inventory_fact
    WHERE is_low_stock = true
    GROUP BY sku, new_quantity, alert_threshold
    ORDER BY new_quantity ASC
"""

GET_TOP_REVENUE_SKUS = """
    SELECT
        sku,
        COUNT(*) as order_count,
        SUM(total_amount) / CAST(COUNT(*) AS DOUBLE) as avg_revenue_per_order
    FROM orders_fact, UNNEST(sku_list) as sku(value)
    WHERE status = 'created'
    AND created_at >= ?
    GROUP BY sku
    ORDER BY avg_revenue_per_order DESC
    LIMIT ?
"""

GET_CUSTOMER_LIFETIME_VALUE = """
    SELECT
        customer_id,
        total_orders,
        total_spent,
        total_spent / NULLIF(total_orders, 0) as avg_order_value,
        DATEDIFF('day', first_seen, last_seen) as customer_tenure_days
    FROM customers_dim
    WHERE total_spent > 0
    ORDER BY total_spent DESC
    LIMIT ?
"""

# Update queries
UPDATE_CUSTOMER_METRICS = """
    INSERT INTO customers_dim
    SELECT
        customer_id,
        MIN(CASE WHEN first_seen IS NULL THEN created_at ELSE first_seen END) as first_seen,
        MAX(created_at) as last_seen,
        COUNT(*) as total_orders,
        SUM(total_amount) as total_spent,
        COUNT(*) > 1 as is_repeat_customer,
        current_timestamp as updated_at
    FROM orders_fact
    WHERE customer_id NOT IN (SELECT customer_id FROM customers_dim)
    GROUP BY customer_id
    ON CONFLICT (customer_id) DO UPDATE SET
        last_seen = EXCLUDED.last_seen,
        total_orders = EXCLUDED.total_orders,
        total_spent = EXCLUDED.total_spent,
        is_repeat_customer = EXCLUDED.is_repeat_customer,
        updated_at = current_timestamp
"""
