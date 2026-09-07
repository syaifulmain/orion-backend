package main

import (
	"context"
	"fmt"
	"log"

	"orion-backend/internal/auth"
	"orion-backend/internal/config"
	"orion-backend/internal/database"
	"orion-backend/internal/server"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// @title Orion Backend API
// @version 1.0
// @description High-performance Indonesian speech & text corpus repository and catalog API backend.
//
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Pass your HMAC-SHA256 API key prefixed by Bearer (e.g. 'Bearer cps_xxx').

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description Alternative header passing your HMAC-SHA256 API key directly (e.g. 'cps_xxx').
func main() {
	// Load configuration from environment variables / .env
	cfg := config.LoadConfig()

	// Validate mandatory authentication configuration
	if err := auth.ValidateSecret(cfg.CorpusAPISecret); err != nil {
		log.Fatalf("Fatal: Authentication is mandatory. %v", err)
	}

	// Connect to MongoDB
	var db *mongo.Database
	client, err := database.ConnectMongoDB(cfg.MongoDBURI)
	if err != nil {
		log.Printf("Warning: Failed to connect to MongoDB: %v. Corpus routes will not be mounted.", err)
	} else {
		log.Println("Connected to MongoDB successfully!")
		defer func() {
			if err := client.Disconnect(context.Background()); err != nil {
				log.Printf("Error disconnecting MongoDB: %v", err)
			}
		}()
		db = client.Database("orion_db")
	}

	// Initialize Fiber v3 application
	app := server.New(cfg, db)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting %s in [%s] mode on port %s...", cfg.ServerName, cfg.ServerEnv, cfg.ServerPort)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
