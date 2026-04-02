-- Seed initial categories
INSERT INTO categories (name, icon, color, type, is_system) 
VALUES 
('Food & Dining', 'dining', '#FF5733', 'EXPENSE', true),
('Transportation', 'commute', '#3357FF', 'EXPENSE', true),
('Housing', 'home', '#33FF57', 'EXPENSE', true),
('Entertainment', 'movie', '#FF33A1', 'EXPENSE', true),
('Shopping', 'shopping_cart', '#FFD433', 'EXPENSE', true),
('Salary', 'attach_money', '#33FFD4', 'INCOME', true),
('Investments', 'trending_up', '#A133FF', 'INCOME', true);

-- Add sample user (for local testing/development)
-- id is UUID. using a fixed one for convenience.
-- 11111111-1111-1111-1111-111111111111
INSERT INTO users (id, email, name, avatar_url, status, currency_preference)
VALUES ('11111111-1111-1111-1111-111111111111', 'test@spendly.id', 'Test User', 'https://ui-avatars.com/api/?name=Test+User', 'ACTIVE', 'IDR');
