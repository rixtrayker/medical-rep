package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"os"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/cloudflare/tableflip"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rixtrayker/medical-rep/internal/app"
)

func main() {
	// Initialize tableflip for zero-downtime deployments
	upg, err := tableflip.New(tableflip.Options{})
	// Create and initialize the application
	application, err := app.New()
	if err != nil {
		log.Fatal("Failed to create tableflip upgrader:", err)
		log.Fatal("Failed to create application:", err)
	}
	defer upg.Stop()

	// Listen on the upgradeable socket
	ln, err := upg.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Initialize health checker
	h := gosundheit.New()

	// Add a simple HTTP check (checking our own server)
	httpCheck, err := checks.NewHTTPCheck(checks.HTTPCheckConfig{
		CheckName: "http_check",
		Timeout:   1 * time.Second,
		URL:       "http://localhost:8080/ping",
	})
	if err != nil {
		log.Fatal("Failed to create HTTP health check:", err)
	}

	// Add a custom check example
	customCheck := checks.NewCustomCheck("custom_check", func(ctx context.Context) (details interface{}, err error) {
		// Add your custom health check logic here
		// For example, check database connectivity, external services, etc.
		return map[string]string{"status": "healthy", "timestamp": time.Now().Format(time.RFC3339)}, nil
	})

	// Register health checks
	err = h.RegisterCheck(httpCheck, gosundheit.InitialDelay(2*time.Second), gosundheit.ExecutionPeriod(10*time.Second))
	if err != nil {
		log.Fatal("Failed to register HTTP health check:", err)
	// Run the application
	if err := application.Run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}

	err = h.RegisterCheck(customCheck, gosundheit.InitialDelay(1*time.Second), gosundheit.ExecutionPeriod(5*time.Second))
	if err != nil {
		log.Fatal("Failed to register custom health check:", err)
	}

	// Create router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Heartbeat("/ping"))

	// Routes
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	})

	// Health check endpoints
	router.Mount("/health", h.Handler())

	// Additional health endpoints for convenience
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		results, healthy := h.Results()
		if !healthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		// Simple response for k8s/docker health checks
		if healthy {
			w.Write([]byte("OK"))
		} else {
			w.Write([]byte("UNHEALTHY"))
		}

		// Log detailed results
		log.Printf("Health check results: %+v, healthy: %v", results, healthy)
	})

	// Create server
	server := &http.Server{
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on %s", ln.Addr())
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Tell tableflip that initialization is complete
	if err := upg.Ready(); err != nil {
		log.Fatal("Failed to signal ready:", err)
	}

	// Wait for upgrade signal or termination
	<-upg.Exit()

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Stop health checker
	h.DeregisterAll()
	log.Println("Server stopped")
}
