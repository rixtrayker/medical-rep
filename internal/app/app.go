package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/cloudflare/tableflip"

	"github.com/rixtrayker/medical-rep/configs"
	"github.com/rixtrayker/medical-rep/internal/platform/database"
	"github.com/rixtrayker/medical-rep/internal/platform/logger"
	"github.com/rixtrayker/medical-rep/internal/platform/redis"
)

// App represents the main application
type App struct {
	config     *configs.Config
	logger     *logger.Logger
	router     *chi.Mux
	server     *http.Server
	health     gosundheit.Health
	db         *database.DB
	redis      *redis.Client
	upgrader   *tableflip.Upgrader
}

// Dependencies holds all application dependencies
type Dependencies struct {
	Config *configs.Config
	Logger *logger.Logger
	DB     *database.DB
	Redis  *redis.Client
	Health gosundheit.Health
}

// New creates a new application instance
func New() (*App, error) {
	// Load configuration
	if err := configs.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg := configs.Get()

	// Initialize logger
	logger, err := logger.New(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis
	redisClient, err := redis.New(cfg.Redis)
	if err != nil {
		logger.Error("Failed to initialize Redis", "error", err)
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize tableflip for zero-downtime deployments
	upgrader, err := tableflip.New(tableflip.Options{})
	if err != nil {
		logger.Error("Failed to create tableflip upgrader", "error", err)
		return nil, fmt.Errorf("failed to create tableflip upgrader: %w", err)
	}

	// Initialize health checker
	health := gosundheit.New()

	app := &App{
		config:   cfg,
		logger:   logger,
		db:       db,
		redis:    redisClient,
		health:   health,
		upgrader: upgrader,
	}

	// Setup router and server
	if err := app.setupRouter(); err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}

	if err := app.setupServer(); err != nil {
		return nil, fmt.Errorf("failed to setup server: %w", err)
	}

	// Setup health checks
	if err := app.setupHealthChecks(); err != nil {
		return nil, fmt.Errorf("failed to setup health checks: %w", err)
	}

	return app, nil
}

// setupRouter configures the HTTP router with middleware and routes
func (a *App) setupRouter() error {
	a.router = chi.NewRouter()

	// Basic middleware
	a.router.Use(middleware.RequestID)
	a.router.Use(middleware.RealIP)
	a.router.Use(middleware.Logger)
	a.router.Use(middleware.Recoverer)
	a.router.Use(middleware.Heartbeat("/ping"))

	// Timeout middleware
	a.router.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	a.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   a.config.HTTP.CORS.AllowedOrigins,
		AllowedMethods:   a.config.HTTP.CORS.AllowedMethods,
		AllowedHeaders:   a.config.HTTP.CORS.AllowedHeaders,
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting (if enabled)
	if a.config.HTTP.RateLimit.Enabled {
		// TODO: Implement rate limiting middleware
		a.logger.Info("Rate limiting is enabled but not implemented yet")
	}

	// Health check routes
	a.router.Mount("/health", a.health.Handler())
	a.router.Get("/healthz", a.healthzHandler)
	a.router.Get("/readiness", a.readinessHandler)
	a.router.Get("/liveness", a.livenessHandler)

	// API routes
	a.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// TODO: Add API routes here
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "Medical Rep API v1", "status": "ok"}`))
			})
		})
	})

	// Catch-all route
	a.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"message": "Welcome to %s", "version": "%s"}`, 
			a.config.App.Name, a.config.App.Version)))
	})

	return nil
}

// setupServer configures the HTTP server
func (a *App) setupServer() error {
	addr := fmt.Sprintf("%s:%d", a.config.HTTP.Host, a.config.HTTP.Port)

	a.server = &http.Server{
		Addr:           addr,
		Handler:        a.router,
		ReadTimeout:    a.config.HTTP.ReadTimeout,
		WriteTimeout:   a.config.HTTP.WriteTimeout,
		IdleTimeout:    a.config.HTTP.IdleTimeout,
		MaxHeaderBytes: a.config.HTTP.MaxHeaderBytes,
	}

	return nil
}

// setupHealthChecks configures health checks
func (a *App) setupHealthChecks() error {
	if !a.config.Health.Enabled {
		return nil
	}

	// Database health check
	if a.config.Health.DatabaseCheck && a.db != nil {
		dbCheck := checks.NewCustomCheck("database", func(ctx context.Context) (interface{}, error) {
			if err := a.db.Ping(ctx); err != nil {
				return nil, fmt.Errorf("database ping failed: %w", err)
			}
			return map[string]string{"status": "healthy"}, nil
		})

		if err := a.health.RegisterCheck(dbCheck,
			gosundheit.InitialDelay(2*time.Second),
			gosundheit.ExecutionPeriod(a.config.Health.CheckInterval),
		); err != nil {
			return fmt.Errorf("failed to register database health check: %w", err)
		}
	}

	// Redis health check
	if a.config.Health.RedisCheck && a.redis != nil {
		redisCheck := checks.NewCustomCheck("redis", func(ctx context.Context) (interface{}, error) {
			if err := a.redis.Ping(ctx); err != nil {
				return nil, fmt.Errorf("redis ping failed: %w", err)
			}
			return map[string]string{"status": "healthy"}, nil
		})

		if err := a.health.RegisterCheck(redisCheck,
			gosundheit.InitialDelay(2*time.Second),
			gosundheit.ExecutionPeriod(a.config.Health.CheckInterval),
		); err != nil {
			return fmt.Errorf("failed to register redis health check: %w", err)
		}
	}

	// External service health checks
	for _, url := range a.config.Health.ExternalChecks {
		httpCheck, err := checks.NewHTTPCheck(checks.HTTPCheckConfig{
			CheckName: fmt.Sprintf("http_%s", url),
			Timeout:   a.config.Health.Timeout,
			URL:       url,
		})
		if err != nil {
			return fmt.Errorf("failed to create HTTP health check for %s: %w", url, err)
		}

		if err := a.health.RegisterCheck(httpCheck,
			gosundheit.InitialDelay(5*time.Second),
			gosundheit.ExecutionPeriod(a.config.Health.CheckInterval),
		); err != nil {
			return fmt.Errorf("failed to register HTTP health check for %s: %w", url, err)
		}
	}

	return nil
}

// healthzHandler provides a simple health check endpoint for Kubernetes
func (a *App) healthzHandler(w http.ResponseWriter, r *http.Request) {
	results, healthy := a.health.Results()
	
	w.Header().Set("Content-Type", "application/json")
	
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": "unhealthy"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}
	
	a.logger.Debug("Health check", "results", results, "healthy", healthy)
}

// readinessHandler checks if the application is ready to serve traffic
func (a *App) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check critical dependencies
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ready := true
	checks := make(map[string]string)

	// Check database
	if a.db != nil {
		if err := a.db.Ping(ctx); err != nil {
			ready = false
			checks["database"] = "unhealthy"
		} else {
			checks["database"] = "healthy"
		}
	}

	// Check Redis
	if a.redis != nil {
		if err := a.redis.Ping(ctx); err != nil {
			ready = false
			checks["redis"] = "unhealthy"
		} else {
			checks["redis"] = "healthy"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	
	if !ready {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	response := map[string]interface{}{
		"ready":  ready,
		"checks": checks,
	}

	// Simple JSON encoding
	if ready {
		w.Write([]byte(`{"ready": true}`))
	} else {
		w.Write([]byte(`{"ready": false}`))
	}
}

// livenessHandler checks if the application is alive
func (a *App) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"alive": true}`))
}

// Run starts the application
func (a *App) Run() error {
	// Listen on the upgradeable socket
	addr := fmt.Sprintf("%s:%d", a.config.HTTP.Host, a.config.HTTP.Port)
	ln, err := a.upgrader.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	a.logger.Info("Starting server", 
		"addr", addr,
		"environment", a.config.App.Environment,
		"version", a.config.App.Version,
	)

	// Start the server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if a.config.HTTP.TLS.Enabled {
			errChan <- a.server.ServeTLS(ln, a.config.HTTP.TLS.CertFile, a.config.HTTP.TLS.KeyFile)
		} else {
			errChan <- a.server.Serve(ln)
		}
	}()

	// Tell tableflip that initialization is complete
	if err := a.upgrader.Ready(); err != nil {
		return fmt.Errorf("failed to signal ready: %w", err)
	}

	// Wait for shutdown signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigChan:
		a.logger.Info("Received shutdown signal", "signal", sig.String())
	case <-a.upgrader.Exit():
		a.logger.Info("Received upgrade signal")
	}

	return a.Shutdown()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	a.logger.Info("Shutting down application...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), a.config.App.Shutdown.Timeout)
	defer cancel()

	// Shutdown HTTP server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Server shutdown error", "error", err)
	}

	// Stop health checker
	if a.health != nil {
		a.health.DeregisterAll()
	}

	// Close database connections
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("Database close error", "error", err)
		}
	}

	// Close Redis connections
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			a.logger.Error("Redis close error", "error", err)
		}
	}

	// Stop upgrader
	if a.upgrader != nil {
		a.upgrader.Stop()
	}

	a.logger.Info("Application shutdown complete")
	return nil
}

// GetDependencies returns application dependencies for testing
func (a *App) GetDependencies() Dependencies {
	return Dependencies{
		Config: a.config,
		Logger: a.logger,
		DB:     a.db,
		Redis:  a.redis,
		Health: a.health,
	}
}