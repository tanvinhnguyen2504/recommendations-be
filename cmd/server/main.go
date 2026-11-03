package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/nguyentanvinh/recommendations-be/backend/docs/swagger"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/config"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/db"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/middleware"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/router"
)

// @title			Recommendations API
// @version		1.0
// @description	REST API over the gm_places dataset: food and drink venues
// @description	collected by a scraper. Lists are keyset paged via metadata.next_cursor.
// @host			localhost:9998
// @BasePath		/api/v1
// @schemes		http https
// @accept			json
// @produce		json
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := config.Load()

	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.Connect(ctx, config)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	log.Printf("database connected (%s)", config.RedactedDatabaseTarget())

	r := &router.Router{
		Config: config,
		Pool:   pool,
	}

	router := r.NewRouter()

	cors := middleware.NewCORS(*config)

	server := &http.Server{
		Addr:              config.ServerPort(),
		Handler:           cors.Handler(router),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("listening on %s (env=%s image=%s)", config.ServerPort(), config.Env, config.Image)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
