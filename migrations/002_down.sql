-- Rollback script
DROP TRIGGER IF EXISTS update_cashback_rules_updated_at ON cashback_rules;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS cashback_rules;
DROP EXTENSION IF EXISTS pg_trgm;

