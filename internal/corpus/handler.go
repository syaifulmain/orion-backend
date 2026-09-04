package corpus

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"

	"orion-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c fiber.Ctx) error {
	var params FilterParams
	if err := c.Bind().Query(&params); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid query parameters: "+err.Error())
	}

	records, total, validParams, err := h.service.GetCorpusList(c.Context(), params)
	if err != nil {
		if errors.Is(err, ErrInvalidSplit) || errors.Is(err, ErrQueryTooLong) {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error())
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve corpus: "+err.Error())
	}

	return response.JSONPaginated(c, records, total, validParams.Limit, validParams.Offset)
}

func (h *Handler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.service.GetCorpusByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.JSONError(c, fiber.StatusNotFound, "Corpus record not found")
		}
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to retrieve corpus: "+err.Error())
	}

	return response.JSONSuccess(c, fiber.StatusOK, item)
}

func (h *Handler) Export(c fiber.Ctx) error {
	// Reject pagination parameters
	queries := c.Queries()
	if _, hasLimit := queries["limit"]; hasLimit {
		return response.JSONError(c, fiber.StatusBadRequest, "limit parameter is not allowed for export")
	}
	if _, hasOffset := queries["offset"]; hasOffset {
		return response.JSONError(c, fiber.StatusBadRequest, "offset parameter is not allowed for export")
	}

	var params FilterParams
	if err := c.Bind().Query(&params); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid query parameters: "+err.Error())
	}

	validated, err := h.service.ValidateFilter(params)
	if err != nil {
		if errors.Is(err, ErrInvalidSplit) || errors.Is(err, ErrQueryTooLong) {
			return response.JSONError(c, fiber.StatusBadRequest, err.Error())
		}
		return response.JSONError(c, fiber.StatusBadRequest, err.Error())
	}

	c.Set("Content-Type", "application/x-ndjson; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="corpus-export.jsonl"`)
	c.Set("Cache-Control", "no-store")
	c.Set("X-Content-Type-Options", "nosniff")

	return c.SendStreamWriter(func(sw *bufio.Writer) {
		_ = h.service.ExportCorpusStream(context.Background(), validated, func(record *Corpus) error {
			data, err := json.Marshal(record)
			if err != nil {
				return err
			}
			if _, err := sw.Write(data); err != nil {
				return err
			}
			if err := sw.WriteByte('\n'); err != nil {
				return err
			}
			return sw.Flush()
		})
	})
}
