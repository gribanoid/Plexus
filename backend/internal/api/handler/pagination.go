package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/plexus/backend/internal/repository"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 100
)

func parsePageParams(c *fiber.Ctx) (repository.PageParams, error) {
	limit := defaultPageLimit
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return repository.PageParams{}, fiber.NewError(fiber.StatusBadRequest, "invalid limit")
		}
		if n > maxPageLimit {
			n = maxPageLimit
		}
		limit = n
	}
	return repository.PageParams{
		Limit:  limit,
		Cursor: c.Query("cursor"),
	}, nil
}

func pageResponse[T any, D any](result repository.PageResult[T], mapFn func(T) D) fiber.Map {
	resp := fiber.Map{
		"items": mapSlice(result.Items, mapFn),
	}
	if result.NextCursor != nil {
		resp["next_cursor"] = *result.NextCursor
	}
	return resp
}
