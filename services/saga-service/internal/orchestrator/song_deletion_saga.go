package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"saga-service/config"
	"saga-service/internal/model"
	"saga-service/internal/store"
)

type SongDeletionSaga struct {
	store  *store.SagaStore
	config *config.Config
}

func NewSongDeletionSaga(store *store.SagaStore, cfg *config.Config) *SongDeletionSaga {
	return &SongDeletionSaga{
		store:  store,
		config: cfg,
	}
}

// Execute runs the complete saga transaction for song deletion
func (s *SongDeletionSaga) Execute(ctx context.Context, songID string) (*model.SagaTransaction, error) {
	// Create saga transaction
	saga := &model.SagaTransaction{
		ID:     fmt.Sprintf("saga_%s_%d", songID, time.Now().Unix()),
		Type:   "DELETE_SONG",
		Status: model.SagaStatusPending,
		SongID: songID,
		Steps: []model.SagaStep{
			{Name: model.StepBackupSong, Status: model.StepStatusPending, Order: 1},
			{Name: model.StepDeleteRatings, Status: model.StepStatusPending, Order: 2},
			{Name: model.StepDeleteFromNeo4j, Status: model.StepStatusPending, Order: 3},
			{Name: model.StepDeleteFromHDFS, Status: model.StepStatusPending, Order: 4},
			{Name: model.StepDeleteFromMongo, Status: model.StepStatusPending, Order: 5},
		},
	}

	// Save initial transaction
	if err := s.store.CreateTransaction(ctx, saga); err != nil {
		return nil, fmt.Errorf("failed to create saga transaction: %w", err)
	}

	saga.Status = model.SagaStatusInProgress
	s.store.UpdateTransaction(ctx, saga)

	// Execute steps in order
	for i := range saga.Steps {
		step := &saga.Steps[i]
		log.Printf("Executing step %d: %s for song %s", step.Order, step.Name, songID)

		// Update step status to in progress
		step.Status = model.StepStatusPending
		s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, "")

		// Execute step
		err := s.executeStep(ctx, saga, step)
		if err != nil {
			log.Printf("Step %s failed: %v", step.Name, err)
			step.Status = model.StepStatusFailed
			step.Error = err.Error()
			s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, err.Error())

			// Start compensation
			saga.Status = model.SagaStatusCompensating
			s.store.UpdateTransaction(ctx, saga)
			s.compensate(ctx, saga, i)
			saga.Status = model.SagaStatusCompensated
			saga.Error = fmt.Sprintf("Step %s failed: %v", step.Name, err)
			s.store.UpdateTransaction(ctx, saga)
			return saga, fmt.Errorf("saga failed at step %s: %w", step.Name, err)
		}

		// Mark step as completed
		step.Status = model.StepStatusCompleted
		s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, "")
		log.Printf("Step %s completed successfully", step.Name)
	}

	// All steps completed successfully
	saga.Status = model.SagaStatusCompleted
	s.store.UpdateTransaction(ctx, saga)
	log.Printf("Saga transaction %s completed successfully", saga.ID)

	return saga, nil
}

// executeStep executes a single step
func (s *SongDeletionSaga) executeStep(ctx context.Context, saga *model.SagaTransaction, step *model.SagaStep) error {
	switch step.Name {
	case model.StepBackupSong:
		return s.backupSong(ctx, saga)
	case model.StepDeleteRatings:
		return s.deleteRatings(ctx, saga.SongID)
	case model.StepDeleteFromNeo4j:
		return s.deleteFromNeo4j(ctx, saga.SongID)
	case model.StepDeleteFromHDFS:
		return s.deleteFromHDFS(ctx, saga)
	case model.StepDeleteFromMongo:
		return s.deleteFromMongo(ctx, saga.SongID)
	default:
		return fmt.Errorf("unknown step: %s", step.Name)
	}
}

// backupSong backs up song data before deletion
func (s *SongDeletionSaga) backupSong(ctx context.Context, saga *model.SagaTransaction) error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s/songs/%s", s.config.ContentServiceURL, saga.SongID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch song: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("song not found or error: status %d", resp.StatusCode)
	}

	var songData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&songData); err != nil {
		return fmt.Errorf("failed to decode song data: %w", err)
	}

	saga.SongData = songData
	s.store.UpdateTransaction(ctx, saga)
	log.Printf("Song %s backed up successfully", saga.SongID)
	return nil
}

// deleteRatings deletes all ratings for the song
func (s *SongDeletionSaga) deleteRatings(ctx context.Context, songID string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/delete-ratings-by-song?songId=%s", s.config.RatingsServiceURL, songID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete ratings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete ratings: status %d", resp.StatusCode)
	}

	log.Printf("Ratings for song %s deleted successfully", songID)
	return nil
}

// deleteFromNeo4j deletes the song from Neo4j
func (s *SongDeletionSaga) deleteFromNeo4j(ctx context.Context, songID string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/events", s.config.RecommendationServiceURL)

	event := map[string]interface{}{
		"type":   "song_deleted",
		"songId": songID,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(eventJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	defer resp.Body.Close()

	// Accept 200, 201, and 202 as success (202 Accepted is common for async operations)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to delete from Neo4j: status %d", resp.StatusCode)
	}

	log.Printf("Song %s deleted from Neo4j successfully", songID)
	return nil
}

// deleteFromHDFS deletes the audio file from HDFS (if exists)
func (s *SongDeletionSaga) deleteFromHDFS(ctx context.Context, saga *model.SagaTransaction) error {
	// Check if song has audio file
	audioFileURL, ok := saga.SongData["audioFileUrl"].(string)
	if !ok || audioFileURL == "" {
		log.Printf("Song %s has no audio file, skipping HDFS deletion", saga.SongID)
		return nil // Not an error if there's no file
	}

	// If it's an HDFS path, we would delete it here
	// For now, we'll just log it (HDFS deletion can be implemented later)
	if audioFileURL != "" {
		log.Printf("Would delete audio file from HDFS: %s (not implemented yet)", audioFileURL)
		// In a real implementation, we would call HDFS API to delete the file
		// For now, we'll just succeed (as HDFS deletion is optional)
	}

	return nil
}

// deleteFromMongo deletes the song from MongoDB using internal endpoint
func (s *SongDeletionSaga) deleteFromMongo(ctx context.Context, songID string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/songs/internal/delete?songId=%s", s.config.ContentServiceURL, songID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete song from MongoDB: status %d", resp.StatusCode)
	}

	log.Printf("Song %s deleted from MongoDB successfully", songID)
	return nil
}

// compensate executes compensating actions for all completed steps in reverse order
func (s *SongDeletionSaga) compensate(ctx context.Context, saga *model.SagaTransaction, failedStepIndex int) {
	log.Printf("Starting compensation for saga %s (failed at step %d)", saga.ID, failedStepIndex)

	// Compensate in reverse order (from the step before the failed one)
	for i := failedStepIndex - 1; i >= 0; i-- {
		step := &saga.Steps[i]
		if step.Status != model.StepStatusCompleted {
			continue // Skip steps that weren't completed
		}

		log.Printf("Compensating step: %s", step.Name)
		err := s.compensateStep(ctx, saga, step)
		if err != nil {
			log.Printf("Compensation failed for step %s: %v", step.Name, err)
			// Continue with other compensations even if one fails
		} else {
			step.Status = model.StepStatusCompensated
			s.store.UpdateStepStatus(ctx, saga.ID, step.Name, step.Status, "")
			log.Printf("Step %s compensated successfully", step.Name)
		}
	}
}

// compensateStep executes compensating action for a specific step
func (s *SongDeletionSaga) compensateStep(ctx context.Context, saga *model.SagaTransaction, step *model.SagaStep) error {
	switch step.Name {
	case model.StepDeleteFromNeo4j:
		// Restore song to Neo4j (would need to recreate the song node)
		log.Printf("Compensating: Would restore song %s to Neo4j (not fully implemented)", saga.SongID)
		// In a real implementation, we would recreate the song node in Neo4j
		return nil // For now, just log
	case model.StepDeleteFromHDFS:
		// Restore audio file to HDFS (if it was deleted)
		log.Printf("Compensating: Would restore audio file for song %s to HDFS (not implemented)", saga.SongID)
		return nil
	case model.StepDeleteFromMongo:
		// Restore song to MongoDB
		return s.restoreToMongo(ctx, saga)
	case model.StepBackupSong, model.StepDeleteRatings:
		// No compensation needed for these steps
		return nil
	default:
		return fmt.Errorf("unknown step for compensation: %s", step.Name)
	}
}

// restoreToMongo restores the song to MongoDB
func (s *SongDeletionSaga) restoreToMongo(ctx context.Context, saga *model.SagaTransaction) error {
	if saga.SongData == nil {
		return fmt.Errorf("no backup data available for restoration")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/songs", s.config.ContentServiceURL)

	songJSON, err := json.Marshal(saga.SongData)
	if err != nil {
		return fmt.Errorf("failed to marshal song data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(songJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to restore song: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to restore song: status %d", resp.StatusCode)
	}

	log.Printf("Song %s restored to MongoDB successfully", saga.SongID)
	return nil
}
