package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	Name       string `yaml:"name"`
	JWTSecret  string `yaml:"jwt_secret"`
	ServerPort string `yaml:"server_port"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	cfg.Host = getEnv("HOST", cfg.Host, "localhost")
	cfg.Port = getEnv("PORT", cfg.Port, "5432")
	cfg.User = getEnv("DB_USER", cfg.User, "postgres")
	cfg.Password = getEnv("PASSWORD", cfg.Password, "admin")
	cfg.Name = getEnv("NAME", cfg.Name, "auth_service")
	cfg.JWTSecret = getEnv("JWT_SECRET", cfg.JWTSecret, "")
	cfg.ServerPort = getEnv("SERVER_PORT", cfg.ServerPort, "8081")

	return cfg, nil
}

func getEnv(key, current, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}
	if current != "" {
		return current
	}
	return defaultValue
}
