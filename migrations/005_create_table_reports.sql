-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('monthly', 'yearly', 'custom')),
    title VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    data JSONB,
    -- Store report data as JSON
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
-- Create indexes for better performance
CREATE INDEX idx_reports_user_id ON reports(user_id);
CREATE INDEX idx_reports_type ON reports(type);
CREATE INDEX idx_reports_dates ON reports(start_date, end_date);
CREATE INDEX idx_reports_created_at ON reports(created_at);
-- Trigger to update updated_at automatically
CREATE TRIGGER trigger_update_reports_updated_at BEFORE
UPDATE ON reports FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
---- create above / drop below ----
DROP TABLE IF EXISTS reports;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.