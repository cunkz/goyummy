package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/cunkz/goyummy/bin/config"
	"github.com/cunkz/goyummy/bin/helpers/db"
	"github.com/cunkz/goyummy/bin/helpers/utils"
	"github.com/cunkz/goyummy/bin/middleware"

	handlers "github.com/cunkz/goyummy/bin/modules/default"
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

	// Add route based on configuration
	for _, mod := range cfg.Modules {
		log.Info().Msg(mod.Name)
		for _, op := range mod.Operations {
			log.Info().Msg(op)
			module := utils.ToSlug(mod.Name)
			baseRoute := fmt.Sprintf("/api/%s/v1", module)

			switch strings.ToLower(op) {
			case "create":
				app.Post(baseRoute, handlers.CreateDefaultHanlder)
				log.Info().Msgf("Add Route POST %s", baseRoute)
			case "read_list":
				app.Get(baseRoute, handlers.ReadListDefaultHanlder)
				log.Info().Msgf("Add Route GET %s", baseRoute)
			case "read_single":
				app.Get(baseRoute+"/:id", handlers.ReadSingleDefaultHanlder)
				log.Info().Msgf("Add Route GET %s", baseRoute+"/:id")
			case "update":
				app.Patch(baseRoute+"/:id", handlers.UpdateDefaultHanlder)
				log.Info().Msgf("Add Route PATCH %s", baseRoute+"/:id")
			case "delete":
				app.Delete(baseRoute+"/:id", handlers.DeleteDefaultHanlder)
				log.Info().Msgf("Add Route DELETE %s", baseRoute+"/:id")
			default:
				log.Info().Msgf("Invalid Operation for Module: %s", mod.Name)
			}
		}
	}

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
