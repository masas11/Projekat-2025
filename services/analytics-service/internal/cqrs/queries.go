package cqrs

import "analytics-service/internal/model"

// Query represents a query in CQRS pattern (2.15)
type Query interface {
	GetUserID() string
}

// GetUserAnalyticsQuery represents a query to get user analytics
type GetUserAnalyticsQuery struct {
	UserID string
}

func (q *GetUserAnalyticsQuery) GetUserID() string {
	return q.UserID
}

// QueryResult represents the result of executing a query
type QueryResult struct {
	Analytics *model.UserAnalytics
	Error     error
}
