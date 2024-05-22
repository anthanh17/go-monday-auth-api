package api

import (
	"context"
	"fmt"
	db "monday-auth-api/db/sqlc"
	token "monday-auth-api/token"
	"monday-auth-api/util"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	config     util.Config
	store      *db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	redisdb    *redis.Client
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store *db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		redisdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
	}

	err = server.redisdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("cannot connect redis: %w", err)
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	// Use CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Chỉnh sửa cho phù hợp với yêu cầu của bạn
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Authorization", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "accept", "origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/login", server.loginUser)
	router.POST("/otp", server.verifyOtp)

	router.POST("/users", server.createUser)
	router.GET("/user/:id", server.getUser)
	router.GET("/users", server.listUser)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
