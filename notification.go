package notification

import (
	"github.com/aghape/admin"
	"github.com/aghape/core"
	"github.com/aghape/core/resource"
	"github.com/aghape/core/utils"
	"github.com/aghape/roles"
)

type Notification struct {
	Config   *Config
	Channels []ChannelInterface
	Actions  []*Action
}

func New(config *Config) *Notification {
	notification := &Notification{Config: config}
	return notification
}

func (notification *Notification) Send(message *Message, context *core.Context) error {
	for _, channel := range notification.Channels {
		if err := channel.Send(message, context); err != nil {
			return err
		}
	}
	return nil
}

type NotificationsResult struct {
	Notification  *Notification
	Notifications []*QorNotification
	Resolved      []*QorNotification
}

func (notification *Notification) GetNotifications(user interface{}, context *core.Context) *NotificationsResult {
	var results = NotificationsResult{
		Notification: notification,
	}

	for _, channel := range notification.Channels {
		channel.GetNotifications(user, &results, notification, context)
	}

	return &results
}

func (notification *Notification) GetUnresolvedNotificationsCount(user interface{}, context *core.Context) uint {
	var result uint
	for _, channel := range notification.Channels {
		result += channel.GetUnresolvedNotificationsCount(user, notification, context)
	}
	return result
}

func (notification *Notification) GetNotification(user interface{}, messageID string, context *core.Context) *QorNotification {
	for _, channel := range notification.Channels {
		if message, err := channel.GetNotification(user, messageID, notification, context); err == nil {
			return message
		}
	}
	return nil
}

func (notification *Notification) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		Admin := res.GetAdmin()

		if len(notification.Channels) == 0 {
			utils.ExitWithMsg("No channel defined for notification")
		}

		Admin.RegisterFuncMap("unresolved_notifications_count", func(context *admin.Context) uint {
			return notification.GetUnresolvedNotificationsCount(context.CurrentUser, context.Context)
		})

		router := Admin.Router
		notificationController := controller{Notification: notification}

		router.Get("/!notifications", admin.NewHandler(notificationController.List, &admin.RouteConfig{
			PermissionMode: roles.Read,
			Resource:       res,
		}))

		for _, action := range notification.Actions {
			actionController := controller{Notification: notification, action: action}
			actionParam := "/!notifications/" + action.ToParam()
			res.ObjectRouter.Get(actionParam, admin.NewHandler(actionController.Action, &admin.RouteConfig{
				PermissionMode: roles.Update,
				Resource:       res,
			}))

			res.ObjectRouter.Put(actionParam, admin.NewHandler(actionController.Action, &admin.RouteConfig{
				PermissionMode: roles.Update,
				Resource:       res,
			}))

			if action.Undo != nil {
				res.ObjectRouter.Put(actionParam + "/undo", admin.NewHandler(actionController.UndoAction, &admin.RouteConfig{
					PermissionMode: roles.Update,
					Resource:       res,
				}))
			}
		}
	}
}
