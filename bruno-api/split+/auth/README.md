# Authentication Endpoints

This folder contains Bruno API requests for JWT-based authentication.

## Quick Start

1. **Create a user** (if you haven't already):
   - Go to `users/create-user.yml`
   - Run the request
   - Note the user's email and password

2. **Login**:
   - Run `auth/login.yml`
   - Access token and refresh token are automatically saved to environment variables
   - You're now authenticated for all protected endpoints!

3. **Use protected endpoints**:
   - All other API requests will automatically use the saved `access_token`
   - No manual token management needed

## Endpoints

### 1. Login

**File**: `login.yml`  
**Method**: `POST /auth/login`

Authenticates with email/password and returns JWT tokens.

**Auto-saves**:

- `access_token` → Used for all authenticated requests
- `refresh_token` → Used to get new access tokens

**Token Lifetimes**:

- Access token: 7 days
- Refresh token: 30 days

---

### 2. Refresh Token

**File**: `refresh-token.yml`  
**Method**: `POST /auth/refresh`

Gets a new access token using your refresh token.

**Use when**:

- Access token expires
- You want to extend your session

**Auto-updates**: `access_token`

---

### 3. Logout

**File**: `logout.yml`  
**Method**: `POST /auth/logout`

Logs out from current device/session.

**What happens**:

- Invalidates the refresh token
- Blacklists the access token (immediate revocation)
- Clears tokens from environment

---

### 4. Logout All

**File**: `logout-all.yml`  
**Method**: `POST /auth/logout-all`

Logs out from ALL devices/sessions.

**Use when**:

- Security concern (unauthorized access)
- Want to force logout everywhere
- After password change

**What happens**:

- Invalidates all refresh tokens
- User logged out from all devices
- Clears tokens from environment

---

## Environment Variables

The following variables are used by auth endpoints:

| Variable        | Description        | Auto-set?              |
| --------------- | ------------------ | ---------------------- |
| `user_email`    | Email for login    | No - set manually      |
| `user_password` | Password for login | No - set manually      |
| `access_token`  | JWT access token   | Yes - by login/refresh |
| `refresh_token` | JWT refresh token  | Yes - by login         |

## Authentication Flow

```
1. Login
   ↓
2. Get access_token + refresh_token (auto-saved)
   ↓
3. Use access_token for API requests (automatic)
   ↓
4. When access_token expires (7 days):
   → Refresh token to get new access_token
   ↓
5. When done:
   → Logout (single device) or Logout All (all devices)
```

## Testing Tips

1. **First time setup**:

   ```
   1. Create user (users/create-user.yml)
   2. Update user_email and user_password in environment
   3. Run login
   ```

2. **Daily usage**:

   ```
   1. Run login once
   2. All other requests work automatically
   ```

3. **Token expiration testing**:

   ```
   1. Login
   2. Wait 7 days (or manually expire token in DB)
   3. Try protected endpoint → should fail
   4. Run refresh-token
   5. Try protected endpoint → should work
   ```

4. **Logout testing**:
   ```
   1. Login
   2. Try protected endpoint → works
   3. Logout
   4. Try protected endpoint → fails (401)
   ```

## Response Format

All auth endpoints return:

**Success**:

```json
{
  "status": true,
  "data": {
    // endpoint-specific data
  }
}
```

**Error**:

```json
{
  "status": false,
  "error": {
    "message": "error description"
  }
}
```

## Common Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Invalid credentials or token
- `500 Internal Server Error` - Server error

## Notes

- Tokens are automatically managed by Bruno scripts
- Access tokens expire in 7 days
- Refresh tokens expire in 30 days
- Blacklisted tokens are cleaned up hourly by background worker
- All timestamps are in UTC
