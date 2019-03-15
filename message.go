package notification

import (
	"github.com/aghape/common"
	"time"

	"github.com/moisespsena-go/aorm"
	"github.com/aghape/admin"
)

type Message struct {
	From        common.User
	To          common.User
	Title       string
	Body        string
	MessageType string
	ResolvedAt  *time.Time
}

type QorNotification struct {
	aorm.KeyStringSerial
	aorm.Timestamps
	From        string
	To          string
	Title       string
	Body        string `sql:"size:65532"`
	MessageType string
	ResolvedAt  *time.Time
}

func (qorNotification QorNotification) IsResolved() bool {
	return qorNotification.ResolvedAt != nil
}

func (qorNotification *QorNotification) Actions(context *admin.Context) (actions []*Action) {
	var globalActions []*Action
	if n := context.Get("Notification"); n != nil {
		if notification, ok := n.(*Notification); ok {
			for _, action := range notification.Actions {
				if action.HasMessageType(qorNotification.MessageType) {
					if action.Visible != nil {
						if !action.Visible(qorNotification, context) {
							continue
						}
					}

					if len(action.MessageTypes) == 0 {
						globalActions = append(globalActions, action)
					} else {
						actions = append(actions, action)
					}
				}
			}
		}
	}

	return append(actions, globalActions...)
}
