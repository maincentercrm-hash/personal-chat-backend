// pkg/configs/redis.go
package configs

import (
	"os"
	"strconv"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadRedisConfig() RedisConfig {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}
}
