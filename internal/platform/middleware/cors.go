package middleware

import (
	"net/http"

	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/config"
	"github.com/rs/cors"
)

func NewCORS(config config.Config) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{config.FrontendURL},
		AllowedHeaders: []string{
			"Accept",
			"Content-Type",
			"Origin",
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowCredentials: true,
	})
}
