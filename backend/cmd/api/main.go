package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/cbalite/backend/internal/cache"
	"github.com/cbalite/backend/internal/config"
	"github.com/cbalite/backend/internal/database"
	"github.com/cbalite/backend/internal/middleware"
	"github.com/cbalite/backend/internal/websocket"
	"github.com/cbalite/backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	log, err := logger.New(cfg.Logger.Level, cfg.Logger.Output)
	if err != nil {
		logger.Fatal("Failed to initialize logger: %v", err)
	}
	defer log.Close()

	log.Info("Starting CBA Lite Backend...")

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()
	log.Info("Connected to PostgreSQL database")

	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisCache.Close()
	log.Info("Connected to Redis cache")

	wsHub := websocket.NewHub(log)
	go wsHub.Run()
	log.Info("WebSocket hub started")

	authMiddleware := middleware.NewAuthMiddleware(&cfg.JWT, log)

	app := &Application{
		Config:         cfg,
		Logger:         log,
		DB:             db,
		Cache:          redisCache,
		WSHub:          wsHub,
		AuthMiddleware: authMiddleware,
	}

	router := app.setupRoutes()

	corsMiddleware := middleware.NewCORSMiddleware(&cfg.CORS)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(&cfg.RateLimit, redisCache)
	loggingMiddleware := middleware.NewLoggingMiddleware(log)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(log)

	handler := recoveryMiddleware(
		loggingMiddleware(
			corsMiddleware(
				rateLimitMiddleware(router),
			),
		),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("Server starting on %s", srv.Addr)
		if cfg.TLS.Enabled {
			if err := srv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Fatal("Failed to start HTTPS server")
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Fatal("Failed to start HTTP server")
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server forced to shutdown")
	}

	log.Info("Server exited gracefully")
}

type Application struct {
	Config         *config.Config
	Logger         *logger.Logger
	DB             *database.PostgresDB
	Cache          *cache.RedisCache
	WSHub          *websocket.Hub
	AuthMiddleware *middleware.AuthMiddleware
}

func (app *Application) setupRoutes() *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", app.healthCheckHandler).Methods("GET")

	api.HandleFunc("/auth/register", app.registerHandler).Methods("POST")
	api.HandleFunc("/auth/login", app.loginHandler).Methods("POST")
	api.HandleFunc("/auth/refresh", app.refreshTokenHandler).Methods("POST")
	api.HandleFunc("/auth/logout", app.logoutHandler).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(app.AuthMiddleware.Authenticate)

	protected.HandleFunc("/users/me", app.getCurrentUserHandler).Methods("GET")
	protected.HandleFunc("/users/me", app.updateCurrentUserHandler).Methods("PUT")

	protected.HandleFunc("/teams", app.createTeamHandler).Methods("POST")
	protected.HandleFunc("/teams", app.getTeamsHandler).Methods("GET")
	protected.HandleFunc("/teams/{teamId}", app.getTeamHandler).Methods("GET")
	protected.HandleFunc("/teams/{teamId}", app.updateTeamHandler).Methods("PUT")
	protected.HandleFunc("/teams/{teamId}", app.deleteTeamHandler).Methods("DELETE")

	protected.HandleFunc("/teams/{teamId}/members", app.getTeamMembersHandler).Methods("GET")
	protected.HandleFunc("/teams/{teamId}/members", app.inviteTeamMemberHandler).Methods("POST")
	protected.HandleFunc("/teams/{teamId}/members/{userId}", app.removeTeamMemberHandler).Methods("DELETE")

	protected.HandleFunc("/teams/{teamId}/channels", app.createChannelHandler).Methods("POST")
	protected.HandleFunc("/teams/{teamId}/channels", app.getChannelsHandler).Methods("GET")
	protected.HandleFunc("/channels/{channelId}", app.getChannelHandler).Methods("GET")
	protected.HandleFunc("/channels/{channelId}", app.updateChannelHandler).Methods("PUT")
	protected.HandleFunc("/channels/{channelId}", app.deleteChannelHandler).Methods("DELETE")

	protected.HandleFunc("/channels/{channelId}/messages", app.sendMessageHandler).Methods("POST")
	protected.HandleFunc("/channels/{channelId}/messages", app.getMessagesHandler).Methods("GET")
	protected.HandleFunc("/messages/{messageId}", app.updateMessageHandler).Methods("PUT")
	protected.HandleFunc("/messages/{messageId}", app.deleteMessageHandler).Methods("DELETE")

	protected.HandleFunc("/teams/{teamId}/tasks", app.createTaskHandler).Methods("POST")
	protected.HandleFunc("/teams/{teamId}/tasks", app.getTasksHandler).Methods("GET")
	protected.HandleFunc("/tasks/{taskId}", app.getTaskHandler).Methods("GET")
	protected.HandleFunc("/tasks/{taskId}", app.updateTaskHandler).Methods("PUT")
	protected.HandleFunc("/tasks/{taskId}", app.deleteTaskHandler).Methods("DELETE")

	protected.HandleFunc("/tasks/{taskId}/comments", app.createTaskCommentHandler).Methods("POST")
	protected.HandleFunc("/tasks/{taskId}/comments", app.getTaskCommentsHandler).Methods("GET")

	protected.HandleFunc("/ws", app.websocketHandler)

	return r
}

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"services": map[string]string{
			"database": "unknown",
			"cache":    "unknown",
		},
	}

	if err := app.DB.HealthCheck(); err == nil {
		health["services"].(map[string]string)["database"] = "healthy"
	} else {
		health["services"].(map[string]string)["database"] = "unhealthy"
	}

	if err := app.Cache.HealthCheck(); err == nil {
		health["services"].(map[string]string)["cache"] = "healthy"
	} else {
		health["services"].(map[string]string)["cache"] = "unhealthy"
	}

	respondWithJSON(w, http.StatusOK, health)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}