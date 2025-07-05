-- Исправление длины поля ID в таблице users
ALTER TABLE users ALTER COLUMN id TYPE VARCHAR(36); 