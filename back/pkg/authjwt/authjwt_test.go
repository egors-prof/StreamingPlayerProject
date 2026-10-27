package authjwt

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-that-is-long-enough-for-hs256"

func newTestManager(t *testing.T, now time.Time) *Manager {
	t.Helper()
	m, err := NewManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return now }
	return m
}

func TestIssueSetsRealExpiry(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	m := newTestManager(t, now)

	_, claims, err := m.Issue(Claims{UserID: 7, Type: TypeAccess}, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := claims.ExpiresAt.Time, now.Add(5*time.Minute); !got.Equal(want) {
		t.Fatalf("ExpiresAt = %v, want %v", got, want)
	}
}

func TestParse(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	m := newTestManager(t, now)

	access, _, _ := m.Issue(Claims{UserID: 7, Username: "egor", Role: RoleUser, Type: TypeAccess}, 5*time.Minute)
	refresh, _, _ := m.Issue(Claims{UserID: 7, Type: TypeRefresh}, 24*time.Hour)

	other, _ := NewManager("another-secret-that-is-long-enough-too")
	forged, _, _ := other.Issue(Claims{UserID: 1, Type: TypeAccess}, time.Hour)

	noneAlg, _ := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: 1, Type: TypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{Issuer: issuer, ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	tests := []struct {
		name    string
		token   string
		want    TokenType
		at      time.Time
		wantErr error
	}{
		{"valid access", access, TypeAccess, now, nil},
		{"valid refresh", refresh, TypeRefresh, now, nil},
		{"refresh used as access", refresh, TypeAccess, now, ErrWrongTokenType},
		{"access used as refresh", access, TypeRefresh, now, ErrWrongTokenType},
		{"expired access", access, TypeAccess, now.Add(6 * time.Minute), ErrExpiredToken},
		{"signed with another secret", forged, TypeAccess, now, ErrInvalidToken},
		{"alg none", noneAlg, TypeAccess, now, ErrInvalidToken},
		{"garbage", "not.a.token", TypeAccess, now, ErrInvalidToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.now = func() time.Time { return tt.at }
			claims, err := m.Parse(tt.token, tt.want)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err == nil && claims.UserID != 7 {
				t.Fatalf("UserID = %d, want 7", claims.UserID)
			}
		})
	}
}

func TestNewManagerRejectsShortSecret(t *testing.T) {
	if _, err := NewManager("short"); err == nil {
		t.Fatal("expected error for short secret")
	}
}

func TestBearerToken(t *testing.T) {
	tok, err := BearerToken("Bearer abc.def.ghi")
	if err != nil || tok != "abc.def.ghi" {
		t.Fatalf("got %q, %v", tok, err)
	}
	for _, h := range []string{"", "abc", "Basic abc", "Bearer ", strings.Repeat(" ", 3)} {
		if _, err := BearerToken(h); !errors.Is(err, ErrNoBearerToken) {
			t.Fatalf("header %q: err = %v", h, err)
		}
	}
}
