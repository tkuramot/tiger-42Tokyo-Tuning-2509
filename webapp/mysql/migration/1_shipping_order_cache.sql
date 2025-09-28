USE `42Tokyo2508-db`;

CREATE TABLE IF NOT EXISTS shipping_order_cache (
    order_id INT UNSIGNED PRIMARY KEY,
    weight INT UNSIGNED NOT NULL,
    value INT UNSIGNED NOT NULL,
    INDEX idx_shipping_lookup (weight, value DESC),
    FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE
) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

ALTER TABLE `shipping_order_cache` ADD INDEX `idx_weight_value` (`weight`, `value` DESC);

INSERT INTO shipping_order_cache (order_id, weight, value)
SELECT o.order_id, p.weight, p.value
FROM orders o
JOIN products p ON o.product_id = p.product_id
WHERE o.shipped_status = 'shipping';
