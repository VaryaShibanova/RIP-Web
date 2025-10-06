package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost   string
	ServicePort   int
	JWTSecret     string
	JWTExpiration int
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	// Устанавливаем значения по умолчанию
	viper.SetDefault("ServiceHost", "localhost")
	viper.SetDefault("ServicePort", 8080)
	viper.SetDefault("JWTSecret", "fallback-secret-key")
	viper.SetDefault("JWTExpiration", 24)
	viper.SetDefault("RedisHost", "localhost")
	viper.SetDefault("RedisPort", 6379)
	viper.SetDefault("RedisPassword", "")
	viper.SetDefault("RedisDB", 0)

	err = viper.ReadInConfig()
	if err != nil {
		log.Warnf("No config file found, using defaults: %v", err)
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}
