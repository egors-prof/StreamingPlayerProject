package registration

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"golang.org/x/crypto/bcrypt"
)

type fakeUsers struct {
	created []domain.User
	taken   map[string]bool
}

func (f *fakeUsers) CreateUser(_ context.Context, u domain.User) (int, error) {
	if f.taken[u.Username] {
		return 0, errs.ErrUsernameTaken
	}
	f.created = append(f.created, u)
	return len(f.created), nil
}
func (f *fakeUsers) GetUserByUsername(context.Context, string) (domain.User, error) {
	return domain.User{}, errs.ErrNotFound
}
func (f *fakeUsers) GetUserByID(context.Context, int) (domain.User, error) {
	return domain.User{}, errs.ErrNotFound
}

type fakePublisher struct {
	events []domain.UserRegistered
	err    error
}

func (f *fakePublisher) PublishUserRegistered(_ context.Context, e domain.UserRegistered) error {
	f.events = append(f.events, e)
	return f.err
}

func newUC(users *fakeUsers, pub *fakePublisher) *UseCase {
	uc := New(users, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))
	uc.hashCost = bcrypt.MinCost // в тестах не тратим время на медленный хеш
	return uc
}

func TestSignUpHashesPasswordAndPublishesEvent(t *testing.T) {
	users, pub := &fakeUsers{}, &fakePublisher{}
	user, err := newUC(users, pub).SignUp(context.Background(), domain.SignUpInput{
		FullName: " Egor ", Username: "egor", Password: "long-enough-password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.FullName != "Egor" {
		t.Fatalf("full name must be trimmed, got %q", user.FullName)
	}
	stored := users.created[0].PasswordHash
	if stored == "long-enough-password" || bcrypt.CompareHashAndPassword([]byte(stored), []byte("long-enough-password")) != nil {
		t.Fatal("password must be stored as bcrypt hash")
	}
	if len(pub.events) != 1 || pub.events[0].UserID != user.ID {
		t.Fatalf("want one user.registered event, got %+v", pub.events)
	}
}

func TestSignUpValidation(t *testing.T) {
	tests := []struct {
		name string
		in   domain.SignUpInput
	}{
		{"short password", domain.SignUpInput{FullName: "A", Username: "egor", Password: "short"}},
		{"too long password", domain.SignUpInput{FullName: "A", Username: "egor", Password: string(make([]byte, 73))}},
		{"bad username", domain.SignUpInput{FullName: "A", Username: "eg or!", Password: "long-enough-password"}},
		{"empty name", domain.SignUpInput{FullName: "  ", Username: "egor", Password: "long-enough-password"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newUC(&fakeUsers{}, &fakePublisher{}).SignUp(context.Background(), tt.in)
			if !errors.Is(err, errs.ErrInvalidInput) {
				t.Fatalf("want ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestSignUpDuplicateUsername(t *testing.T) {
	users := &fakeUsers{taken: map[string]bool{"egor": true}}
	_, err := newUC(users, &fakePublisher{}).SignUp(context.Background(), domain.SignUpInput{
		FullName: "Egor", Username: "egor", Password: "long-enough-password",
	})
	if !errors.Is(err, errs.ErrUsernameTaken) {
		t.Fatalf("want ErrUsernameTaken, got %v", err)
	}
}

func TestSignUpSucceedsWhenBrokerIsDown(t *testing.T) {
	pub := &fakePublisher{err: errors.New("broker unavailable")}
	if _, err := newUC(&fakeUsers{}, pub).SignUp(context.Background(), domain.SignUpInput{
		FullName: "Egor", Username: "egor", Password: "long-enough-password",
	}); err != nil {
		t.Fatalf("registration must not fail because of the broker: %v", err)
	}
}
