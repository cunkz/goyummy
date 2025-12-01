package main

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	recipe "github.com/cunkz/goyummy/bin/config"
	logger "github.com/cunkz/goyummy/bin/helpers/utils"
	middleware "github.com/cunkz/goyummy/bin/middleware"
)

func main() {
	cfg, _ := recipe.LoadAuto()

	// Initialize logger
	logger.InitLogger(cfg)
	log.Info().Msg("Init Logger")

	// Create Fiber instance
	app := fiber.New()

	// Add request logging middleware
	app.Use(middleware.RequestLogger())

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

	// Combine host + port
	address := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)

	// Bind manually so we can log AFTER server is ready
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to bind port")
	}

	log.Info().Msgf("🚀 Server is ready at http://%s:%s", cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	// Start Fiber
	go func() {
		if err := app.Listener(ln); err != nil {
			log.Error().Err(err).Msg("server stopped")
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Warn().Msg("🛑 Stopping server...")
	_ = app.Shutdown()
	log.Info().Msg("✔ Shutdown complete")
}
