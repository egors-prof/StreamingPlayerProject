// Package dbstore — реализация портов хранения на Postgres.
package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"github.com/egors-prof/streamingplayer/pkg/pgdb"
	"github.com/jmoiron/sqlx"
)

type UserStorage struct {
	db *sqlx.DB
}

func NewUserStorage(db *sqlx.DB) *UserStorage { return &UserStorage{db: db} }

// userRow — представление строки таблицы. Отделено от domain.User, чтобы
// теги БД не протекали в бизнес-логику.
type userRow struct {
	ID           int       `db:"id"`
	FullName     string    `db:"full_name"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (r userRow) toDomain() domain.User {
	return domain.User{
		ID: r.ID, FullName: r.FullName, Username: r.Username,
		PasswordHash: r.PasswordHash, Role: domain.Role(r.Role),
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

const userColumns = `id, full_name, username, password_hash, role, created_at, updated_at`

func (s *UserStorage) CreateUser(ctx context.Context, u domain.User) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (full_name, username, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		u.FullName, u.Username, u.PasswordHash, string(u.Role)).Scan(&id)
	if pgdb.IsUniqueViolation(err) {
		return 0, errs.ErrUsernameTaken
	}
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (s *UserStorage) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	return s.getUser(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username)
}

func (s *UserStorage) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	return s.getUser(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
}

func (s *UserStorage) getUser(ctx context.Context, query string, arg any) (domain.User, error) {
	var row userRow
	err := s.db.GetContext(ctx, &row, query, arg)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, errs.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("select user: %w", err)
	}
	return row.toDomain(), nil
}
