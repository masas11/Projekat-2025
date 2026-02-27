package cqrs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"analytics-service/config"
	"analytics-service/internal/model"
	"analytics-service/internal/store"
)

// QueryHandler handles queries and reads from read model (CQRS Query Side - 2.15)
type QueryHandler struct {
	projectionStore *store.ProjectionStore
	config          *config.Config
}

func NewQueryHandler(projectionStore *store.ProjectionStore, cfg *config.Config) *QueryHandler {
	return &QueryHandler{
		projectionStore: projectionStore,
		config:          cfg,
	}
}

// HandleQuery processes a query and returns results from read model
func (qh *QueryHandler) HandleQuery(ctx context.Context, query Query) *QueryResult {
	switch q := query.(type) {
	case *GetUserAnalyticsQuery:
		return qh.handleGetUserAnalyticsQuery(ctx, q)
	default:
		return &QueryResult{
			Error: fmt.Errorf("unknown query type"),
		}
	}
}

func (qh *QueryHandler) handleGetUserAnalyticsQuery(ctx context.Context, query *GetUserAnalyticsQuery) *QueryResult {
	// Read from projection (read model)
	projection, err := qh.projectionStore.GetProjection(ctx, query.UserID)
	if err != nil {
		log.Printf("Error getting projection: %v", err)
		return &QueryResult{
			Error: err,
		}
	}
	
	// Convert projection to UserAnalytics
	analytics := &model.UserAnalytics{
		UserID:             projection.UserID,
		TotalSongsPlayed:   projection.TotalSongsPlayed,
		SongsPlayedByGenre: projection.SongsPlayedByGenre,
	}
	
	// Calculate average rating
	if projection.TotalRatingsCount > 0 {
		analytics.AverageRating = projection.TotalRatingsSum / float64(projection.TotalRatingsCount)
	}
	
	// Get top 5 artists
	analytics.Top5Artists = qh.getTop5Artists(projection)
	
	// Get subscribed artists count
	analytics.SubscribedArtistsCount = len(projection.SubscribedArtists)
	
	// If we need additional data (artist names), fetch from content service
	if len(analytics.Top5Artists) > 0 {
		qh.enrichArtistNames(ctx, analytics, projection)
	}
	
	// Get subscribed artists count from subscriptions service (for accuracy)
	qh.updateSubscribedArtistsCount(ctx, analytics, query.UserID)
	
	return &QueryResult{
		Analytics: analytics,
	}
}

func (qh *QueryHandler) getTop5Artists(projection *store.AnalyticsProjection) []model.ArtistPlayCount {
	type artistData struct {
		ID        string
		Name      string
		PlayCount int
	}
	
	artists := make([]artistData, 0, len(projection.ArtistPlayCounts))
	
	// Collect all artist IDs that need names
	artistIDsToFetch := []string{}
	for artistID, count := range projection.ArtistPlayCounts {
		name := projection.ArtistNames[artistID]
		// Check if name is missing or is a fallback name
		if name == "" || strings.HasPrefix(name, "Umetnik ") {
			artistIDsToFetch = append(artistIDsToFetch, artistID)
		}
		artists = append(artists, artistData{
			ID:        artistID,
			Name:      name,
			PlayCount: count,
		})
	}
	
	// Fetch missing artist names from content service
	if len(artistIDsToFetch) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Timeout:   5 * time.Second,
			Transport: tr,
		}
		
		artistsURL := qh.config.ContentServiceURL + "/artists"
		artistsReq, err := http.NewRequestWithContext(ctx, "GET", artistsURL, nil)
		if err == nil {
			artistsResp, err := client.Do(artistsReq)
			if err == nil && artistsResp != nil && artistsResp.StatusCode == http.StatusOK {
				var allArtists []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}
				if json.NewDecoder(artistsResp.Body).Decode(&allArtists) == nil {
					// Build artist name map
					artistNameMap := make(map[string]string)
					for _, artist := range allArtists {
						artistNameMap[artist.ID] = artist.Name
					}
					
					// Update artist names in the list
					for i := range artists {
						if artists[i].Name == "" || strings.HasPrefix(artists[i].Name, "Umetnik ") {
							if name, ok := artistNameMap[artists[i].ID]; ok && name != "" {
								artists[i].Name = name
							} else if artists[i].Name == "" {
								// Only use fallback if we couldn't fetch the name
								artists[i].Name = "Umetnik " + artists[i].ID[:min(8, len(artists[i].ID))] + "..."
							}
						}
					}
				}
				artistsResp.Body.Close()
			}
		}
	}
	
	// Sort by play count descending
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].PlayCount > artists[j].PlayCount
	})
	
	// Take top 5
	top5Count := 5
	if len(artists) < top5Count {
		top5Count = len(artists)
	}
	
	result := make([]model.ArtistPlayCount, top5Count)
	for i := 0; i < top5Count; i++ {
		result[i] = model.ArtistPlayCount{
			ArtistID:   artists[i].ID,
			ArtistName: artists[i].Name,
			PlayCount:  artists[i].PlayCount,
		}
	}
	
	return result
}

func (qh *QueryHandler) enrichArtistNames(ctx context.Context, analytics *model.UserAnalytics, projection *store.AnalyticsProjection) {
	// Fetch artist names from content service for missing names
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}
	
	artistsURL := qh.config.ContentServiceURL + "/artists"
	artistsReq, err := http.NewRequestWithContext(ctx, "GET", artistsURL, nil)
	if err != nil {
		return
	}
	
	artistsResp, err := client.Do(artistsReq)
	if err != nil || artistsResp == nil || artistsResp.StatusCode != http.StatusOK {
		return
	}
	defer artistsResp.Body.Close()
	
	var allArtists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if json.NewDecoder(artistsResp.Body).Decode(&allArtists) != nil {
		return
	}
	
	// Update artist names in analytics
	artistNameMap := make(map[string]string)
	for _, artist := range allArtists {
		artistNameMap[artist.ID] = artist.Name
	}
	
	// Update all artist names - replace fallback names with real names
	for i := range analytics.Top5Artists {
		// Always try to get real name from content service if available
		if name, ok := artistNameMap[analytics.Top5Artists[i].ArtistID]; ok && name != "" {
			analytics.Top5Artists[i].ArtistName = name
		} else if analytics.Top5Artists[i].ArtistName == "" {
			// Only use fallback if we couldn't fetch the name
			analytics.Top5Artists[i].ArtistName = "Umetnik " + analytics.Top5Artists[i].ArtistID[:min(8, len(analytics.Top5Artists[i].ArtistID))] + "..."
		}
	}
}

func (qh *QueryHandler) updateSubscribedArtistsCount(ctx context.Context, analytics *model.UserAnalytics, userID string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
	}
	
	subsURL := qh.config.SubscriptionsServiceURL + "/subscriptions?userId=" + userID
	subsReq, err := http.NewRequestWithContext(ctx, "GET", subsURL, nil)
	if err != nil {
		return
	}
	
	subsResp, err := client.Do(subsReq)
	if err != nil || subsResp == nil || subsResp.StatusCode != http.StatusOK {
		return
	}
	defer subsResp.Body.Close()
	
	var subscriptions []struct {
		Type     string `json:"type"`
		ArtistID string `json:"artistId,omitempty"`
	}
	if json.NewDecoder(subsResp.Body).Decode(&subscriptions) != nil {
		return
	}
	
	count := 0
	uniqueArtists := make(map[string]bool)
	for _, sub := range subscriptions {
		if sub.Type == "artist" && sub.ArtistID != "" {
			if !uniqueArtists[sub.ArtistID] {
				uniqueArtists[sub.ArtistID] = true
				count++
			}
		}
	}
	
	analytics.SubscribedArtistsCount = count
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
