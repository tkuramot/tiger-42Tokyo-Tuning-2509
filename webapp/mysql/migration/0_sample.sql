-- このファイルに記述されたSQLコマンドが、マイグレーション時に実行されます。
ALTER TABLE `products` ADD INDEX `idx_name_product_id` (`name`, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_user_id` (`user_id`);
ALTER TABLE `orders` ADD INDEX `idx_product_id` (`product_id`);
ALTER TABLE `users` ADD INDEX `idx_user_name` (`user_name`);

ALTER TABLE `orders` ADD INDEX `idx_shipped_product` (`shipped_status`, `product_id`, `created_at`);
ALTER TABLE `products` ADD INDEX `idx_weight_value` (`weight`, `value` DESC);
