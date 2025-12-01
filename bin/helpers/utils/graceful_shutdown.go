package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func GracefulShutdown(app *fiber.App) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Warn().Msg("🛑 Stopping server...")
	_ = app.Shutdown()
	log.Info().Msg("✔ Shutdown complete")
}
