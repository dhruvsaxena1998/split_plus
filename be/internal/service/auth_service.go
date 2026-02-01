package service

import (
	"context"
	"errors"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionNotFound     = errors.New("session not found")
)

type AuthService interface {
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (accessToken, refreshToken string, expiresIn int64, err error)
	RefreshToken(ctx context.Context, refreshToken, userAgent, ipAddress string) (accessToken string, expiresIn int64, err error)
	Logout(ctx context.Context, refreshToken, accessTokenJTI string, userID pgtype.UUID) error
	LogoutAllSessions(ctx context.Context, userID pgtype.UUID) error
	RevokeToken(ctx context.Context, tokenJTI string, userID pgtype.UUID, reason string) error
	CleanupExpiredSessions(ctx context.Context) error
}

type authService struct {
	userService        UserService
	sessionRepository  repository.SessionRepository
	jwtService         JWTService
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewAuthService(
	userService UserService,
	sessionRepository repository.SessionRepository,
	jwtService JWTService,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) AuthService {
	return &authService{
		userService:        userService,
		sessionRepository:  sessionRepository,
		jwtService:         jwtService,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (s *authService) Login(
	ctx context.Context,
	email, password, userAgent, ipAddress string,
) (string, string, int64, error) {
	// Authenticate user
	user, err := s.userService.AuthenticateUser(ctx, email, password)
	if err != nil {
		return "", "", 0, err
	}

	// Generate access token
	// Convert pgtype.UUID to uuid.UUID for string conversion
	userUUID, err := uuid.FromBytes(user.ID.Bytes[:])
	if err != nil {
		return "", "", 0, err
	}
	accessToken, _, err := s.jwtService.GenerateAccessToken(userUUID.String(), user.Email)
	if err != nil {
		return "", "", 0, err
	}

	// Generate refresh token (cryptographically secure random string)
	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return "", "", 0, err
	}

	// Store session with refresh token
	// Note: Refresh token is 32 random bytes base64 encoded, secure enough to store directly
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(s.refreshTokenExpiry),
		Valid: true,
	}

	_, err = s.sessionRepository.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: refreshToken, // Storing directly for efficient lookup
		UserAgent:        pgtype.Text{String: userAgent, Valid: userAgent != ""},
		IpAddress:        pgtype.Text{String: ipAddress, Valid: ipAddress != ""},
		ExpiresAt:        expiresAt,
	})
	if err != nil {
		return "", "", 0, err
	}

	expiresIn := int64(s.accessTokenExpiry.Seconds())
	return accessToken, refreshToken, expiresIn, nil
}

func (s *authService) RefreshToken(
	ctx context.Context,
	refreshToken, userAgent, ipAddress string,
) (string, int64, error) {
	// Look up session by refresh token
	session, err := s.sessionRepository.GetSessionByRefreshTokenHash(ctx, refreshToken)
	if err != nil {
		return "", 0, ErrSessionNotFound
	}

	// Update last used timestamp
	_ = s.sessionRepository.UpdateSessionLastUsed(ctx, session.ID)

	// Generate new access token
	// Convert pgtype.UUID to uuid.UUID for string conversion
	sessionUserUUID, err := uuid.FromBytes(session.UserID.Bytes[:])
	if err != nil {
		return "", 0, err
	}
	accessToken, _, err := s.jwtService.GenerateAccessToken(sessionUserUUID.String(), "")
	if err != nil {
		return "", 0, err
	}

	expiresIn := int64(s.accessTokenExpiry.Seconds())
	return accessToken, expiresIn, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken, accessTokenJTI string, userID pgtype.UUID) error {
	// Delete session by refresh token
	if err := s.sessionRepository.DeleteSession(ctx, refreshToken); err != nil {
		return err
	}

	// Blacklist access token for immediate revocation
	if accessTokenJTI != "" {
		expiresAt := pgtype.Timestamptz{
			Time:  time.Now().Add(s.accessTokenExpiry),
			Valid: true,
		}

		err := s.sessionRepository.BlacklistToken(ctx, sqlc.BlacklistTokenParams{
			TokenJti:  accessTokenJTI,
			UserID:    userID,
			ExpiresAt: expiresAt,
			Reason:    pgtype.Text{String: "user logout", Valid: true},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) LogoutAllSessions(ctx context.Context, userID pgtype.UUID) error {
	// Delete all user sessions
	// Note: This doesn't blacklist active access tokens since we don't track JTIs per session
	// Access tokens will naturally expire after their TTL
	return s.sessionRepository.DeleteAllUserSessions(ctx, userID)
}

func (s *authService) RevokeToken(ctx context.Context, tokenJTI string, userID pgtype.UUID, reason string) error {
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(s.accessTokenExpiry),
		Valid: true,
	}

	return s.sessionRepository.BlacklistToken(ctx, sqlc.BlacklistTokenParams{
		TokenJti:  tokenJTI,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Reason:    pgtype.Text{String: reason, Valid: reason != ""},
	})
}

func (s *authService) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.sessionRepository.DeleteExpiredSessions(ctx); err != nil {
		return err
	}
	return s.sessionRepository.DeleteExpiredBlacklistedTokens(ctx)
}
