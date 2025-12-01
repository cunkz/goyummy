package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	recipe "github.com/cunkz/goyummy/bin/config"
	logger "github.com/cunkz/goyummy/bin/helpers/utils"
)

func main() {
	cfg, _ := recipe.LoadAuto()

	// Initialize logger
	logger.InitLogger(cfg)
	log.Info().Msg("Init Logger")

	fmt.Println("App Name:", cfg.App.Name)
	fmt.Println("Environment:", cfg.App.Environment)
	fmt.Println("Server running on:", cfg.Server.Host, cfg.Server.Port)
	fmt.Println("Log Level:", cfg.Logging.Level)
	fmt.Println("Log Output:", cfg.Logging.Output)

	// Use the database config
	for _, db := range cfg.Databases {
		fmt.Println("Database:", db.Name)
		fmt.Println(" Engine:", db.Engine)
		fmt.Println(" URI:", db.URI)
	}

	// Use module config
	for _, mod := range cfg.Modules {
		fmt.Println("\nModule:", mod.Name)
		fmt.Println(" Table:", mod.Table)
		fmt.Println(" Operations:", mod.Operations)
	}
}
