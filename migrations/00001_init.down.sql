-- Удаление индексов
DROP INDEX IF EXISTS idx_order_items_book_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_books_category_id;

-- Удаление таблиц
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
