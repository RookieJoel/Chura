package in

type NotificationEventHandler interface {
	HandleNotificationMessage(payload []byte) (requeue bool, err error)
}

type NotificationService interface {
	SendToReporter(reporterID, subject, body string) error
	SendSprintFinishedNotifications(sprintID string, reporterIDs []string) error
	SendWorkItemNotification(userID, subject, body string) error
}
