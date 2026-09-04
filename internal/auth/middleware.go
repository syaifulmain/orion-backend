package auth

import (
	"strings"

	"orion-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

// NewAuthMiddleware returns a Fiber middleware that validates incoming Bearer / X-API-Key.
func NewAuthMiddleware(secret, keyStorePath string) fiber.Handler {
	return func(c fiber.Ctx) error {
		apiKey := ""

		// 1. Check Authorization header (Bearer <key>)
		authHeader := strings.TrimSpace(c.Get("Authorization"))
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				apiKey = strings.TrimSpace(parts[1])
			}
		}

		// 2. Fallback to X-API-Key header
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.Get("X-API-Key"))
		}

		if apiKey == "" {
			c.Set("WWW-Authenticate", "Bearer")
			return response.JSONError(c, fiber.StatusUnauthorized, "Invalid or Expired API Key")
		}

		claims, err := ValidateAPIKey(apiKey, secret, keyStorePath)
		if err != nil {
			c.Set("WWW-Authenticate", "Bearer")
			return response.JSONError(c, fiber.StatusUnauthorized, "Invalid or Expired API Key")
		}

		// Store claims in context for handlers if needed
		c.Locals("apiKeyClaims", claims)
		return c.Next()
	}
}
