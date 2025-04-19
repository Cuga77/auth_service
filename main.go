package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/auth-service/config"
	"github.com/auth-service/internal/handlers"
	"github.com/auth-service/internal/middleware"
	"github.com/auth-service/internal/repository"
	"github.com/auth-service/internal/services"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	repo, err := initRepository(cfg, 5)
	if err != nil {
		log.Fatal(err)
	}
	defer closeResource(repo.Close, "DB connection")

	tokenService := services.NewTokenService(cfg.JWTSecret)
	emailNotifier := services.NewEmailNotifier()
	authService := services.NewAuthService(repo, tokenService, emailNotifier)
	authHandler := handlers.NewAuthHandler(authService, emailNotifier)

	router := setupRouter(authHandler, tokenService)
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: withPanicRecovery(router),
	}

	startServer(srv, cfg.ServerPort)
	waitForShutdownSignal()
	shutdownServer(srv, 5*time.Second)
}

func initRepository(cfg *config.Config, maxRetries int) (repository.Repository, error) {
	var repo repository.Repository
	var err error

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to DB (attempt %d/%d)", i+1, maxRetries)
		repo, err = repository.NewPostgres(cfg)
		if err == nil {
			return repo, nil
		}
		log.Printf("Connection failed: %v", err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to DB after %d attempts: %v", maxRetries, err)
}

func setupRouter(authHandler *handlers.AuthHandler, tokenService *services.TokenService) *gin.Engine {
	router := gin.Default()

	authGroup := router.Group("/auth")
	{
		authGroup.GET("/tokens", authHandler.GenerateTokens)
		authGroup.POST("/refresh", authHandler.RefreshTokens)
		authGroup.POST("/logout", authHandler.Logout)
	}

	protected := router.Group("/api")
	protected.Use(middleware.JWTValidator(tokenService))
	{
		protected.GET("/user", authHandler.GetUserData)
	}

	return router
}

func startServer(srv *http.Server, port string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Server panic recovered: %v\n%s", r, string(debug.Stack()))
			}
		}()

		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
}

func waitForShutdownSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}

func shutdownServer(srv *http.Server, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Shutdown panic recovered: %v\n%s", r, string(debug.Stack()))
		}
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server exited properly")
}

func withPanicRecovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v\n%s", r, string(debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func closeResource(closeFunc func() error, resourceName string) {
	if err := closeFunc(); err != nil {
		log.Printf("Error closing %s: %v", resourceName, err)
	}
}
