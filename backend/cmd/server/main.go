package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"tathya-avalokan/backend/internal/config"
	"tathya-avalokan/backend/internal/crypto"
	"tathya-avalokan/backend/internal/database"
	"tathya-avalokan/backend/internal/handlers"
	"tathya-avalokan/backend/internal/middleware"
	"tathya-avalokan/backend/internal/proxy"
	"tathya-avalokan/backend/internal/repository"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize encryption cipher with key or dev fallback
	crypto.InitCipherKey(cfg.EncryptionKey, cfg.AppSecretKey)

	log.Printf("Connecting to SQLite metadata database at: %s", cfg.MetadataDatabaseURL)
	db, err := database.InitDB(cfg.MetadataDatabaseURL)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Internal metadata schema ready.")

	// Instantiate proxy engine for target database connectivity
	proxyEngine := proxy.NewProxyEngine()
	defer func() {
		if err := proxyEngine.CloseAll(); err != nil {
			log.Printf("Error closing proxy connection pools: %v", err)
		}
	}()

	// Instantiate repositories
	projectRepo := repository.NewProjectRepository(db)
	instanceRepo := repository.NewInstanceRepository(db)

	// Instantiate HTTP handlers
	healthHandler := handlers.NewHealthHandler()
	projectsHandler := handlers.NewProjectsHandler(projectRepo, proxyEngine)
	instancesHandler := handlers.NewInstancesHandler(projectRepo, instanceRepo, proxyEngine)
	queryHandler := handlers.NewQueryHandler(instanceRepo, proxyEngine)

	// Router setup
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.NewCORSMiddleware(cfg.CORSOrigins))

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.HealthCheck)

		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projectsHandler.CreateProject)
			r.Get("/", projectsHandler.ListProjects)
			r.Get("/{id}", projectsHandler.GetProject)
			r.Patch("/{id}", projectsHandler.UpdateProject)
			r.Delete("/{id}", projectsHandler.DeleteProject)

			r.Post("/{project_id}/instances", instancesHandler.CreateInstance)
		})

		r.Route("/instances", func(r chi.Router) {
			r.Get("/{id}", instancesHandler.GetInstance)
			r.Patch("/{id}", instancesHandler.UpdateInstance)
			r.Delete("/{id}", instancesHandler.DeleteInstance)

			r.Post("/{id}/test-connection", queryHandler.TestConnection)
			r.Post("/{id}/query", queryHandler.ExecuteQuery)
		})
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  125 * time.Second,
		WriteTimeout: 125 * time.Second,
		IdleTimeout:  180 * time.Second,
	}

	// Server shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Tathya-Avalokan Go Backend listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-shutdownChan
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
