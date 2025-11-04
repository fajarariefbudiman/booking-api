package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	loadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, db := InitMongo()
	defer func() {
		_ = client.Disconnect(ctxBackground())
	}()

	RunSeeder(ctx, db)
	r := gin.Default()
	r.Use(JSONContentTypeMiddleware())

	handlers := NewHandlers(db)

	SetupRoutes(r, handlers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
