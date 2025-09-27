-- このファイルに記述されたSQLコマンドが、マイグレーション時に実行されます。
ALTER TABLE `products` ADD INDEX `idx_name_product_id` (`name`, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_user_id` (`user_id`);
ALTER TABLE `orders` ADD INDEX `idx_product_id` (`product_id`);
ALTER TABLE `orders` ADD INDEX `idx_shipped_status_created_at` (`shipped_status`, `created_at`);
ALTER TABLE `users` ADD INDEX `idx_user_name` (`user_name`);
