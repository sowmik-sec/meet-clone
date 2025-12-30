package auth

import (
	"meet-clone/internal/modules/auth/adapters/http"
	"meet-clone/internal/modules/auth/adapters/mongodb"
	"meet-clone/internal/modules/auth/application"
	"meet-clone/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Module struct {
	Handler *http.AuthHandler
}

func NewModule(db *mongo.Database, jwtService *jwt.Service) *Module {
	repo := mongodb.NewAuthRepository(db)
	// service := domain.NewAuthService(repo) // Wait, I didn't create Service factory. I created Usecase directly.
	// The implementation plan had NewAuthService, but I implemented NewAuthUseCase.
	// Let's fix this wiring.
	useCase := application.NewAuthUseCase(repo, jwtService)
	handler := http.NewAuthHandler(useCase)

	return &Module{
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	http.RegisterAuthRoutes(router, m.Handler)
}
