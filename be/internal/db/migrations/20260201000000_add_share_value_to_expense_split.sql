-- +goose Up
-- +goose StatementBegin
-- Add share_value column for storing percentage or shares
ALTER TABLE expense_split
ADD COLUMN share_value DECIMAL(10, 2) CHECK (share_value IS NULL OR share_value >= 0);

-- Add comment for clarity
COMMENT ON COLUMN expense_split.share_value IS
  'For percentage splits: the percentage value (0-100). For shares splits: the share count. NULL for equal/fixed/custom.';

-- Update split_type CHECK constraint to include 'shares'
ALTER TABLE expense_split
DROP CONSTRAINT IF EXISTS expense_split_split_type_check;

ALTER TABLE expense_split
ADD CONSTRAINT expense_split_split_type_check
CHECK (split_type IN ('equal', 'percentage', 'shares', 'fixed', 'custom'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove share_value column
ALTER TABLE expense_split
DROP COLUMN IF EXISTS share_value;

-- Restore original CHECK constraint
ALTER TABLE expense_split
DROP CONSTRAINT IF EXISTS expense_split_split_type_check;

ALTER TABLE expense_split
ADD CONSTRAINT expense_split_split_type_check
CHECK (split_type IN ('equal', 'percentage', 'fixed', 'custom'));
-- +goose StatementEnd
