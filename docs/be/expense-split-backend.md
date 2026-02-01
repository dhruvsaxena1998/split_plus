# Plan: Move Split Calculations to Backend

## Overview
Refactor the expense creation flow to move split amount calculations from frontend to backend. This ensures accuracy, consistency, and keeps the frontend "dumb."

## Current vs New Architecture

### Current (Frontend Calculates)
```
Frontend → calculates all amounts → sends {user_id, amount_owned, split_type}
Backend → validates totals match → stores as-is
```

### New (Backend Calculates)
```
Frontend → sends minimal data based on split_type → {user_id, type, percentage?, amount?}
Backend → calculates amounts (equal/percentage) OR validates (fixed/custom) → stores calculated amounts
```

## New Request Structure

```json
{
  "title": "Team Lunch",
  "notes": "optional",
  "amount": "120.00",
  "currency_code": "USD",
  "date": "2026-01-31",
  "payments": [
    { "user_id": "uuid", "amount": "120.00" }
  ],
  "splits": [
    { "user_id": "uuid1", "type": "equal" },
    { "user_id": "uuid2", "type": "equal" },
    { "user_id": "uuid3", "type": "equal" }
  ]
}
```

### Split Input by Type

| Split Type | Required Fields | Optional Fields | Backend Action |
|------------|----------------|-----------------|----------------|
| `equal` | `user_id`, `type` | - | Calculate: total ÷ count |
| `percentage` | `user_id`, `type`, `percentage` | - | Calculate: total × (percentage/100) |
| `shares` | `user_id`, `type`, `shares` | - | Calculate: total × (shares/total_shares) |
| `fixed` | `user_id`, `type`, `amount` | - | Use provided amount |
| `custom` | `user_id`, `type`, `amount` | - | Use provided amount |

**Important:** All splits in an expense MUST have the same type. Mixed types are not allowed.

## Implementation Plan

### Phase 1: Update HTTP Handler Request Types

**File:** `be/internal/http/handlers/expenses.go`

Update `SplitRequest` struct:
```go
type SplitRequest struct {
    UserID        string  `json:"user_id"`         // UUID
    PendingUserID string  `json:"pending_user_id"` // UUID (alternative to user_id)
    Type          string  `json:"type"`            // Required: 'equal', 'percentage', 'shares', 'fixed', 'custom'
    Percentage    *string `json:"percentage"`      // Required for 'percentage' type (e.g., "60.00")
    Shares        *int    `json:"shares"`          // Required for 'shares' type (e.g., 2)
    Amount        *string `json:"amount"`          // Required for 'fixed' and 'custom' types
}
```

Update validation in `validateSplitRequest()`:
- Type is required (must be one of: equal, percentage, shares, fixed, custom)
- **All splits must have the same type** (validate at expense level)
- For `percentage`: percentage field is required, must be > 0
- For `shares`: shares field is required, must be > 0 (integer)
- For `fixed`/`custom`: amount field is required
- For `equal`: neither percentage, shares, nor amount required

### Phase 2: Update Service Layer Input Types

**File:** `be/internal/service/expense_service.go`

Update `SplitInput` struct:
```go
type SplitInput struct {
    UserID        pgtype.UUID
    PendingUserID *pgtype.UUID
    Type          string   // Required: 'equal', 'percentage', 'shares', 'fixed', 'custom'
    Percentage    *string  // For percentage splits (e.g., "60.00")
    Shares        *int     // For shares splits (e.g., 2)
    Amount        *string  // For fixed/custom splits
}

// calculatedSplit is the result after backend calculation
type calculatedSplit struct {
    UserID        pgtype.UUID
    PendingUserID *pgtype.UUID
    Type          string
    Amount        string  // Calculated amount
    ShareValue    *string // Original percentage or shares (for DB storage)
}
```

### Phase 3: Add Calculation Logic in Service Layer

**File:** `be/internal/service/expense_service.go`

Add new function `calculateSplitAmounts()`:

```go
func (s *expenseService) calculateSplitAmounts(expenseAmount string, splits []SplitInput) ([]calculatedSplit, error) {
    // 1. Validate all splits have same type
    splitType := splits[0].Type
    for _, split := range splits {
        if split.Type != splitType {
            return nil, ErrMixedSplitTypes
        }
    }

    // 2. Calculate based on type
    switch splitType {
    case "equal":
        return s.calculateEqualSplits(expenseAmount, splits)
    case "percentage":
        return s.calculatePercentageSplits(expenseAmount, splits)
    case "shares":
        return s.calculateSharesSplits(expenseAmount, splits)
    case "fixed", "custom":
        return s.validateFixedSplits(expenseAmount, splits)
    default:
        return nil, ErrInvalidSplitType
    }
}
```

#### Equal Split Calculation
```go
func (s *expenseService) calculateEqualSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
    totalDec, _ := decimal.NewFromString(total)
    count := len(splits)

    // Calculate base amount (rounded down to 2 decimal places)
    baseAmount := totalDec.Div(decimal.NewFromInt(int64(count))).Round(2)

    // Calculate remainder to add to last split
    allocatedTotal := baseAmount.Mul(decimal.NewFromInt(int64(count - 1)))
    lastAmount := totalDec.Sub(allocatedTotal)

    result := make([]calculatedSplit, count)
    for i, split := range splits {
        if i == count-1 {
            result[i].Amount = lastAmount.String()
        } else {
            result[i].Amount = baseAmount.String()
        }
        result[i].UserID = split.UserID
        result[i].Type = "equal"
    }
    return result, nil
}
```

#### Percentage Split Calculation
```go
func (s *expenseService) calculatePercentageSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
    totalDec, _ := decimal.NewFromString(total)

    // Validate percentages sum to 100
    var percentageSum decimal.Decimal
    for _, split := range splits {
        pct, _ := decimal.NewFromString(*split.Percentage)
        percentageSum = percentageSum.Add(pct)
    }
    if !percentageSum.Equal(decimal.NewFromInt(100)) {
        return nil, ErrPercentageTotalMismatch
    }

    // Calculate amounts, adjust last for rounding
    result := make([]calculatedSplit, len(splits))
    var allocatedTotal decimal.Decimal

    for i, split := range splits {
        pct, _ := decimal.NewFromString(*split.Percentage)

        if i == len(splits)-1 {
            // Last split gets remainder
            result[i].Amount = totalDec.Sub(allocatedTotal).String()
        } else {
            amount := totalDec.Mul(pct).Div(decimal.NewFromInt(100)).Round(2)
            result[i].Amount = amount.String()
            allocatedTotal = allocatedTotal.Add(amount)
        }
        result[i].UserID = split.UserID
        result[i].Type = "percentage"
        result[i].Percentage = split.Percentage
    }
    return result, nil
}
```

#### Shares Split Calculation
```go
func (s *expenseService) calculateSharesSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
    totalDec, _ := decimal.NewFromString(total)

    // Calculate total shares
    var totalShares int64
    for _, split := range splits {
        if split.Shares == nil || *split.Shares <= 0 {
            return nil, ErrSharesRequired
        }
        totalShares += int64(*split.Shares)
    }

    // Calculate amounts based on share proportion
    result := make([]calculatedSplit, len(splits))
    var allocatedTotal decimal.Decimal

    for i, split := range splits {
        if i == len(splits)-1 {
            // Last split gets remainder to ensure exact total
            result[i].Amount = totalDec.Sub(allocatedTotal).String()
        } else {
            // amount = total * (shares / totalShares)
            shareRatio := decimal.NewFromInt(int64(*split.Shares)).Div(decimal.NewFromInt(totalShares))
            amount := totalDec.Mul(shareRatio).Round(2)
            result[i].Amount = amount.String()
            allocatedTotal = allocatedTotal.Add(amount)
        }
        result[i].UserID = split.UserID
        result[i].Type = "shares"
        result[i].ShareValue = decimal.NewFromInt(int64(*split.Shares)).String()
    }
    return result, nil
}
```

#### Fixed/Custom Split Validation
```go
func (s *expenseService) validateFixedSplits(total string, splits []SplitInput) ([]calculatedSplit, error) {
    totalDec, _ := decimal.NewFromString(total)

    var splitSum decimal.Decimal
    result := make([]calculatedSplit, len(splits))

    for i, split := range splits {
        if split.Amount == nil {
            return nil, ErrAmountRequired
        }
        amount, _ := decimal.NewFromString(*split.Amount)
        splitSum = splitSum.Add(amount)

        result[i].UserID = split.UserID
        result[i].Amount = *split.Amount
        result[i].Type = split.Type
    }

    if !splitSum.Equal(totalDec) {
        return nil, ErrSplitTotalMismatch
    }

    return result, nil
}
```

### Phase 4: Update CreateExpense Flow

**File:** `be/internal/service/expense_service.go`

In `CreateExpense()`:
1. Replace `validateSplitsTotal()` with `calculateSplitAmounts()`
2. Use calculated amounts when creating splits in database
3. Return calculated amounts in response

### Phase 5: Database Schema Update

Add flexible columns to `expense_split` table to support percentage, shares, and future split types:

**File:** `be/internal/db/migrations/20260201000000_add_share_value_to_expense_split.sql`

```sql
-- Add share_value column for storing percentage or shares
-- Interpretation depends on split_type:
--   percentage: value is percentage (e.g., 60.00 means 60%)
--   shares: value is share count (e.g., 2 means 2 shares)
--   equal/fixed/custom: NULL (not applicable)
ALTER TABLE expense_split
ADD COLUMN share_value DECIMAL(10, 2) CHECK (share_value IS NULL OR share_value >= 0);

-- Add comment for clarity
COMMENT ON COLUMN expense_split.share_value IS
  'For percentage splits: the percentage value (0-100). For shares splits: the share count. NULL for equal/fixed/custom.';
```

**Why this design:**
- Single `share_value` column works for both percentage and shares
- Semantics determined by `split_type` field
- NULL for types that don't need it (equal, fixed, custom)
- Extensible for future split types
- Preserves original input for editing/display

### Phase 6: Update Tests

**File:** `be/internal/service/expense_service_test.go`

Update all expense creation tests to use new request format:
- Remove `amount_owned` from split inputs
- Add `type` field to all splits
- Add `percentage` field for percentage tests
- Add `amount` field for fixed/custom tests

### Phase 7: Update Bruno API Files

Update all Bruno files in `bruno-api/split+/expenses/`:
- `create-expense-equal-split.yml` - Remove amounts, just user_id + type
- `create-expense-percentage-split.yml` - Add percentage field, remove amount
- `create-expense-shares-split.yml` - **NEW**: Add shares field example
- `create-expense-fixed-split.yml` - Keep amount field, remove amount_owned
- `create-expense-custom-split.yml` - Keep amount field, remove amount_owned

## Critical Files to Modify

| File | Changes |
|------|---------|
| `be/internal/http/handlers/expenses.go` | Update SplitRequest struct, add validation |
| `be/internal/service/expense_service.go` | Add calculation logic, update CreateExpense |
| `be/internal/service/expense_service_test.go` | Update all tests for new format |
| `be/internal/db/migrations/20260201000000_add_share_value_to_expense_split.sql` | Add share_value column |
| `be/internal/db/queries/expenses.sql` | Update CreateExpenseSplit to include share_value |
| `be/internal/db/sqlc/` | Regenerate after query changes |
| `bruno-api/split+/expenses/*.yml` | Update all request formats, add shares example |

## New Error Types

```go
var (
    ErrPercentageTotalMismatch = errors.New("percentages must sum to 100")
    ErrAmountRequired          = errors.New("amount required for fixed/custom splits")
    ErrPercentageRequired      = errors.New("percentage required for percentage splits")
    ErrSharesRequired          = errors.New("shares required and must be > 0 for shares splits")
    ErrMixedSplitTypes         = errors.New("all splits must have the same type")
    ErrInvalidSplitType        = errors.New("invalid split type")
)
```

## Validation Rules Summary

**All split types:** All splits in an expense must have the same type (no mixing).

### Equal Split
- At least 1 split required
- Each split must have valid user_id
- No amount, percentage, or shares needed
- Backend calculates equal amounts

### Percentage Split
- At least 1 split required
- Each split must have valid user_id and percentage
- Percentages MUST sum to exactly 100.00
- Backend calculates amounts from percentages

### Shares Split
- At least 1 split required
- Each split must have valid user_id and shares (integer > 0)
- No percentage validation needed (proportional to total shares)
- Backend calculates amounts based on share proportion
- Example: Person A: 2 shares, Person B: 1 share → 66.67% / 33.33%

### Fixed/Custom Split
- At least 1 split required
- Each split must have valid user_id and amount
- Amounts MUST sum to exactly expense total
- Backend uses provided amounts directly

## Response Structure

Response includes backend-calculated `amount_owned` and `share_value` where applicable:

```json
{
  "expense": { ... },
  "payments": [ ... ],
  "splits": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "amount_owned": "40.00",   // Backend calculated
      "split_type": "equal",
      "share_value": null        // null for equal/fixed/custom
    }
  ]
}
```

**share_value interpretation by split_type:**
- `equal`: null (not applicable)
- `percentage`: the percentage value (e.g., "60.00" for 60%)
- `shares`: the share count (e.g., "2" for 2 shares)
- `fixed`/`custom`: null (not applicable)

## Verification Steps

1. Run migration: `just migrate-up`
2. Regenerate sqlc: `just sqlc`
3. Run tests: `just test` (should fail initially)
4. Update tests to new format
5. Run tests again: `just test` (should pass)
6. Start backend: `just be-dev`
7. Test each split type via Bruno:
   - **Equal**: 3-person $100 split → verify $33.33, $33.33, $33.34
   - **Percentage**: 60/40 $1000 split → verify $600, $400
   - **Shares**: 2/1 shares $300 split → verify $200, $100
   - **Fixed**: Specific amounts → verify validation works
   - **Custom**: Same as fixed
8. Test error cases:
   - Percentages not summing to 100 → 422 error
   - Fixed amounts not matching total → 422 error
   - Mixed split types → 422 error
   - Missing required fields → 400 error

## Success Criteria

- [ ] Handler accepts new request format (type, percentage?, shares?, amount?)
- [ ] Service calculates equal splits correctly (handles rounding)
- [ ] Service calculates percentage splits correctly (validates sum to 100%)
- [ ] Service calculates shares splits correctly (proportional)
- [ ] Service validates fixed/custom splits sum to total
- [ ] Database stores share_value for percentage/shares types
- [ ] All existing tests updated and passing
- [ ] Bruno files updated to new format + new shares example
- [ ] Response includes calculated amounts and share_value where applicable
- [ ] Error messages are clear and helpful
- [ ] Mixed split types properly rejected
