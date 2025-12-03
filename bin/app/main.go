package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/cunkz/goyummy/bin/config"
	"github.com/cunkz/goyummy/bin/helpers/db"
	"github.com/cunkz/goyummy/bin/helpers/utils"
	"github.com/cunkz/goyummy/bin/middleware"

	"github.com/cunkz/goyummy/bin/modules"
)

func main() {
	cfg, _ := config.LoadAuto()

	// Initialize logger
	utils.InitLogger(cfg)
	log.Info().Msg("Init Logger")

	// Create Fiber instance
	app := fiber.New()

	// Add request logging middleware
	app.Use(middleware.RequestLogger())

	if err := db.InitDatabases(cfg); err != nil {
		log.Error().Err(err).Msg("error init databases")
	}

	// Add Ready and Health check Route
	utils.RegisterHealthCheckRoutes(app)

	// Register routes and controllers for each module
	modules.RegisterModules(app, cfg.Modules)

	// Network check
	ln := utils.NetCheck(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	// Start Fiber
	go func() {
		log.Info().Msgf("🚀 Server is ready at http://%s:%s", cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
		if err := app.Listener(ln); err != nil {
			log.Error().Err(err).Msg("server stopped")
		}
	}()

	// Graceful shutdown
	utils.GracefulShutdown(app)
}
