package handlers

import (
	"errors"
	"strconv"
	"strings"

	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type coverImagePayload struct {
	Title       string `json:"title" form:"title"`
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Type        string `json:"type" form:"type"`
	CoverImage  string `json:"cover_image" form:"cover_image"`
	Order       int    `json:"order" form:"order"`
}

func parseCoverImagePayload(c *fiber.Ctx, payload *coverImagePayload) error {
	if !isMultipartForm(c) {
		return c.BodyParser(payload)
	}

	payload.Title = c.FormValue("title")
	payload.Name = c.FormValue("name")
	payload.Description = c.FormValue("description")
	payload.Type = c.FormValue("type")
	payload.CoverImage = c.FormValue("cover_image")
	if order := c.FormValue("order"); order != "" {
		parsedOrder, err := strconv.Atoi(order)
		if err != nil {
			return errors.New("order must be a number")
		}
		payload.Order = parsedOrder
	}

	coverImage, err := uploadOptionalCoverImage(c)
	if err != nil {
		return err
	}
	if coverImage != "" {
		payload.CoverImage = coverImage
	}

	return nil
}

func isMultipartForm(c *fiber.Ctx) bool {
	contentType := strings.ToLower(c.Get(fiber.HeaderContentType))
	return strings.HasPrefix(contentType, "multipart/form-data")
}

func uploadOptionalCoverImage(c *fiber.Ctx) (string, error) {
	for _, field := range []string{"cover_image", "cover"} {
		file, err := c.FormFile(field)
		if err == nil && file != nil {
			return utils.SaveCoverImage(file)
		}
	}

	return "", nil
}
