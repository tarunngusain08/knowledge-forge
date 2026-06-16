package notifications

type Sender interface {
	Send(to string, subject string, body string) error
}

type NotificationService struct {
	sender Sender
}

func NewNotificationService(sender Sender) *NotificationService {
	return &NotificationService{sender: sender}
}

func (s *NotificationService) SendWelcomeEmail(email string) error {
	return s.sender.Send(email, "Welcome", "Thanks for joining")
}
