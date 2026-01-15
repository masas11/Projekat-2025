package store

import (
	"log"

	"github.com/gocql/gocql"
)

type CassandraStore struct {
	Session *gocql.Session
}

func NewCassandraStore(hosts []string, keyspace string) (*CassandraStore, error) {
	// First connect without keyspace to create it
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	// Create keyspace if it doesn't exist
	if err := session.Query(`
		CREATE KEYSPACE IF NOT EXISTS ` + keyspace + `
		WITH REPLICATION = {
			'class' : 'SimpleStrategy',
			'replication_factor' : 1
		}`).Exec(); err != nil {
		session.Close()
		return nil, err
	}

	// Close session and reconnect with keyspace
	session.Close()

	// Reconnect with keyspace
	cluster.Keyspace = keyspace
	session, err = cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	if err := session.Query(`
		CREATE TABLE IF NOT EXISTS notifications (
			id TEXT,
			user_id TEXT,
			type TEXT,
			message TEXT,
			content_id TEXT,
			read BOOLEAN,
			created_at TIMESTAMP,
			PRIMARY KEY (user_id, created_at, id)
		) WITH CLUSTERING ORDER BY (created_at DESC)`).Exec(); err != nil {
		session.Close()
		return nil, err
	}

	log.Println("Connected to Cassandra")
	return &CassandraStore{Session: session}, nil
}

func (s *CassandraStore) Close() {
	s.Session.Close()
}
