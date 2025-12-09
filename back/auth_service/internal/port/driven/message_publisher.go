package driven

import "github.com/egors-prof/auth_service/internal/domain"

type MessagePublisher interface {
	PublishMessage(message domain.Message) error
}
