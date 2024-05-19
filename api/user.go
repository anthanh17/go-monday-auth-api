package api

import (
	"database/sql"
	"errors"
	"fmt"
	db "monday-auth-api/db/sqlc"
	"monday-auth-api/util"
	"net/http"
	mail "monday-auth-api/mail"
	"time"

	"github.com/gin-gonic/gin"
)

type createUserRequest struct {
	UserName string `json:"user_name" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Mail     string `json:"mail" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=user"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		UserName: req.UserName,
		FullName: req.FullName,
		Mail:     req.Mail,
		Role:     req.Role,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type getUserRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fmt.Println("AAAAA-id:", req.ID)

	user, err := server.store.GetUser(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type listUserRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listUser(ctx *gin.Context) {
	var req listUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, users)
}

type loginUserRequest struct {
	UserName string `json:"user_name" binding:"required"`
	Mail     string `json:"mail" binding:"required"`
}

type loginUserResponse struct {
	Message string `json:"message"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Query user in DB
	user, err := server.store.GetUserByUserName(ctx, req.UserName)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Check email
	if req.Mail != user.Mail {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("invalid email")))
		return
	}

	/*
	* Save redis login
	* key: <user_name>_<email>
	* value: Random OTP 4 charactor - 5p time to live
	 */

	key := req.UserName + "_" + req.Mail
	value_otp, err := util.GenerateNumericOTP(6)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("generate otp false")))
		return
	}

	ttl := 300 * time.Second // 5m seconds TTL
	err = server.redisdb.Set(ctx, key, value_otp, ttl).Err()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("save redis false")))
		return
	}

	// Send email
	mail.SendEmail(user.Mail, value_otp)


	ctx.JSON(http.StatusOK, loginUserResponse{
		Message: "OK",
	})
}

type verifyOtpRequest struct {
	UserName string `json:"user_name" binding:"required"`
	Mail     string `json:"mail" binding:"required"`
	Otp      string `json:"otp" binding:"required"`
}

type verifyOtpResponse struct {
	AccessToken           string    `json:"access_token"`
	AcessTokenExpiresAt   time.Time `json:"access_token_expores_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expores_at"`
	User                  db.User   `json:"user"`
}

func (server *Server) verifyOtp(ctx *gin.Context) {
	var req verifyOtpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Query user in DB
	user, err := server.store.GetUserByUserName(ctx, req.UserName)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Check email
	if req.Mail != user.Mail {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("invalid email")))
		return
	}

	/*
	* Check OTP
	* key: <user_name>_<email>
	* value: Random OTP 4 charactor - 5p time to live
	 */

	key := req.UserName + "_" + req.Mail
	otp, err := server.redisdb.Get(ctx, key).Result()
	if err != nil || otp != req.Otp {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("invalid otp")))
		return
	}

	// Create a new Access Token for this user
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.UserName,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create a new Refresh Token for this user
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.UserName,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := verifyOtpResponse{
		AccessToken:           accessToken,
		AcessTokenExpiresAt:   accessPayload.ExpireAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpireAt,
		User:                  user,
	}

	ctx.JSON(http.StatusOK, rsp)

}
