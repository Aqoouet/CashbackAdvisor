-- Enable pg_trgm extension for fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create cashback_rules table
CREATE TABLE IF NOT EXISTS cashback_rules (
    id BIGSERIAL PRIMARY KEY,
    group_name TEXT NOT NULL,
    category TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    user_display_name TEXT NOT NULL,
    month_year DATE NOT NULL,
    cashback_percent NUMERIC(5,2) NOT NULL CHECK (cashback_percent >= 0.00 AND cashback_percent <= 100.00),
    max_amount NUMERIC(10,2) NOT NULL CHECK (max_amount >= 0.00),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create trigram indexes for fuzzy search
CREATE INDEX idx_group_trgm ON cashback_rules USING GIN (group_name gin_trgm_ops);
CREATE INDEX idx_category_trgm ON cashback_rules USING GIN (category gin_trgm_ops);
CREATE INDEX idx_bank_trgm ON cashback_rules USING GIN (bank_name gin_trgm_ops);
CREATE INDEX idx_user_name_trgm ON cashback_rules USING GIN (user_display_name gin_trgm_ops);

-- Create composite index for common queries
CREATE INDEX idx_group_month_cat ON cashback_rules (group_name, month_year, category);

-- Create index for user queries
CREATE INDEX idx_user_id ON cashback_rules (user_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_cashback_rules_updated_at
    BEFORE UPDATE ON cashback_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

