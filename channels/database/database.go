package database

import (
	"fmt"
	"github.com/moisespsena-go/bid"
	"strconv"

	"github.com/ecletus/common"

	"github.com/ecletus/core"
	"github.com/ecletus/notification"
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
		return context.DB()
	}
	return context.Site.GetDB(database.Config.DBNameOrSystem()).DB
}

func (database *Database) Send(message *notification.Message, context *core.Context) error {
	var ibid = func(id interface{}) bid.BID {
		return id.(bid.BID)
	}
	notice := notification.QorNotification{
		From:        ibid(aorm.IdOf(message.From)),
		To:          ibid(aorm.IdOf(message.To)),
		Title:       message.Title,
		Body:        message.Body,
		MessageType: message.MessageType,
		ResolvedAt:  message.ResolvedAt,
	}

	return database.GetDB(context).Save(&notice).Error
}

func (database *Database) GetNotifications(user common.User, results *notification.NotificationsResult, _ *notification.Notification, context *core.Context) error {
	var to = aorm.IdOf(user)
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

	commonDB := db.Order("created_at DESC").Where(fmt.Sprintf("%v = ?", aorm.Quote(db.Dialect(), "to")), to)

	// get unresolved notifications
	if err := commonDB.Offset(offset).Limit(perPage).Find(&results.Notifications, fmt.Sprintf("%v IS NULL", aorm.Quote(db.Dialect(), "resolved_at"))).Error; err != nil {
		return err
	}

	if len(results.Notifications) == perPage {
		return nil
	}

	if len(results.Notifications) == 0 {
		var unreadedCount int
		commonDB.Model(&notification.QorNotification{}).Where(fmt.Sprintf("%v IS NULL", aorm.Quote(db.Dialect(), "resolved_at"))).Count(&unreadedCount)
		offset -= unreadedCount
	} else if len(results.Notifications) < perPage {
		offset = 0
		perPage -= len(results.Notifications)
	}

	// get resolved notifications
	return commonDB.Offset(offset).Limit(perPage).Find(&results.Resolved, fmt.Sprintf("%v IS NOT NULL", aorm.Quote(db.Dialect(), "resolved_at"))).Error
}

func (database *Database) GetUnresolvedNotificationsCount(user common.User, _ *notification.Notification, context *core.Context) uint {
	var to = aorm.IdOf(user)
	var db = database.GetDB(context)

	var result uint
	db.Model(&notification.QorNotification{}).Where(fmt.Sprintf("%v = ? AND %v IS NULL", aorm.Quote(db.Dialect(), "to"), aorm.Quote(db.Dialect(), "resolved_at")), to).Count(&result)
	return result
}

func (database *Database) GetNotification(user common.User, notificationID string, _ *notification.Notification, context *core.Context) (*notification.QorNotification, error) {
	var (
		notice notification.QorNotification
		to     = aorm.IdOf(user)
		db     = database.GetDB(context)
	)

	err := db.First(&notice, fmt.Sprintf("%v = ? AND %v = ?", aorm.Quote(db.Dialect(), "to"), aorm.Quote(db.Dialect(), "id")), to, notificationID).Error
	return &notice, err
}
