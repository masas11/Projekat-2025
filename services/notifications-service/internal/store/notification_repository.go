package store

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"

	"notifications-service/internal/model"
)

type NotificationRepository struct {
	session *gocql.Session
}

func NewNotificationRepository(session *gocql.Session) *NotificationRepository {
	return &NotificationRepository{
		session: session,
	}
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Notification, error) {
	query := `SELECT id, user_id, type, message, content_id, read, created_at 
	          FROM notifications 
	          WHERE user_id = ? 
	          ORDER BY created_at DESC`

	iter := r.session.Query(query, userID).WithContext(ctx).Iter()
	defer iter.Close()

	var notifications []*model.Notification
	var id, userId, notifType, message, contentID string
	var read bool
	var createdAt time.Time

	for iter.Scan(&id, &userId, &notifType, &message, &contentID, &read, &createdAt) {
		notifications = append(notifications, &model.Notification{
			ID:        id,
			UserID:    userId,
			Type:      notifType,
			Message:   message,
			ContentID: contentID,
			Read:      read,
			CreatedAt: createdAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close iterator: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	query := `INSERT INTO notifications (id, user_id, type, message, content_id, read, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	err := r.session.Query(
		query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Message,
		notification.ContentID,
		notification.Read,
		notification.CreatedAt,
	).WithContext(ctx).Exec()

	if err != nil {
		return fmt.Errorf("failed to insert notification: %w", err)
	}

	return nil
}
