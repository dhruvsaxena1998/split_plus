# Frontend Integration Guide

This guide provides a comprehensive reference for integrating the frontend with the Split+ backend API. All API endpoints are fully documented in the `bruno-api` collection, which serves as the source of truth for request/response formats.

## 📚 API Documentation (Bruno)
We use [Bruno](https://www.usebruno.com/) for API documentation and testing. The collection is located in `bruno-api/split+`.

**Key Features of our Bruno Collection:**
1. **Auto-managed Authentication**: Login once, and tokens are automatically saved to environment variables (`access_token`, `refresh_token`).
2. **Comprehensive Examples**: Every endpoint has a working example with documentation.
3. **organized Structure**: Grouped by domain (Auth, Groups, Expenses, etc.).

## 🔐 Authentication & Security
**Bruno Reference**: `bruno-api/split+/auth`

### Token Strategy
- **Access Token**: Short-lived (7 days), used for all API requests.
- **Refresh Token**: Long-lived (30 days), used to get new access tokens.
- **Storage**: Store tokens in secure storage (e.g., `SecureStore` in mobile, `HttpOnly` cookie or memory in web).

### Auth Flows

1. **Login**
   - **Endpoint**: `POST /auth/login` (`auth/login.yml`)
   - **Action**: Save `access_token` and `refresh_token`.
   
2. **Token Refresh** (Automatic)
   - **Endpoint**: `POST /auth/refresh` (`auth/refresh-token.yml`)
   - **Trigger**: When any API returns `401 Unauthorized` with `token_expired` error.
   - **Action**: Call refresh endpoint with `refresh_token`. If successful, retry original request with new `access_token`. If fails, log user out.

3. **Logout**
   - **Endpoint**: `POST /auth/logout` (`auth/logout.yml`)
   - **Action**: Call API to blacklist token, then clear local tokens.

## 👥 User & Friends
**Bruno Reference**: `bruno-api/split+/friends`

- **List Friends**: `GET /friends` (`friends/list-friends.yml`)
- **Add Friend**: `POST /friends/requests` (`friends/send-friend-request.yml`)
- **Accept/Decline**: `POST /friends/requests/{id}/accept|decline`

## 🏘️ Groups & Members
**Bruno Reference**: `bruno-api/split+/groups`

- **Create Group**: `POST /groups` (`groups/create-group.yml`)
- **List User Groups**: `GET /groups` (`groups/list-user-groups.yml`)
- **Group Details**: `GET /groups/{group_id}` (`groups/get-group.yml`)

### Invitations
**Bruno Reference**: `bruno-api/split+/invitations`
- **Generate Link**: `POST /groups/{group_id}/invitations`
- **View Invite**: `GET /invitations/{token}` (Public endpoint)
- **Accept Invite**: `POST /invitations/{token}/accept` (Authenticated)

## 💸 Expenses
**Bruno Reference**: `bruno-api/split+/expenses`

### Creating an Expense
**File**: `expenses/create-expense.yml`
- **Endpoint**: `POST /groups/{group_id}/expenses`
- **Split Types**:
  - `EQUAL`: Splits equally among participants.
  - `EXACT`: Specify exact amount for each person.
  - `PERCENTAGE`: Specify percentage share (must sum to 100).
  - `SHARES`: Proportional split (e.g., 2 shares vs 1 share).

### Categories
**Bruno Reference**: `bruno-api/split+/categories`
- **Presets**: `GET /categories/presets` (Public) - Use for onboarding/creation.
- **Group Categories**: `GET /groups/{group_id}/categories` - Custom categories for the group.

### Recurring Expenses
**Bruno Reference**: `bruno-api/split+/recurring-expenses`
- **Create**: `POST /groups/{group_id}/recurring-expenses`
- **Frequency**: `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY`.

## 💰 Balances & Settlements
**Bruno Reference**: `bruno-api/split+/balances` & `bruno-api/split+/settlements`

- **Group Balances**: `GET /groups/{group_id}/balances` - Shows who owes whom.
- **Settle Up**: `POST /groups/{group_id}/settlements` - Record a payment.

## 💬 Social Features
**Bruno Reference**: `bruno-api/split+/comments` & `bruno-api/split+/activities`

- **Activity Feed**: `GET /groups/{group_id}/activity`
- **Expense Comments**: `GET /groups/{group_id}/expenses/{id}/comments`

## 🛠️ Error Handling

API errors follow a standard format:
```json
{
  "status": false,
  "error": {
    "message": "Human readable error",
    "code": "ERROR_CODE_STRING" // Optional code for programmatic handling
  }
}
```

**Common Status Codes**:
- `401 Unauthorized`: Token missing, invalid, or expired.
- `403 Forbidden`: User does not have permission (e.g., not in group).
- `422 Unprocessable Entity`: Validation failed (e.g., missing field).
- `409 Conflict`: Resource already exists (e.g., duplicate email).

## 🚀 Integration Checklist

1. [ ] **Setup AuthContext**: Implement Login/Logout/Refresh logic.
2. [ ] **Axios Interceptor**: Handle token attachment and 401 refresh flow.
3. [ ] **Type Generation**: (Optional) Generate TS types from API responses.
4. [ ] **Error Boundaries**: Handle API errors gracefully in UI.
