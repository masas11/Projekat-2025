package config

import "os"

type Config struct {
	Port              string
	CassandraHosts    string
	CassandraKeyspace string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	cassandraHosts := os.Getenv("CASSANDRA_HOSTS")
	if cassandraHosts == "" {
		cassandraHosts = "localhost:9042"
	}

	cassandraKeyspace := os.Getenv("CASSANDRA_KEYSPACE")
	if cassandraKeyspace == "" {
		cassandraKeyspace = "notifications_keyspace"
	}

	return &Config{
		Port:              port,
		CassandraHosts:    cassandraHosts,
		CassandraKeyspace: cassandraKeyspace,
	}
}
