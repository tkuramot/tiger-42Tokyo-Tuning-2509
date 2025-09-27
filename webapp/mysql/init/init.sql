USE `42Tokyo2508-db`;

DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `user_id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  `password_hash` VARCHAR(255) NOT NULL,
  `user_name` VARCHAR(255) NOT NULL
  );

-- LOAD DATA INFILE '/docker-entrypoint-initdb.d/csv/users.csv'
-- INTO TABLE users
-- FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n'
-- IGNORE 1 ROWS
-- (password_hash, user_name);


-- productsテーブルの作成
CREATE TABLE products (
    product_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value INT UNSIGNED NOT NULL,
    weight INT UNSIGNED NOT NULL,
    image VARCHAR(500),
    description TEXT
) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

-- 商品一覧のソート／検索を高速化する補助インデックス
CREATE INDEX idx_products_value ON products (value);
CREATE INDEX idx_products_weight ON products (weight);
CREATE FULLTEXT INDEX idx_products_name_description ON products (name, description);

-- LOAD DATA INFILE '/docker-entrypoint-initdb.d/csv/products.csv'
-- INTO TABLE products
-- FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n'
-- IGNORE 1 ROWS
-- ( name, value, weight, image, description);

CREATE TABLE orders (
    order_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    shipped_status VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL,
    arrived_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
);

-- 履歴照会と配送候補抽出のためのインデックス
CREATE INDEX idx_orders_user_created ON orders (user_id, created_at);
CREATE INDEX idx_orders_status_created ON orders (shipped_status, created_at);

-- LOAD DATA INFILE '/docker-entrypoint-initdb.d/csv/orders.csv'
-- INTO TABLE orders
-- FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n'
-- IGNORE 1 ROWS
-- (user_id, product_id, shipped_status, created_at, arrived_at);

CREATE TABLE `user_sessions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `session_uuid` VARCHAR(36) NOT NULL,
  `user_id` INT UNSIGNED NOT NULL,
  `expires_at` DATETIME NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `session_uuid` (`session_uuid`),
  FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- 期限切れセッションの検索を支援
CREATE INDEX idx_user_sessions_expires_at ON user_sessions (expires_at);
