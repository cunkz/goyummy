package auth

import (
	"encoding/base64"
	"fmt"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"

	"github.com/rs/zerolog/log"

	"github.com/cunkz/goyummy/bin/config"
)

func BuildAuthMap(cfg *config.AppConfig) (map[string]fiber.Handler, error) {
	log.Info().Msg("Build Auth")
	m := make(map[string]fiber.Handler)

	for _, a := range cfg.Auths {
		switch a.Type {

		case "basic":
			m[a.Name] = basicauth.New(basicauth.Config{
				Users: map[string]string{
					a.BasicUsername: a.BasicPassword,
				},
				// Optional: Customize the unauthorized response
				Unauthorized: func(c *fiber.Ctx) error {
					return c.Status(fiber.StatusUnauthorized).SendString("Custom Unauthorized Message")
				},
			})

		case "jwt":
			// decode base64 PEM
			pemBytes, err := base64.StdEncoding.DecodeString(a.JWTPubKey64)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 public key for %s: %w", a.Name, err)
			}

			// parse PEM → *rsa.PublicKey
			pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
			if err != nil {
				return nil, fmt.Errorf("invalid RSA public key PEM for %s: %w", a.Name, err)
			}

			m[a.Name] = jwtware.New(jwtware.Config{
				SigningKey: jwtware.SigningKey{
					JWTAlg: jwtware.RS256,
					Key:    pubKey,
				},
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					fmt.Print(err)
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "invalid or missing token",
					})
				},
			})

		default:
			return nil, fmt.Errorf("unsupported auth type: %s", a.Type)
		}
	}

	return m, nil
}
