package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alvor-technologies/iag-platform-go/authclient"
	platformotel "github.com/alvor-technologies/iag-platform-go/otel"

	"iag-inventory/backend/internal/config"
	"iag-inventory/backend/internal/db"
	"iag-inventory/backend/internal/handlers"
	"iag-inventory/backend/internal/middleware"
	"iag-inventory/backend/internal/migrate"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// OpenTelemetry → otel-collector:4317 (non-blocking dial).
	if tp, err := platformotel.Init(ctx, platformotel.Config{
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
	}); err != nil {
		log.Printf("otel disabled: %v", err)
	} else {
		defer func() {
			sc, c := context.WithTimeout(context.Background(), 5*time.Second)
			defer c()
			_ = tp.Shutdown(sc)
		}()
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := migrate.Up(ctx, pool); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	var verifier *authclient.Verifier
	if cfg.AuthMode == "jwt" {
		verifier = authclient.NewVerifier(authclient.Options{
			JWKSURL:  cfg.JWKSURL,
			Issuer:   cfg.JWTIssuer,
			Audience: cfg.Audience,
		})
		initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := verifier.Refresh(initCtx); err != nil {
			cancel()
			log.Fatalf("jwks refresh: %v", err)
		}
		cancel()
		go jwksRefreshLoop(verifier)
	}

	platformAuth := middleware.NewPlatformAuth(middleware.PlatformAuthOptions{
		Mode:     cfg.AuthMode,
		Verifier: verifier,
	})

	go registerPermissionsLoop(ctx, cfg)

	router := handlers.NewRouter(handlers.RouterDeps{
		Cfg:          cfg,
		Pool:         pool,
		PlatformAuth: platformAuth,
		StrictRBAC:   cfg.StrictRBAC(),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("inventory listening on :%s (aud=%s)", cfg.Port, cfg.Audience)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func jwksRefreshLoop(v *authclient.Verifier) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := v.Refresh(ctx); err != nil {
			log.Printf("jwks refresh: %v", err)
		}
		cancel()
	}
}
