package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"github.com/jmoiron/sqlx"
)

type TokenStorage struct {
	db *sqlx.DB
}

func NewTokenStorage(db *sqlx.DB) *TokenStorage { return &TokenStorage{db: db} }

func (s *TokenStorage) SaveRefreshToken(ctx context.Context, t domain.RefreshToken) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		t.ID, t.UserID, t.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// RotateRefreshToken выполняется в транзакции. SELECT ... FOR UPDATE блокирует
// строку старого токена: если два запроса придут с одним токеном одновременно,
// второй дождётся первого и увидит, что токен уже отозван.
func (s *TokenStorage) RotateRefreshToken(ctx context.Context, oldID string, next domain.RefreshToken) (err error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var revokedAt sql.NullTime
	err = tx.GetContext(ctx, &revokedAt,
		`SELECT revoked_at FROM refresh_tokens WHERE id = $1 FOR UPDATE`, oldID)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("select refresh token: %w", err)
	}
	if revokedAt.Valid {
		return errs.ErrTokenReused
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1`, oldID); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		next.ID, next.UserID, next.ExpiresAt); err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *TokenStorage) RevokeRefreshToken(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (s *TokenStorage) RevokeAllUserTokens(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("revoke user tokens: %w", err)
	}
	return nil
}
