-- このファイルに記述されたSQLコマンドが、マイグレーション時に実行されます。
ALTER TABLE `products` ADD INDEX `idx_name_product_id` (`name`, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_user_id` (`user_id`);
ALTER TABLE `orders` ADD INDEX `idx_product_id` (`product_id`);
ALTER TABLE `orders` ADD INDEX `idx_shipped_status_created_at` (`shipped_status`, `created_at`);
ALTER TABLE `users` ADD INDEX `idx_user_name` (`user_name`);

ALTER TABLE `orders` ADD INDEX `idx_shipped_product` (`shipped_status`, `product_id`, `created_at`);
ALTER TABLE `products` ADD INDEX `idx_weight_value` (`weight`, `value` DESC);
ALTER TABLE `products` ADD FULLTEXT INDEX `idx_name_desc_fulltext` (`name`, `description`) WITH PARSER ngram;

ALTER TABLE `user_sessions` ADD INDEX `idx_session_expires` (`session_uuid`, `expires_at`, `user_id`);
ALTER TABLE `orders` ADD INDEX `idx_user_product` (`user_id`, `product_id`);
ALTER TABLE `products` ADD INDEX `idx_name_prefix` (`name`(50));

ALTER TABLE `orders` ADD INDEX `idx_user_created` (`user_id`, `created_at`, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_user_shipped` (`user_id`, `shipped_status`, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_user_arrived` (`user_id`, `arrived_at`, `product_id`);

ALTER TABLE `products` ADD INDEX `idx_weight_value_pid` (`weight`, `value` DESC, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_shipped_created_pid` (`shipped_status`, `created_at`, `product_id`);

ALTER TABLE `products` ADD INDEX `idx_name_desc_product_id` (`name` DESC, `product_id`);
ALTER TABLE `products` ADD INDEX `idx_value_product_id` (`value`, `product_id`);
ALTER TABLE `products` ADD INDEX `idx_weight_desc_product_id` (`weight` DESC, `product_id`);
ALTER TABLE `products` ADD INDEX `idx_value_desc_product_id` (`value` DESC, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_shipped_weight_value_created` (`shipped_status`, `product_id`);
ALTER TABLE `orders` ADD INDEX `idx_shipped_created` (`shipped_status`, `created_at`);
