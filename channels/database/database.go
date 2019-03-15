package database

import (
	"fmt"
	"strconv"

	"github.com/aghape/common"

	"github.com/aghape/core"
	"github.com/aghape/notification"
	"github.com/moisespsena-go/aorm"
)

type Config struct {
	DBName string
}

func (c *Config) DBNameOrSystem() string {
	if c.DBName == "" {
		c.DBName = "system"
	}
	return c.DBName
}

func New() *Database {
	return &Database{&Config{}}
}

type Database struct {
	Config *Config
}

func (d *Database) Setup(db *aorm.DB) error {
	return db.AutoMigrate(&notification.QorNotification{}).Error
}

func (database *Database) GetDB(context *core.Context) *aorm.DB {
	if context.Site == nil {
		return context.DB
	}
	return context.Site.GetDB(database.Config.DBNameOrSystem()).DB
}

func (database *Database) Send(message *notification.Message, context *core.Context) error {
	notice := notification.QorNotification{
		From:        message.From.GetID(),
		To:          message.From.GetID(),
		Title:       message.Title,
		Body:        message.Body,
		MessageType: message.MessageType,
		ResolvedAt:  message.ResolvedAt,
	}

	return database.GetDB(context).Save(&notice).Error
}

func (database *Database) GetNotifications(user common.User, results *notification.NotificationsResult, _ *notification.Notification, context *core.Context) error {
	var to = user.GetID()
	var db = database.GetDB(context)

	var currentPage, perPage int

	if context.Request != nil {
		if p, err := strconv.Atoi(context.Request.URL.Query().Get("page")); err == nil {
			currentPage = p
		}

		if p, err := strconv.Atoi(context.Request.URL.Query().Get("per_page")); err == nil {
			perPage = p
		}
	}

	if perPage == 0 {
		perPage = 10
	}
	offset := currentPage * perPage

	commonDB := db.Order("created_at DESC").Where(fmt.Sprintf("%v = ?", db.Dialect().Quote("to")), to)

	// get unresolved notifications
	if err := commonDB.Offset(offset).Limit(perPage).Find(&results.Notifications, fmt.Sprintf("%v IS NULL", db.Dialect().Quote("resolved_at"))).Error; err != nil {
		return err
	}

	if len(results.Notifications) == perPage {
		return nil
	}

	if len(results.Notifications) == 0 {
		var unreadedCount int
		commonDB.Model(&notification.QorNotification{}).Where(fmt.Sprintf("%v IS NULL", db.Dialect().Quote("resolved_at"))).Count(&unreadedCount)
		offset -= unreadedCount
	} else if len(results.Notifications) < perPage {
		offset = 0
		perPage -= len(results.Notifications)
	}

	// get resolved notifications
	return commonDB.Offset(offset).Limit(perPage).Find(&results.Resolved, fmt.Sprintf("%v IS NOT NULL", db.Dialect().Quote("resolved_at"))).Error
}

func (database *Database) GetUnresolvedNotificationsCount(user common.User, _ *notification.Notification, context *core.Context) uint {
	var to = user.GetID()
	var db = database.GetDB(context)

	var result uint
	db.Model(&notification.QorNotification{}).Where(fmt.Sprintf("%v = ? AND %v IS NULL", db.Dialect().Quote("to"), db.Dialect().Quote("resolved_at")), to).Count(&result)
	return result
}

func (database *Database) GetNotification(user common.User, notificationID string, _ *notification.Notification, context *core.Context) (*notification.QorNotification, error) {
	var (
		notice notification.QorNotification
		to     = user.GetID()
		db     = database.GetDB(context)
	)

	err := db.First(&notice, fmt.Sprintf("%v = ? AND %v = ?", db.Dialect().Quote("to"), db.Dialect().Quote("id")), to, notificationID).Error
	return &notice, err
}
