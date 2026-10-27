package session

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"github.com/egors-prof/streamingplayer/pkg/authjwt"
	"golang.org/x/crypto/bcrypt"
)

// Фейковые хранилища в памяти: гексагональная архитектура позволяет
// тестировать бизнес-логику без базы данных.
type fakeUsers struct{ byName map[string]domain.User }

func (f *fakeUsers) CreateUser(context.Context, domain.User) (int, error) { return 0, nil }
func (f *fakeUsers) GetUserByUsername(_ context.Context, name string) (domain.User, error) {
	u, ok := f.byName[name]
	if !ok {
		return domain.User{}, errs.ErrNotFound
	}
	return u, nil
}
func (f *fakeUsers) GetUserByID(_ context.Context, id int) (domain.User, error) {
	for _, u := range f.byName {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, errs.ErrNotFound
}

type tokenRow struct {
	userID    int
	revoked   bool
	rotatedAt time.Time // заполняется только при ротации
}

type fakeTokens struct {
	mu   sync.Mutex
	rows map[string]*tokenRow
}

func (f *fakeTokens) SaveRefreshToken(_ context.Context, t domain.RefreshToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows[t.ID] = &tokenRow{userID: t.UserID}
	return nil
}
func (f *fakeTokens) RotateRefreshToken(_ context.Context, oldID string, next domain.RefreshToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	row, ok := f.rows[oldID]
	if !ok {
		return errs.ErrNotFound
	}
	if row.revoked {
		return errs.ErrTokenReused
	}
	row.revoked, row.rotatedAt = true, time.Now()
	f.rows[next.ID] = &tokenRow{userID: next.UserID}
	return nil
}
func (f *fakeTokens) RefreshTokenRotatedAt(_ context.Context, id string) (time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	row, ok := f.rows[id]
	if !ok {
		return time.Time{}, errs.ErrNotFound
	}
	return row.rotatedAt, nil
}
func (f *fakeTokens) RevokeRefreshToken(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if row, ok := f.rows[id]; ok {
		row.revoked = true
	}
	return nil
}
func (f *fakeTokens) RevokeAllUserTokens(_ context.Context, userID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, row := range f.rows {
		if row.userID == userID {
			row.revoked = true
		}
	}
	return nil
}
func (f *fakeTokens) activeCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, row := range f.rows {
		if !row.revoked {
			n++
		}
	}
	return n
}

func newTestUseCase(t *testing.T) (*UseCase, *fakeTokens) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	users := &fakeUsers{byName: map[string]domain.User{
		"egor": {ID: 1, Username: "egor", PasswordHash: string(hash), Role: domain.RoleUser},
	}}
	tokens := &fakeTokens{rows: map[string]*tokenRow{}}
	m, err := authjwt.NewManager("test-secret-that-is-long-enough-for-hs256")
	if err != nil {
		t.Fatal(err)
	}
	return New(users, tokens, m, 15*time.Minute, 24*time.Hour), tokens
}

func TestSignInSameErrorForUnknownUserAndWrongPassword(t *testing.T) {
	uc, _ := newTestUseCase(t)
	ctx := context.Background()

	_, errUnknown := uc.SignIn(ctx, "nobody", "whatever-password")
	_, errWrong := uc.SignIn(ctx, "egor", "wrong-password")
	if !errors.Is(errUnknown, errs.ErrInvalidCredentials) || !errors.Is(errWrong, errs.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials for both, got %v and %v", errUnknown, errWrong)
	}
}

func TestSignInIssuesAccessAndRefresh(t *testing.T) {
	uc, tokens := newTestUseCase(t)
	pair, err := uc.SignIn(context.Background(), "egor", "correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uc.jwt.Parse(pair.AccessToken, authjwt.TypeAccess); err != nil {
		t.Fatalf("access token invalid: %v", err)
	}
	if _, err := uc.jwt.Parse(pair.RefreshToken, authjwt.TypeAccess); !errors.Is(err, authjwt.ErrWrongTokenType) {
		t.Fatalf("refresh token must not work as access, got %v", err)
	}
	if tokens.activeCount() != 1 {
		t.Fatalf("want 1 stored refresh token, got %d", tokens.activeCount())
	}
}

func TestRefreshRotatesAndDetectsReuse(t *testing.T) {
	uc, tokens := newTestUseCase(t)
	ctx := context.Background()

	first, err := uc.SignIn(ctx, "egor", "correct-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := uc.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh token must rotate")
	}

	// Старый токен пришёл снова через минуту — похоже на кражу. Отзываем всё.
	uc.now = func() time.Time { return time.Now().Add(time.Minute) }
	if _, err := uc.Refresh(ctx, first.RefreshToken); !errors.Is(err, errs.ErrTokenReused) {
		t.Fatalf("want ErrTokenReused, got %v", err)
	}
	if tokens.activeCount() != 0 {
		t.Fatalf("all sessions must be revoked, %d still active", tokens.activeCount())
	}
	// Новейший токен тоже отозван, и это не «гонка вкладок»: он отозван
	// из-за кражи, а не ротацией.
	if _, err := uc.Refresh(ctx, second.RefreshToken); !errors.Is(err, errs.ErrTokenReused) {
		t.Fatalf("the newest token must be revoked as reused, got %v", err)
	}
}

func TestRefreshRejectsAccessToken(t *testing.T) {
	uc, _ := newTestUseCase(t)
	pair, _ := uc.SignIn(context.Background(), "egor", "correct-password")
	if _, err := uc.Refresh(context.Background(), pair.AccessToken); !errors.Is(err, errs.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestLogoutRevokesRefreshToken(t *testing.T) {
	uc, _ := newTestUseCase(t)
	ctx := context.Background()
	pair, _ := uc.SignIn(ctx, "egor", "correct-password")
	if err := uc.Logout(ctx, pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Refresh(ctx, pair.RefreshToken); err == nil {
		t.Fatal("refresh after logout must fail")
	}
}

func TestRefreshRaceBetweenTabsIsNotTreatedAsTheft(t *testing.T) {
	uc, tokens := newTestUseCase(t)
	ctx := context.Background()

	pair, _ := uc.SignIn(ctx, "egor", "correct-password")
	if _, err := uc.Refresh(ctx, pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	// Вторая вкладка пришла с тем же токеном через долю секунды.
	if _, err := uc.Refresh(ctx, pair.RefreshToken); !errors.Is(err, errs.ErrConcurrentRefresh) {
		t.Fatalf("want ErrConcurrentRefresh within grace period, got %v", err)
	}
	if tokens.activeCount() != 1 {
		t.Fatalf("the rotated session must survive, active = %d", tokens.activeCount())
	}
}
