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
	cfAnalytics "github.com/meet-clone/backend/internal/adapters/output/cloudflare"
	"github.com/meet-clone/backend/internal/adapters/output/mongodb"
	"github.com/meet-clone/backend/internal/config"
	"github.com/meet-clone/backend/internal/core/domain/appointment"
	"github.com/meet-clone/backend/internal/core/domain/availability"
	"github.com/meet-clone/backend/internal/core/domain/billing"
	"github.com/meet-clone/backend/internal/core/domain/chat"
	"github.com/meet-clone/backend/internal/core/domain/eventtype"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/calendar"
	"github.com/meet-clone/backend/internal/pkg/cloudflare"
	"github.com/meet-clone/backend/internal/pkg/email"
	"github.com/meet-clone/backend/internal/pkg/jwt"
	"github.com/meet-clone/backend/internal/pkg/logger"
	"github.com/meet-clone/backend/internal/pkg/scheduler"
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
	billingRepo := mongodb.NewBillingRepository(mongoClient.Database())
	appointmentRepo := mongodb.NewAppointmentRepository(mongoClient.Database())
	availabilityRepo := mongodb.NewAvailabilityRepository(mongoClient.Database())
	eventTypeRepo := mongodb.NewEventTypeRepository(mongoClient.Database())

	// Initialize Email service
	emailService := email.NewResendService(os.Getenv("RESEND_API_KEY"), "onboarding@resend.dev")

	// Initialize Google Calendar service
	googleService := calendar.NewGoogleService(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
	)

	// Initialize Cloudflare Analytics Client
	cfClient := cfAnalytics.NewClient(cfg)

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize Cloudflare Calls service
	callsService := cloudflare.NewCallsService(cfg.CloudflareAccountID, cfg.CloudflareAPIToken, cfg.CloudflareAppID, cfg.CloudflareAppSecret)

	// Initialize services
	userService := user.NewService(userRepo)
	roomService := room.NewService(roomRepo, callsService)
	chatService := chat.NewService(chatRepo)
	billingService := billing.NewService(billingRepo, cfClient)
	appointmentService := appointment.NewService(appointmentRepo, roomService, emailService, eventTypeRepo, userRepo, availabilityRepo, googleService)
	availabilityService := availability.NewService(availabilityRepo)
	eventTypeService := eventtype.NewService(eventTypeRepo)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(chatService)
	go wsHub.Run()
	logger.Info.Println("WebSocket hub started")

	// Initialize and start Reminder Scheduler
	// Frontend URL for links (should be in config, but using hardcoded or cfg if available)
	// Assuming cfg.FrontendURL exists or we construct it.
	// The config loaded in line 38 likely has it or we can default to localhost:3000
	// Checking main.go imports... `config` package is used.
	// Let's assume cfg.FrontendURL or use a default if not present.
	// Actually, I don't see FrontendURL in the config usage here.
	// I'll check config.go if I need to, but for now I'll use a safe default or string.
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	reminderService := scheduler.NewReminderService(appointmentRepo, userRepo, emailService, 5*time.Minute, frontendURL)
	go reminderService.Start(context.Background())
	logger.Info.Println("Reminder scheduler started")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, jwtService, cfg, googleService)
	userHandler := handlers.NewUserHandler(userService)
	roomHandler := handlers.NewRoomHandler(roomService, appointmentService, billingService)
	chatHandler := handlers.NewChatHandler(chatService)
	callsHandler := handlers.NewCallsHandler(callsService, roomService, billingService)
	billingHandler := handlers.NewBillingHandler(billingService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)
	availabilityHandler := handlers.NewAvailabilityHandler(availabilityService)
	eventTypeHandler := handlers.NewEventTypeHandler(eventTypeService)
	wsHandler := websocket.NewHandler(wsHub, jwtService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup router
	router := httpRouter.NewRouter(
		authHandler,
		userHandler,
		roomHandler,
		chatHandler,
		callsHandler,
		billingHandler,
		appointmentHandler,
		availabilityHandler,
		eventTypeHandler,
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
