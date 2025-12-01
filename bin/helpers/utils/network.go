package utils

import (
	"net"

	"github.com/rs/zerolog/log"
)

func NetCheck(host string, port string) net.Listener {
	// Combine host + port
	address := host + ":" + port

	// Bind manually so we can log AFTER server is ready
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to bind port")
	}

	return ln
}
