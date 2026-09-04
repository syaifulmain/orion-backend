package server

import (
	"orion-backend/internal/auth"
	"orion-backend/internal/config"
	"orion-backend/internal/corpus"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// New initializes and returns a configured Fiber v3 application instance.
func New(cfg *config.Config, db *mongo.Database) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.ServerName,
	})

	// Global Middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowOrigins,
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	}))

	// Register Routes
	setupRoutes(app, cfg, db)

	return app
}

func setupRoutes(app *fiber.App, cfg *config.Config, db *mongo.Database) {
	api := app.Group("/api")
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Register Corpus routes if DB is connected
	if db != nil {
		corpusRepo := corpus.NewRepository(db)
		corpusService := corpus.NewService(corpusRepo)
		corpusHandler := corpus.NewHandler(corpusService)

		authMiddleware := auth.NewAuthMiddleware(cfg.CorpusAPISecret, cfg.KeyStorePath)

		corpus.RegisterRoutes(api, corpusHandler, authMiddleware)
	}
}
