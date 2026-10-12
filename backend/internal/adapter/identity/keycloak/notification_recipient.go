package keycloak

import (
	"context"
	"errors"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type NotificationRecipientRepository struct {
	directory *Directory
}

var _ out.NotificationRecipientRepository = (*NotificationRecipientRepository)(nil)

func NewNotificationRecipientRepository(directory *Directory) *NotificationRecipientRepository {
	return &NotificationRecipientRepository{directory: directory}
}

func (repository *NotificationRecipientRepository) GetEmailByUserID(userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("user_id is required")
	}
	user, err := repository.directory.FindUserByID(context.Background(), userID)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(user.Email) == "" {
		return "", errors.New("user has no email")
	}
	return user.Email, nil
}
