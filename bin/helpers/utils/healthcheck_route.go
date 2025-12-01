package utils

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterHealthCheckRoutes(app *fiber.App) {

	// Health check: service alive
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Ready check: app dependencies ready
	app.Get("/readyz", func(c *fiber.Ctx) error {
		// add DB/ping check here if needed
		return c.JSON(fiber.Map{
			"ready": true,
		})
	})
}
