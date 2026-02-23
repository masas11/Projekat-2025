package store

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"recommendation-service/config"
)

type Neo4jStore struct {
	driver neo4j.DriverWithContext
	ctx    context.Context
}

func NewNeo4jStore(cfg *config.Config) (*Neo4jStore, error) {
	ctx := context.Background()
	driver, err := neo4j.NewDriverWithContext(
		cfg.Neo4jURI,
		neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	log.Println("Connected to Neo4j successfully")

	store := &Neo4jStore{
		driver: driver,
		ctx:    ctx,
	}

	// Initialize schema
	if err := store.InitializeSchema(); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func (s *Neo4jStore) InitializeSchema() error {
	session := s.driver.NewSession(s.ctx, neo4j.SessionConfig{})
	defer session.Close(s.ctx)

	// Create constraints and indexes
	queries := []string{
		"CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE",
		"CREATE CONSTRAINT artist_id IF NOT EXISTS FOR (a:Artist) REQUIRE a.id IS UNIQUE",
		"CREATE CONSTRAINT song_id IF NOT EXISTS FOR (s:Song) REQUIRE s.id IS UNIQUE",
		"CREATE CONSTRAINT genre_name IF NOT EXISTS FOR (g:Genre) REQUIRE g.name IS UNIQUE",
		"CREATE INDEX user_id_index IF NOT EXISTS FOR (u:User) ON (u.id)",
		"CREATE INDEX artist_id_index IF NOT EXISTS FOR (a:Artist) ON (a.id)",
		"CREATE INDEX song_id_index IF NOT EXISTS FOR (s:Song) ON (s.id)",
		"CREATE INDEX genre_name_index IF NOT EXISTS FOR (g:Genre) ON (g.name)",
	}

	for _, query := range queries {
		_, err := session.Run(s.ctx, query, nil)
		if err != nil {
			log.Printf("Warning: Failed to execute schema query: %s, error: %v", query, err)
			// Continue even if some constraints already exist
		}
	}

	log.Println("Neo4j schema initialized")
	return nil
}

func (s *Neo4jStore) Close() error {
	return s.driver.Close(s.ctx)
}
