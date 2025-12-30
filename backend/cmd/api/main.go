package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpRouter "github.com/meet-clone/backend/internal/adapters/input/http"
	"github.com/meet-clone/backend/internal/adapters/input/http/handlers"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/adapters/input/websocket"
	"github.com/meet-clone/backend/internal/adapters/output/mongodb"
	"github.com/meet-clone/backend/internal/config"
	"github.com/meet-clone/backend/internal/core/domain/chat"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/cloudflare"
	"github.com/meet-clone/backend/internal/pkg/jwt"
	"github.com/meet-clone/backend/internal/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	logger.Info.Println("Starting Meet Clone API server...")

	// Load configuration
	cfg := config.Load()
	logger.Info.Printf("Environment: %s", cfg.Environment)

	// Connect to MongoDB
	mongoClient, err := mongodb.NewClient(cfg.MongoURI, "meet-clone")
	if err != nil {
		logger.Error.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	logger.Info.Println("Connected to MongoDB")

	// Create indexes
	if err := mongoClient.CreateIndexes(context.Background()); err != nil {
		logger.Error.Fatalf("Failed to create indexes: %v", err)
	}
	logger.Info.Println("Database indexes created")

	// Initialize repositories
	userRepo := mongodb.NewUserRepository(mongoClient)
	roomRepo := mongodb.NewRoomRepository(mongoClient)
	chatRepo := mongodb.NewChatRepository(mongoClient)

	// Initialize services
	userService := user.NewService(userRepo)
	roomService := room.NewService(roomRepo)
	chatService := chat.NewService(chatRepo)

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize Cloudflare Calls service
	callsService := cloudflare.NewCallsService(cfg.CloudflareAppID, cfg.CloudflareAppSecret)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(chatService)
	go wsHub.Run()
	logger.Info.Println("WebSocket hub started")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, jwtService)
	roomHandler := handlers.NewRoomHandler(roomService)
	chatHandler := handlers.NewChatHandler(chatService)
	callsHandler := handlers.NewCallsHandler(callsService, roomService)
	wsHandler := websocket.NewHandler(wsHub, jwtService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup router
	router := httpRouter.NewRouter(
		authHandler,
		roomHandler,
		chatHandler,
		callsHandler,
		wsHandler,
		authMiddleware,
		cfg,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router.Setup(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info.Printf("Server starting on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info.Println("Server exited")
}
