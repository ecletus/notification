package database

import (
	"github.com/aghape/notification"
	"github.com/aghape/plug"
)

type ChannelPlugin struct {
	NotificationKey string
}

func (p *ChannelPlugin) RequireOptions() []string {
	return []string{p.NotificationKey}
}

func (p *ChannelPlugin) Init(options *plug.Options) {
	n := options.GetInterface(p.NotificationKey).(*notification.Notification)
	channel := New()
	n.RegisterChannel(channel)
}
