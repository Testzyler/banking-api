# Authentication Flow - Banking API

### 🔐 Authentication Scenarios

#### ✅ **Success Flows**
1. **Login Success**
   - User provides correct username + PIN
   - System generates JWT access token (15 min) + refresh token (24 hours)
   - Tokens stored securely for session management
   - User gains access to protected resources

2. **Token Refresh Success**
   - User presents valid refresh token
   - System generates new access token
   - Session continues without re-login

3. **Protected Access Success**
   - User sends valid access token in Authorization header
   - System validates token (not expired, not banned)
   - Request processed and data returned

#### ❌ **Failure Flows**
1. **Login Failures**
   - **Wrong PIN (< 3 attempts)**: Returns `401 Unauthorized`
   - **Wrong PIN (≥ 3 attempts)**: PIN locked for 30 minutes, returns `423 Locked`
   - **Account not found**: Returns `401 Unauthorized`

2. **Token Failures**
   - **Expired Token**: Returns `401 Unauthorized` - client should use refresh token
   - **Banned Token**: Returns `401 Unauthorized` - client must re-login
   - **Invalid Token**: Returns `401 Unauthorized` - client must re-login

3. **Security Actions**
   - **Suspicious Activity**: All user tokens banned, requires fresh login
   - **Token Versioning**: Old tokens invalidated when security event occurs

### 🛡️ Security Features

- **PIN Lockout**: 3 failed attempts = lock
- **Token Expiry**: Short-lived access tokens (15 min) for security
- **Token Banning**: Immediate token invalidation capability
- **Version Control**: Token versioning prevents replay attacks
- **Exponential Backoff Retry**: Protection against brute force attacks


## Authentication Overview Diagram
![Authentication-Flow](auth-sequendiagram.png)