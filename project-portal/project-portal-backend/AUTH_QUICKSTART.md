# Auth Service Quick Start Guide

## 📦 Installation

### 1. Add the Stellar SDK dependency
```bash
cd project-portal/project-portal-backend
go get github.com/stellar/go/...
```

### 2. Configure Environment
Create `.env` file in `project-portal-backend/`:

```env
# Required
DATABASE_URL=postgres://user:password@localhost:5432/carbon_scribe_portal
PORT=8080

# Auth Configuration
JWT_SECRET=super-secret-key-that-is-at-least-32-characters-long
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d
PASSWORD_HASH_COST=12

# Email & URLs
EMAIL_VERIFICATION_URL=https://yourapp.com/verify-email
PASSWORD_RESET_URL=https://yourapp.com/reset-password

# Stellar Wallet (use Test Network for development)
STELLAR_NETWORK_PASSPHRASE=Test SDF Network ; September 2015

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# CORS
CORS_ALLOWED_ORIGINS=*

# Database
DEBUG=true
SERVER_MODE=development

# Other services (elasticsearch, AWS, etc.)
ELASTICSEARCH_ADDRESSES=http://localhost:9200
AWS_REGION=us-east-1
```

### 3. Start Services
```bash
# PostgreSQL
docker run --name postgres -e POSTGRES_PASSWORD=yourpassword \
  -e POSTGRES_DB=carbon_scribe_portal -p 5432:5432 -d postgres:15

# Redis (optional, for production)
docker run --name redis -p 6379:6379 -d redis:7

# Elasticsearch (optional)
docker run --name elasticsearch -e discovery.type=single-node \
  -e xpack.security.enabled=false -p 9200:9200 -d docker.elastic.co/elasticsearch/elasticsearch:8.0.0
```

### 4. Run Server
```bash
cd project-portal/project-portal-backend
go run cmd/api/main.go
```

Expected output:
```
✅ Database connection established
✅ Elasticsearch client initialized
🚀 Server starting on port 8080
📡 Listening on http://localhost:8080
📊 Health check: http://localhost:8080/health
🔗 Available endpoints:
   - Authentication: /api/v1/auth/*
```

## 🧪 Quick Test Examples

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!",
    "full_name": "Test User",
    "organization": "Test Org"
  }'
```

Response:
```json
{
  "user": {
    "id": "abc123...",
    "email": "test@example.com",
    "full_name": "Test User",
    "role": "farmer",
    "email_verified": false,
    "created_at": "2024-03-20T10:30:00Z"
  },
  "verification_token": "abc123xyz...",
  "message": "User registered successfully. Please verify your email."
}
```

Save the `verification_token` for email verification.

### 3. Verify Email
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{"token": "YOUR_VERIFICATION_TOKEN"}'
```

### 4. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!"
  }'
```

Response:
```json
{
  "user": {...},
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

Save the `access_token` for authenticated requests. The token is valid for 15 minutes.

### 5. Get User Profile (Protected)
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 6. Refresh Token (Before Expiry)
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

### 7. Change Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "TestPassword123!",
    "new_password": "NewPassword456!"
  }'
```

### 8. Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🔐 Wallet Authentication (Stellar)

### 1. Get Challenge
```bash
curl -X POST http://localhost:8080/api/v1/auth/wallet-challenge \
  -H "Content-Type: application/json" \
  -d '{"public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"}'
```

Response:
```json
{
  "challenge": "abc123xyz...",
  "expires_in": 900
}
```

### 2. Sign Challenge
Sign the challenge with your Stellar wallet (using StellarChain, Albedo, etc.)

### 3. Login with Wallet
```bash
curl -X POST http://localhost:8080/api/v1/auth/wallet-login \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "GXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "signed_challenge": "signed_transaction_envelope_xdr..."
  }'
```

## 🛡️ Protected Routes Usage

For all protected endpoints, include the Authorization header:

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 📋 Role-Based Access Examples

The RBAC system will automatically enforce permissions based on user roles:

```go
// In route setup
protected.POST("/sensitive-operation", 
  RequirePermission("admin:manage_users"),
  handler.ManageUsers)

// Or by role
protected.DELETE("/user/:id",
  RequireRole("admin"),
  handler.DeleteUser)
```

Available roles:
- `farmer` - Default role for regular users
- `verifier` - Can verify projects and documents
- `admin` - Full system administration
- `viewer` - Read-only access

## 🐛 Common Issues & Fixes

### 1. "database connection failed"
- Ensure PostgreSQL is running and DATABASE_URL is correct
- Check: `psql $DATABASE_URL -c "SELECT 1"`

### 2. "invalid token"
- Token may have expired (access tokens expire in 15 minutes by default)
- Try refreshing: use the refresh_token endpoint
- Or re-login to get a new access token

### 3. "insufficient permissions"
- User role doesn't have required permission
- Check role permissions in `role_permissions` table
- Admin users have all permissions

### 4. "Invalid Stellar wallet address"
- Ensure the public key is in valid Stellar format (starts with 'G')
- Only 56-character Stellar addresses are supported

## 📚 Database Schema

Key tables:
- `users` - User accounts
- `user_sessions` - Active sessions
- `user_wallets` - Associated wallets
- `auth_tokens` - Email verification and password reset tokens
- `role_permissions` - RBAC configuration

## 🔄 Token Lifecycle

```
1. User Login/Register
   ↓
2. Access Token (15 min) + Refresh Token (7 days)
   ↓
3. Use Access Token for API calls
   ↓
4. Token expires → Use Refresh Token
   ↓
5. Get new Access Token
   ↓
6. Continue using new token
   ↓
7. Logout → Tokens invalidated
```

## 🚀 Production Checklist

- [ ] Change JWT_SECRET to a strong random value
- [ ] Set appropriate token expiry times
- [ ] Enable HTTPS in production
- [ ] Setup email service for verification/reset tokens
- [ ] Configure Redis for session management
- [ ] Enable rate limiting
- [ ] Setup monitoring and alerting
- [ ] Configure database backups
- [ ] Enable audit logging
- [ ] Setup CORS for your domain
- [ ] Use Stellar Public Network (update STELLAR_NETWORK_PASSPHRASE)
- [ ] Implement 2FA support

## 📞 Integration Points

Other services can use auth:
```go
import "carbon-scribe/project-portal/project-portal-backend/internal/auth"

// In your handler
authMiddleware := auth.AuthMiddleware(tokenManager)
protected.GET("/resource", authMiddleware, handler.GetResource)

// Check permissions
permOracle := auth.RequirePermission("resource:read")
protected.GET("/protected", authMiddleware, permOracle, handler.Protected)
```

---

For more details, see [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md)
