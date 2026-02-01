package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-at-least-32-chars", 15*time.Minute, 24*time.Hour)

	t.Run("generates valid access token", func(t *testing.T) {
		userID := "123e4567-e89b-12d3-a456-426614174000"
		email := "test@example.com"

		token, jti, err := jwtService.GenerateAccessToken(userID, email)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, jti)
	})

	t.Run("generates unique JTI for each token", func(t *testing.T) {
		userID := "123e4567-e89b-12d3-a456-426614174000"
		email := "test@example.com"

		_, jti1, err := jwtService.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		_, jti2, err := jwtService.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		assert.NotEqual(t, jti1, jti2, "JTI should be unique for each token")
	})
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-at-least-32-chars", 15*time.Minute, 24*time.Hour)

	t.Run("validates valid token", func(t *testing.T) {
		userID := "123e4567-e89b-12d3-a456-426614174000"
		email := "test@example.com"

		token, jti, err := jwtService.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		claims, err := jwtService.ValidateAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, jti, claims.ID)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		invalidToken := "invalid.token.here"

		_, err := jwtService.ValidateAccessToken(invalidToken)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("rejects token with wrong signature", func(t *testing.T) {
		// Create token with different secret
		otherService := NewJWTService("different-secret-key-32-chars-long", 15*time.Minute, 24*time.Hour)
		token, _, err := otherService.GenerateAccessToken("123", "test@example.com")
		require.NoError(t, err)

		// Try to validate with original service
		_, err = jwtService.ValidateAccessToken(token)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		// Create service with very short expiry
		shortLivedService := NewJWTService("test-secret-key-at-least-32-chars", 1*time.Millisecond, 24*time.Hour)
		token, _, err := shortLivedService.GenerateAccessToken("123", "test@example.com")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortLivedService.ValidateAccessToken(token)
		assert.Error(t, err)
		assert.Equal(t, ErrExpiredToken, err)
	})
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-at-least-32-chars", 15*time.Minute, 24*time.Hour)

	t.Run("generates valid refresh token", func(t *testing.T) {
		token, err := jwtService.GenerateRefreshToken()

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, len(token), 40, "Refresh token should be reasonably long")
	})

	t.Run("generates unique refresh tokens", func(t *testing.T) {
		token1, err := jwtService.GenerateRefreshToken()
		require.NoError(t, err)

		token2, err := jwtService.GenerateRefreshToken()
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2, "Refresh tokens should be unique")
	})
}

func TestJWTService_ParseAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-at-least-32-chars", 15*time.Minute, 24*time.Hour)

	t.Run("parses token without validation", func(t *testing.T) {
		userID := "123e4567-e89b-12d3-a456-426614174000"
		email := "test@example.com"

		token, jti, err := jwtService.GenerateAccessToken(userID, email)
		require.NoError(t, err)

		claims, err := jwtService.ParseAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, jti, claims.ID)
	})

	t.Run("parses expired token without error", func(t *testing.T) {
		// Create service with very short expiry
		shortLivedService := NewJWTService("test-secret-key-at-least-32-chars", 1*time.Millisecond, 24*time.Hour)
		token, _, err := shortLivedService.GenerateAccessToken("123", "test@example.com")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		// Parse should succeed even though token is expired
		claims, err := jwtService.ParseAccessToken(token)
		require.NoError(t, err)
		assert.Equal(t, "123", claims.UserID)
	})
}
