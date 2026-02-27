package model

import (
	"time"
)

// SagaStatus represents the status of a saga transaction
type SagaStatus string

const (
	SagaStatusPending    SagaStatus = "PENDING"
	SagaStatusInProgress SagaStatus = "IN_PROGRESS"
	SagaStatusCompleted  SagaStatus = "COMPLETED"
	SagaStatusFailed     SagaStatus = "FAILED"
	SagaStatusCompensating SagaStatus = "COMPENSATING"
	SagaStatusCompensated  SagaStatus = "COMPENSATED"
)

// StepStatus represents the status of a single step
type StepStatus string

const (
	StepStatusPending    StepStatus = "PENDING"
	StepStatusCompleted  StepStatus = "COMPLETED"
	StepStatusFailed     StepStatus = "FAILED"
	StepStatusCompensated StepStatus = "COMPENSATED"
)

// SagaTransaction represents a saga transaction
type SagaTransaction struct {
	ID          string                 `bson:"_id" json:"id"`
	Type        string                 `bson:"type" json:"type"`
	Status      SagaStatus             `bson:"status" json:"status"`
	SongID      string                 `bson:"songId" json:"songId"`
	SongData    map[string]interface{} `bson:"songData,omitempty" json:"songData,omitempty"` // Backup of song data
	Steps       []SagaStep             `bson:"steps" json:"steps"`
	CreatedAt   time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time              `bson:"updatedAt" json:"updatedAt"`
	Error       string                 `bson:"error,omitempty" json:"error,omitempty"`
}

// SagaStep represents a single step in a saga transaction
type SagaStep struct {
	Name           string                 `bson:"name" json:"name"`
	Status         StepStatus             `bson:"status" json:"status"`
	Order          int                    `bson:"order" json:"order"`
	ExecutedAt     *time.Time             `bson:"executedAt,omitempty" json:"executedAt,omitempty"`
	CompensatedAt  *time.Time             `bson:"compensatedAt,omitempty" json:"compensatedAt,omitempty"`
	Error          string                 `bson:"error,omitempty" json:"error,omitempty"`
	Data           map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"` // Step-specific data
}

// Step names for song deletion saga
const (
	StepBackupSong      = "BACKUP_SONG"
	StepDeleteRatings   = "DELETE_RATINGS"
	StepDeleteFromNeo4j = "DELETE_FROM_NEO4J"
	StepDeleteFromHDFS  = "DELETE_FROM_HDFS"
	StepDeleteFromMongo = "DELETE_FROM_MONGO"
)

// Compensating step names
const (
	CompensateRestoreToNeo4j = "RESTORE_TO_NEO4J"
	CompensateRestoreToHDFS  = "RESTORE_TO_HDFS"
	CompensateRestoreToMongo = "RESTORE_TO_MONGO"
)
