package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pollingapp/backend/config"
	"pollingapp/backend/handlers"
)

func main() {
	godotenv.Load()

	config.ConnectMongo()
	config.ConnectRedis()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := router.Group("/api")
	{
		api.POST("/auth/signup", handlers.Signup)
		api.POST("/auth/login", handlers.Login)

		api.POST("/polls", handlers.AuthMiddleware(), handlers.CreatePoll)
		api.GET("/polls/:id", handlers.GetPoll)
		api.POST("/polls/:id/vote", handlers.Vote)
	}

	router.GET("/ws/polls/:id", handlers.PollSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("server running on port " + port)
	router.Run(":" + port)
}
