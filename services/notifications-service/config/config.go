package config

import "os"
import "strings"

type Config struct {
	Port            string
	CassandraHosts  []string
	CassandraKeyspace string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	cassandraHostsStr := os.Getenv("CASSANDRA_HOSTS")
	if cassandraHostsStr == "" {
		cassandraHostsStr = "localhost:9042"
	}
	cassandraHosts := strings.Split(cassandraHostsStr, ",")
	for i, host := range cassandraHosts {
		cassandraHosts[i] = strings.TrimSpace(host)
	}

	keyspace := os.Getenv("CASSANDRA_KEYSPACE")
	if keyspace == "" {
		keyspace = "notifications_db"
	}

	return &Config{
		Port:              port,
		CassandraHosts:    cassandraHosts,
		CassandraKeyspace: keyspace,
	}
}
