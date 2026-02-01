package repository

import (
	"context"

	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, params sqlc.CreateSessionParams) (sqlc.Session, error)
	GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (sqlc.Session, error)
	UpdateSessionLastUsed(ctx context.Context, sessionID pgtype.UUID) error
	DeleteSession(ctx context.Context, refreshTokenHash string) error
	DeleteAllUserSessions(ctx context.Context, userID pgtype.UUID) error
	DeleteExpiredSessions(ctx context.Context) error
	GetActiveSessionsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.Session, error)

	// Token blacklisting
	BlacklistToken(ctx context.Context, params sqlc.BlacklistTokenParams) error
	IsTokenBlacklisted(ctx context.Context, tokenJTI string) (bool, error)
	DeleteExpiredBlacklistedTokens(ctx context.Context) error
}

type sessionRepository struct {
	queries *sqlc.Queries
}

func NewSessionRepository(queries *sqlc.Queries) SessionRepository {
	return &sessionRepository{queries: queries}
}

func (r *sessionRepository) CreateSession(ctx context.Context, params sqlc.CreateSessionParams) (sqlc.Session, error) {
	return r.queries.CreateSession(ctx, params)
}

func (r *sessionRepository) GetSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (sqlc.Session, error) {
	return r.queries.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
}

func (r *sessionRepository) UpdateSessionLastUsed(ctx context.Context, sessionID pgtype.UUID) error {
	return r.queries.UpdateSessionLastUsed(ctx, sessionID)
}

func (r *sessionRepository) DeleteSession(ctx context.Context, refreshTokenHash string) error {
	return r.queries.DeleteSession(ctx, refreshTokenHash)
}

func (r *sessionRepository) DeleteAllUserSessions(ctx context.Context, userID pgtype.UUID) error {
	return r.queries.DeleteAllUserSessions(ctx, userID)
}

func (r *sessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.queries.DeleteExpiredSessions(ctx)
}

func (r *sessionRepository) GetActiveSessionsByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.Session, error) {
	return r.queries.GetActiveSessionsByUserID(ctx, userID)
}

func (r *sessionRepository) BlacklistToken(ctx context.Context, params sqlc.BlacklistTokenParams) error {
	return r.queries.BlacklistToken(ctx, params)
}

func (r *sessionRepository) IsTokenBlacklisted(ctx context.Context, tokenJTI string) (bool, error) {
	return r.queries.IsTokenBlacklisted(ctx, tokenJTI)
}

func (r *sessionRepository) DeleteExpiredBlacklistedTokens(ctx context.Context) error {
	return r.queries.DeleteExpiredBlacklistedTokens(ctx)
}
