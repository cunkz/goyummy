package main

import (
	"fmt"
	"log"

	recipe "github.com/cunkz/goyummy/bin/config"
)

func main() {
	cfg, err := recipe.LoadAuto()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("App Name:", cfg.App.Name)
	fmt.Println("Environment:", cfg.App.Environment)
	fmt.Println("Server running on:", cfg.Server.Host, cfg.Server.Port)
	fmt.Println("Log Level:", cfg.Logging.Level)

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
