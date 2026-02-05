package store

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

type CassandraStore struct {
	Session *gocql.Session
}

func NewCassandraStore(hosts string, keyspace string) (*CassandraStore, error) {
	// Parse hosts (can be comma-separated)
	hostList := strings.Split(hosts, ",")
	for i := range hostList {
		hostList[i] = strings.TrimSpace(hostList[i])
	}

	// Create cluster configuration
	cluster := gocql.NewCluster(hostList...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	// Create session
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cassandra session: %w", err)
	}

	// Test connection
	if err := session.Query("SELECT now() FROM system.local").Exec(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to test Cassandra connection: %w", err)
	}

	log.Println("Connected to Cassandra successfully")

	return &CassandraStore{
		Session: session,
	}, nil
}

func (s *CassandraStore) Close() {
	if s.Session != nil {
		s.Session.Close()
	}
}

// InitKeyspace creates the keyspace if it doesn't exist
func InitKeyspace(hosts string, keyspace string) error {
	hostList := strings.Split(hosts, ",")
	for i := range hostList {
		hostList[i] = strings.TrimSpace(hostList[i])
	}

	// First, connect without keyspace to create it
	cluster := gocql.NewCluster(hostList...)
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session for keyspace initialization: %w", err)
	}

	// Create keyspace if not exists
	createKeyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		keyspace,
	)
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		session.Close()
		return fmt.Errorf("failed to create keyspace: %w", err)
	}

	log.Printf("Keyspace '%s' created or already exists", keyspace)
	session.Close()

	// Now reconnect with keyspace to create table
	clusterWithKeyspace := gocql.NewCluster(hostList...)
	clusterWithKeyspace.Keyspace = keyspace
	clusterWithKeyspace.Consistency = gocql.One
	clusterWithKeyspace.Timeout = 10 * time.Second
	clusterWithKeyspace.ConnectTimeout = 10 * time.Second

	sessionWithKeyspace, err := clusterWithKeyspace.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to reconnect with keyspace: %w", err)
	}
	defer sessionWithKeyspace.Close()

	createTableQuery := `
		CREATE TABLE IF NOT EXISTS notifications (
			id text,
			user_id text,
			type text,
			message text,
			content_id text,
			read boolean,
			created_at timestamp,
			PRIMARY KEY (user_id, created_at, id)
		) WITH CLUSTERING ORDER BY (created_at DESC)
	`

	if err := sessionWithKeyspace.Query(createTableQuery).Exec(); err != nil {
		return fmt.Errorf("failed to create notifications table: %w", err)
	}

	log.Println("Notifications table created or already exists")
	return nil
}
