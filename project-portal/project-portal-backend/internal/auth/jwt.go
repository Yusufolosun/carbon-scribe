package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenManager handles JWT token operations
type TokenManager struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	BlacklistedTokens  map[string]time.Time // In production, use Redis
}

// NewTokenManager creates a new token manager
func NewTokenManager(secret string, accessExpiry, refreshExpiry time.Duration) *TokenManager {
	return &TokenManager{
		Secret:             secret,
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
		BlacklistedTokens:  make(map[string]time.Time),
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (tm *TokenManager) GenerateTokenPair(user *User, permissions []string) (*TokenResponse, error) {
	accessToken, err := tm.GenerateAccessToken(user, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := tm.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Return both tokens along with metadata
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(tm.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken generates an access token for a user
func (tm *TokenManager) GenerateAccessToken(user *User, permissions []string) (string, error) {
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"permissions": permissions,
		"token_type":  "access",
		"jti":         jti,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(tm.AccessTokenExpiry).Unix(),
	}

	if user.WalletAddress != "" {
		claims["wallet_address"] = user.WalletAddress
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(tm.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenStr, nil
}

// GenerateRefreshToken generates a refresh token for a user
func (tm *TokenManager) GenerateRefreshToken(user *User) (string, error) {
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"token_type": "refresh",
		"jti":        jti,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(tm.RefreshTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(tm.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenStr, nil
}

// ValidateToken validates a JWT token and returns the claims
func (tm *TokenManager) ValidateToken(tokenStr string, expectedType string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is blacklisted
	if jti, ok := claims["jti"].(string); ok && jti != "" {
		if expTime, blacklisted := tm.BlacklistedTokens[jti]; blacklisted {
			if time.Now().Before(expTime) {
				return nil, errors.New("token has been revoked")
			}
		}
	}

	// Verify token type matches expected type
	if expectedType != "" {
		tokenType, ok := claims["token_type"].(string)
		if !ok || tokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, tokenType)
		}
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (tm *TokenManager) ValidateAccessToken(tokenStr string) (jwt.MapClaims, error) {
	return tm.ValidateToken(tokenStr, "access")
}

// ValidateRefreshToken validates a refresh token
func (tm *TokenManager) ValidateRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	return tm.ValidateToken(tokenStr, "refresh")
}

// RevokeToken adds a token to the blacklist
func (tm *TokenManager) RevokeToken(tokenStr string) error {
	claims, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tm.Secret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	if mapClaims, ok := claims.Claims.(jwt.MapClaims); ok {
		if jti, ok := mapClaims["jti"].(string); ok && jti != "" {
			if expTime, ok := mapClaims["exp"].(float64); ok {
				tm.BlacklistedTokens[jti] = time.Unix(int64(expTime), 0)
				return nil
			}
		}
	}

	return errors.New("unable to revoke token: missing JTI or expiration")
}

// GenerateChallengeToken generates a challenge token for Stellar wallet authentication
func GenerateChallengeToken(walletAddress string, timeout time.Duration) (string, error) {
	// Generate a random challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return "", fmt.Errorf("failed to generate random challenge: %w", err)
	}

	// Encode challenge as hex
	challengeStr := hex.EncodeToString(challenge)

	// In a real implementation, you would store this in Redis with a TTL
	// For now, we're just returning it
	return challengeStr, nil
}
