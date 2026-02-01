package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/app"
	"github.com/dhruvsaxena1998/splitplus/internal/config"
	"github.com/dhruvsaxena1998/splitplus/internal/db"
	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to database
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	// Initialize app with JWT configuration
	application := app.New(pool, queries, cfg.JWTSecret, cfg.AccessTokenExpiry, cfg.RefreshTokenExpiry)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      application.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("server running on :8080")
	log.Fatal(server.ListenAndServe())
}
