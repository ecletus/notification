package notification

import (
	"github.com/ecletus/common"
	"github.com/ecletus/core"
)

func (notification *Notification) RegisterChannel(channel ChannelInterface) {
	notification.Channels = append(notification.Channels, channel)
}

type ChannelInterface interface {
	Send(message *Message, context *core.Context) error
	GetNotifications(user common.User, results *NotificationsResult, notification *Notification, context *core.Context) error
	GetUnresolvedNotificationsCount(user common.User, notification *Notification, context *core.Context) uint
	GetNotification(user common.User, notificationID string, notification *Notification, context *core.Context) (*QorNotification, error)
}
