package corpus

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers corpus endpoints to the provided router with optional middlewares.
func RegisterRoutes(router fiber.Router, handler *Handler, middlewares ...any) {
	group := router.Group("/corpus", middlewares...)

	group.Get("/export", handler.Export)
	group.Get("/:id", handler.GetByID)
	group.Get("/", handler.List)
}
