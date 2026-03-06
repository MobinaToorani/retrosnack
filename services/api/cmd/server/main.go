package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aidantrabs/kenko"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MobinaToorani/retrosnack/internal/auth"
	"github.com/MobinaToorani/retrosnack/internal/catalog"
	"github.com/MobinaToorani/retrosnack/internal/instagram"
	"github.com/MobinaToorani/retrosnack/internal/inventory"
	"github.com/MobinaToorani/retrosnack/internal/media"
	"github.com/MobinaToorani/retrosnack/internal/orders"
	"github.com/MobinaToorani/retrosnack/internal/payments"
	"github.com/MobinaToorani/retrosnack/pkg/config"
	"github.com/MobinaToorani/retrosnack/pkg/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("database ping failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	// Wire domain modules
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	catalogRepo := catalog.NewRepository(pool)
	catalogSvc := catalog.NewService(catalogRepo)
	catalogHandler := catalog.NewHandler(catalogSvc, cfg.JWTSecret)

	inventoryRepo := inventory.NewRepository(pool)
	inventorySvc := inventory.NewService(inventoryRepo)
	inventoryHandler := inventory.NewHandler(inventorySvc)

	ordersRepo := orders.NewRepository(pool)
	ordersSvc := orders.NewService(ordersRepo, inventorySvc)
	ordersHandler := orders.NewHandler(ordersSvc, cfg.JWTSecret)

	paymentsSvc := payments.NewService(ordersSvc, cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	paymentsHandler := payments.NewHandler(paymentsSvc)

	instagramRepo := instagram.NewRepository(pool)
	instagramSvc := instagram.NewService(instagramRepo)
	instagramHandler := instagram.NewHandler(instagramSvc, cfg.JWTSecret)

	mediaSvc := media.NewService(cfg)
	mediaHandler := media.NewHandler(mediaSvc, cfg.JWTSecret)

	// Health monitoring
	k, err := kenko.New(
		kenko.WithTarget("stripe", "https://api.stripe.com/v1/"),
		kenko.WithInterval(30*time.Second),
		kenko.WithLogger(logger),
	)
	if err != nil {
		logger.Error("failed to initialize health monitor", "error", err)
		os.Exit(1)
	}
	go k.Run(context.Background())

	// Router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.CORS(cfg.Env))

	checker := k.Checker()
	r.Get("/health", kenko.HandleHealth(checker))
	r.Get("/ready", kenko.HandleReady(checker))
	r.Get("/status", kenko.HandleStatus(checker))

	r.Route("/api", func(r chi.Router) {
		authHandler.Register(r)
		catalogHandler.Register(r)
		inventoryHandler.Register(r)
		ordersHandler.Register(r)
		paymentsHandler.Register(r)
		instagramHandler.Register(r)
		mediaHandler.Register(r)
	})

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", "addr", addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	logger.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
