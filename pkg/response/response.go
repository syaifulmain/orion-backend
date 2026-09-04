package response

import (
	"github.com/gofiber/fiber/v3"
)

// Pagination holds pagination metadata.
type Pagination struct {
	Total   int64 `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	HasMore bool  `json:"has_more"`
}

// PaginatedResponse defines standardized paginated payload.
type PaginatedResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// SuccessResponse defines standardized single or custom response payload.
type SuccessResponse[T any] struct {
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
}

// ErrorDetail describes the error status and message.
type ErrorDetail struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ErrorResponse defines standardized error payload.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// JSONPaginated sends a 200 OK paginated JSON response matching the required schema.
func JSONPaginated[T any](c fiber.Ctx, data []T, total int64, limit, offset int) error {
	if data == nil {
		data = []T{}
	}
	hasMore := int64(offset+len(data)) < total

	return c.Status(fiber.StatusOK).JSON(PaginatedResponse[T]{
		Data: data,
		Pagination: Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	})
}

// JSONSuccess sends a standardized success JSON response.
func JSONSuccess[T any](c fiber.Ctx, statusCode int, data T, message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	return c.Status(statusCode).JSON(SuccessResponse[T]{
		Message: msg,
		Data:    data,
	})
}

// JSONError sends a standardized error JSON response.
func JSONError(c fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Error: ErrorDetail{
			Status:  statusCode,
			Message: message,
		},
	})
}
