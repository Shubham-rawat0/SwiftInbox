package service

var validEvents = map[string]bool{"email.received": true, "email.deleted": true, "mailbox.expired": true}

func ValidateEvents(events []string) error {
	for _, event := range events {
		if !validEvents[event] {
			return ErrInvalidWebhookEvent
		}
	}
	return nil
}
