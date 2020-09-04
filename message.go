package notification

import (
	"time"

	"github.com/moisespsena-go/bid"

	"github.com/ecletus/admin"
	"github.com/ecletus/auth"
	"github.com/moisespsena-go/aorm"
)

type Message struct {
	From        auth.User
	To          auth.User
	Title       string
	Body        string
	MessageType string
	ResolvedAt  *time.Time
}

type QorNotification struct {
	aorm.Model
	aorm.Timestamps
	From        bid.BID
	To          bid.BID
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
	if n := context.GetSettings("Notification"); n != nil {
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
