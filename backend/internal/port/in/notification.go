package in

type NotificationService interface {
	SendToReporter(reporterID, subject, body string) error
	SendSprintFinishedNotifications(sprintID string, reporterIDs []string) error
	SendWorkItemNotification(userID, subject, body string) error
}
