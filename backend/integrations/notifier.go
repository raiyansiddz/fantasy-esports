package integrations

import (
	"fantasy-esports-backend/models"
)

// Notifier interface defines the contract for all notification providers
type Notifier interface {
	Send(request *models.SendNotificationRequest, config map[string]string) (*models.NotificationResponse, error)
	ValidateConfig(config map[string]string) error
	GetProviderName() models.NotificationProvider
	GetChannel() models.NotificationChannel
}

// NotifierFactory creates notifier instances based on provider and channel
type NotifierFactory struct{}

// NewNotifierFactory creates a new notifier factory
func NewNotifierFactory() *NotifierFactory {
	return &NotifierFactory{}
}

// CreateNotifier creates a notifier based on provider and channel
func (f *NotifierFactory) CreateNotifier(provider models.NotificationProvider, channel models.NotificationChannel) (Notifier, error) {
	switch channel {
	case models.ChannelSMS:
		switch provider {
		case models.ProviderFast2SMS:
			return NewFast2SMSNotifier(), nil
		}
	case models.ChannelEmail:
		switch provider {
		case models.ProviderSMTP:
			return NewSMTPNotifier(), nil
		case models.ProviderSES:
			return NewSESNotifier(), nil
		case models.ProviderMailchimp:
			return NewMailchimpNotifier(), nil
		}
	case models.ChannelPush:
		switch provider {
		case models.ProviderFCM:
			return NewFCMNotifier(), nil
		case models.ProviderOneSignal:
			return NewOneSignalNotifier(), nil
		}
	case models.ChannelWhatsApp:
		switch provider {
		case models.ProviderWhatsAppCloud:
			return NewWhatsAppNotifier(), nil
		}
	}
	
	return nil, NewNotificationError("UNSUPPORTED_PROVIDER", "Unsupported provider/channel combination", nil)
}