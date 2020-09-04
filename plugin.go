package notification

import (
	"github.com/ecletus/db"
	"github.com/ecletus/plug"
)

type Plugin struct {
	db.DBNames
	plug.EventDispatcher
}

func (p *Plugin) OnRegister() {
	db.Events(p).DBOnMigrate(func(e *db.DBEvent) error {
		return e.AutoMigrate(&QorNotification{}).Error
	})
}

type PluginDefaultNotification struct {
	NotificationKey string
}

func (p *PluginDefaultNotification) ProvideOptions() []string {
	return []string{p.NotificationKey}
}

func (p *PluginDefaultNotification) ProvidesOptions(options *plug.Options) {
	options.Set(p.NotificationKey, New(&Config{}))
}
