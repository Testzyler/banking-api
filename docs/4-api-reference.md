
# Banking API Reference

**Base URL**: `http://localhost:8080/api/v1`

## Authentication Endpoints

### Verify PIN

```http
POST /api/v1/auth/verify-pin
```

Authenticate user with PIN and receive JWT tokens.

| Parameter  | Type     | Description                     |
| :--------  | :------- | :------------------------------ |
| `username` | `string` | **Required**. User identifier   |
| `pin`      | `string` | **Required**. 6-digit PIN code  |

**Request Body:**
```json
{
  "username": "user123",
  "pin": "123456"
}
```

**Response:**
```json
{
  "code": 10200,
  "message": "Authentication successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiry": "2025-08-01T05:44:00Z",
    "userID": "user123",
    "tokenVersion": 1725175440,
    "tokenID": "token_uuid"
  }
}
```

### Refresh Token

```http
POST /api/v1/auth/refresh
```

Get new access token using refresh token.

| Parameter      | Type     | Description                        |
| :------------- | :------- | :--------------------------------- |
| `refreshToken` | `string` | **Required**. Valid refresh token  |

**Request Body:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:**
```json
{
  "code": 10200,
  "message": "Token refreshed successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiry": "2025-08-01T06:44:00Z",
    "userID": "user123",
    "tokenVersion": 1725179040,
    "tokenID": "new_token_uuid"
  }
}
```

### List All Tokens

```http
GET /api/v1/auth/tokens
```

Get list of all active tokens for debugging/monitoring.

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:**
```json
{
  "code": 10200,
  "message": "Tokens retrieved successfully",
  "data": [
    {
      "token": "eyJhbGciOiJIUzI1NiIs...",
      "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
      "expiry": "2025-08-01T05:44:00Z",
      "userID": "user123",
      "tokenVersion": 1725175440,
      "tokenID": "token_uuid"
    }
  ]
}
```

### Ban All User Tokens

```http
POST /api/v1/auth/ban-tokens
```

Invalidate all tokens for a user (security action).

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:**
```json
{
  "code": 10200,
  "message": "All user tokens banned successfully",
  "data": null
}
```

## Protected Endpoints

### Get Home Data

```http
GET /api/v1/home
```

Retrieve user's complete banking dashboard including accounts, balances, cards and transactions.

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response:**
```json
{
  "code": 10200,
  "message": "Home data retrieved successfully",
  "data": {
    "userID": "user123",
    "name": "John Doe",
    "totalBalance": 25847.50,
    "accounts": [...],
    "debitCards": [...],
    "transactions": [...],
    "banners": [...]
  }
}
```

## Health Check

### Application Health

```http
GET /healthz
```

Check if the application is running and healthy.

**Response:**
```json
{
  "code": 10200,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2025-08-01T10:30:00Z"
  }
}
```

## Error Responses

### Standard Error Format

All API endpoints return errors in a consistent format:

```json
{
  "code": 10401,
  "message": "Authentication failed",
  "error": "Invalid PIN provided",
  "timestamp": "2025-08-01T10:30:00Z"
}
```

### Response Codes

| Code  | Status | Description           |
| :---- | :----- | :-------------------- |
| 10200 | 200    | Success               |
| 10400 | 400    | Bad Request           |
| 10401 | 401    | Unauthorized          |
| 10404 | 404    | Not Found             |
| 10422 | 422    | Validation Failed     |
| 10423 | 423    | Account Locked        |
| 10500 | 500    | Internal Server Error |
