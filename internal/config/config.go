package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Env        string      `yaml:"env" env:"ENV" env-default:"local"`
		PostgreURL postgreURL  `yaml:"postgres"`
		AppInfo    appStruct   `yaml:"app"`
		GRPC       grpcStruct  `yaml:"grpc"`
		Kafka      KafkaConfig `yaml:"kafka"`
	}

	appStruct struct {
		Name    string `yaml:"name" env-required:"true"`
		Version string `yaml:"version" env-required:"true"`
	}

	grpcStruct struct {
		Address string `yaml:"address" env-required:"true"`
	}

	postgreURL struct {
		URL       string `yaml:"url" env-required:"true"`
		Host      string `yaml:"host" env-required:"true"`
		Port      uint16 `yaml:"port" env-required:"true"`
		Database  string `yaml:"database" env-required:"true"`
		User      string `yaml:"user" env-required:"true"`
		Password  string `yaml:"password" env-required:"true"`
		PgPoolMax uint16 `yaml:"pg_pool_max" env-required:"true"`
	}

	KafkaConfig struct {
		Brokers      []string      `yaml:"brokers"`
		TopicRatings string        `yaml:"topic_ratings"`
		GroupID      string        `yaml:"group_id"`
		DialTimeout  time.Duration `yaml:"dial_timeout"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		MinBytes     int           `yaml:"min_bytes"`
		MaxBytes     int           `yaml:"max_bytes"`
	}
)

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var config Config

	err := cleanenv.ReadConfig(configPath, &config)
	if err != nil {
		log.Fatal("Cant read config", err)
	}

	return &config
}
