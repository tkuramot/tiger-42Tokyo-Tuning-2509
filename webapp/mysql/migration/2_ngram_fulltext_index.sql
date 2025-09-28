USE `42Tokyo2508-db`;

ALTER TABLE products DROP INDEX idx_name_desc_fulltext;

ALTER TABLE products ADD FULLTEXT INDEX idx_name_desc_fulltext (name, description) WITH PARSER ngram;