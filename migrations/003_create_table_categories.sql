-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    color VARCHAR(7) NOT NULL,
    -- Hex color like #FF5733
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);
-- Create index for better performance
CREATE INDEX idx_categories_user_id ON categories(user_id);
CREATE INDEX idx_categories_is_active ON categories(is_active);
-- Trigger to update updated_at automatically
CREATE TRIGGER trigger_update_categories_updated_at BEFORE
UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
---- create above / drop below ----
DROP TABLE IF EXISTS categories;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.