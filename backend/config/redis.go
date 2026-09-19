package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	url := os.Getenv("REDIS_URL")

	var opts *redis.Options
	var err error

	if url != "" {
		opts, err = redis.ParseURL(url)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		opts = &redis.Options{Addr: "localhost:6379"}
	}

	RedisClient = redis.NewClient(opts)

	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("redis connected")
}
