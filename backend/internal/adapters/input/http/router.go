package http

import (
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpHandlers "github.com/meet-clone/backend/internal/adapters/input/http/handlers"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/adapters/input/websocket"
	"github.com/meet-clone/backend/internal/config"
)

type Router struct {
	router           *mux.Router
	authHandler      *httpHandlers.AuthHandler
	roomHandler      *httpHandlers.RoomHandler
	chatHandler      *httpHandlers.ChatHandler
	callsHandler     *httpHandlers.CallsHandler
	bandwidthHandler *httpHandlers.BandwidthHandler
	wsHandler        *websocket.Handler
	authMiddleware   *middleware.AuthMiddleware
	config           *config.Config
}

func NewRouter(
	authHandler *httpHandlers.AuthHandler,
	roomHandler *httpHandlers.RoomHandler,
	chatHandler *httpHandlers.ChatHandler,
	callsHandler *httpHandlers.CallsHandler,
	bandwidthHandler *httpHandlers.BandwidthHandler,
	wsHandler *websocket.Handler,
	authMiddleware *middleware.AuthMiddleware,
	cfg *config.Config,
) *Router {
	return &Router{
		router:           mux.NewRouter(),
		authHandler:      authHandler,
		roomHandler:      roomHandler,
		chatHandler:      chatHandler,
		callsHandler:     callsHandler,
		bandwidthHandler: bandwidthHandler,
		wsHandler:        wsHandler,
		authMiddleware:   authMiddleware,
		config:           cfg,
	}
}

func (r *Router) Setup() http.Handler {
	// Apply CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{r.config.CORSOrigin}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Create rate limiter (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	// Apply global middleware
	r.router.Use(middleware.SecurityHeaders(r.config))
	r.router.Use(middleware.RequestValidator)
	r.router.Use(middleware.Logger)

	// API version prefix
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Public routes - Auth (with rate limiting)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.Use(rateLimiter.Limit)
	auth.HandleFunc("/register", r.authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", r.authHandler.Login).Methods("POST")
	auth.HandleFunc("/logout", r.authHandler.Logout).Methods("POST")

	// Protected routes - Auth
	// We attach /me to the auth subrouter but wrap it with the auth middleware
	auth.Handle("/me", r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Me))).Methods("GET")

	// Protected routes - Rooms
	rooms := api.PathPrefix("/rooms").Subrouter()
	rooms.Use(r.authMiddleware.Authenticate)
	rooms.HandleFunc("", r.roomHandler.CreateRoom).Methods("POST")
	rooms.HandleFunc("/my-rooms", r.roomHandler.GetUserRooms).Methods("GET")
	rooms.HandleFunc("/{id}", r.roomHandler.GetRoom).Methods("GET")
	rooms.HandleFunc("/{id}/join", r.roomHandler.JoinRoom).Methods("POST")
	rooms.HandleFunc("/{id}/leave", r.roomHandler.LeaveRoom).Methods("POST")
	rooms.HandleFunc("/{id}", r.roomHandler.EndRoom).Methods("DELETE")
	rooms.HandleFunc("/{id}/participants", r.roomHandler.GetParticipants).Methods("GET")

	// Protected routes - Chat
	chat := api.PathPrefix("/rooms/{id}/messages").Subrouter()
	chat.Use(r.authMiddleware.Authenticate)
	chat.HandleFunc("", r.chatHandler.GetMessages).Methods("GET")

	// WebSocket route
	api.HandleFunc("/ws/room/{id}", r.wsHandler.HandleWebSocket)

	// Protected routes - Calls
	calls := api.PathPrefix("/calls").Subrouter()
	calls.Use(r.authMiddleware.Authenticate)
	calls.HandleFunc("/sessions", r.callsHandler.CreateSession).Methods("POST")
	calls.HandleFunc("/sessions/token", r.callsHandler.GenerateToken).Methods("POST")

	// Protected routes - Bandwidth
	bandwidth := api.PathPrefix("/bandwidth").Subrouter()
	bandwidth.Use(r.authMiddleware.Authenticate)
	bandwidth.HandleFunc("/report", r.bandwidthHandler.ReportBandwidth).Methods("POST")
	bandwidth.HandleFunc("/stats", r.bandwidthHandler.GetStats).Methods("GET")
	bandwidth.HandleFunc("/history", r.bandwidthHandler.GetHistory).Methods("GET")

	// Health check
	r.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Wrap with CORS
	// r.router.Use(corsHandler) // Use wrapper instead for better preflight handling

	return corsHandler(r.router)
}
