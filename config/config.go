package config

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	ServerPort string `json:"server_port"`
	Dev        bool   `json:"dev"`
	ShardURL   string `json:"shard_url"`
	PgConfig   struct {
		Host string `json:"host"`
		Port string `json:"port"`
		User string `json:"user"`
		Pass string
		DB   string `json:"db"`
	} `json:"pg_config"`
	NATS struct {
		URLs []string `json:"urls"`
	} `json:"nats"`
	Shards []int `json:"shards"`
}

// ParseConfig of service
func ParseConfig(configPath, secretPath string) (*Config, error) {
	fileBody, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(fileBody, &cfg)
	if err != nil {
		return nil, err
	}

	pgPass, err := os.ReadFile(path.Join(secretPath, "postgres-password"))
	if err != nil {
		return nil, err
	}
	cfg.PgConfig.Pass = string(pgPass)

	return &cfg, nil
}
