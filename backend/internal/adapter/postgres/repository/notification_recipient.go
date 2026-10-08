package memory

import "errors"

type NotificationRecipientRepository struct {
	TestEmail string
}

func (repository NotificationRecipientRepository) GetEmailByUserID(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("user_id is required")
	}
	return repository.TestEmail, nil
}
