package api

import (
	"context"
	"fmt"
	db "monday-auth-api/db/sqlc"
	token "monday-auth-api/token"
	"monday-auth-api/util"

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
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Update with your allowed origins
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	router.Use(cors.New(config))

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
