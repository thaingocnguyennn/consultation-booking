// internal/database/database.go
package database

import (
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Connected to database successfully")
	return db, nil
}

func InitRedis(redisURL string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	log.Println("Connected to Redis successfully")
	return rdb
}
