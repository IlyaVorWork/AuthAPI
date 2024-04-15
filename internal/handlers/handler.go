package handlers

import (
	"AuthAPI/internal/core/domain/models"
	"AuthAPI/pkg/customError"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

type Service interface {
	RegisterUser(login, pass string) error
	LoginUser(login, pass string) (error, string)
	UnregisterUser(login string) error
	AddRoles(login, newRoles string) (error, map[string]string)
	GetUserData(login string) (error, models.User)
}

type UserHandler struct {
	service Service
}

func NewUserHandler(service Service) *UserHandler {
	return &UserHandler{service: service}
}

// Register	     godoc
// @Summary 	 Register new user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 RegisterDTO	body	models.RegisterDTO		true	"Data of new account"
// @Success 	 200 		"Done"			string
// @Failure 	 400 		{object}		responses.Error
// @Router /register [post]
func (handler *UserHandler) Register(c *gin.Context) {
	var queryData models.RegisterDTO
	err := c.ShouldBindJSON(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = handler.service.RegisterUser(queryData.Login, queryData.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "Done")
	return
}

// Login  	 	 godoc
// @Summary 	 Login user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 LoginDTO	body	models.LoginDTO		true	"Account data"
// @Success 	 200 		{object}		responses.LoginSuccess
// @Failure 	 400 		{object}		responses.Error
// @Router /login [post]
func (handler *UserHandler) Login(c *gin.Context) {
	var queryData models.LoginDTO
	err := c.ShouldBindJSON(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err, token := handler.service.LoginUser(queryData.Login, queryData.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token})
	return
}

// Unregister  godoc
// @Summary 	 Unregister user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 UnregisterDTO	body	models.UnregisterDTO		true	"Data of account to delete"
// @Success 	 200 		"Done"			string
// @Failure 	 400 		{object}		responses.Error
// @Router /unregister [delete]
func (handler *UserHandler) Unregister(c *gin.Context) {

	var queryData models.UnregisterDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = handler.service.UnregisterUser(queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "Done")
	return
}

// AddRoles  	 godoc
// @Summary 	 AddRoles user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 AddRolesDTO	body	models.AddRolesDTO		true	"Login of an account and roles to add"
// @Success 	 200 		{object}		responses.AddRolesSuccess
// @Failure 	 400 		{object}		responses.Error
// @Failure 	 400 		{object}		responses.AddRolesError
// @Router /addRoles [put]
func (handler *UserHandler) AddRoles(c *gin.Context) {

	var queryData models.AddRolesDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err, newRolesStatus := handler.service.AddRoles(queryData.Login, queryData.Roles)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "roles status": newRolesStatus})
		return
	}

	c.JSON(http.StatusOK, gin.H{"login": queryData.Login, "roles status": newRolesStatus})
}

// GetUserData  	 godoc
// @Summary 	 GetUserData user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 GetUserDataDTO	body	models.GetUserDataDTO		true	"Login of an account which data to get"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		{object}		responses.GetUserSuccess
// @Failure 	 400 		{object}		responses.Error
// @Router /getUserData [post]
func (handler *UserHandler) GetUserData(c *gin.Context) {

	var queryData models.GetUserDataDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access_token := c.Request.Header["Authorization"][0]

	JWT_SECRET_KEY, _ := os.LookupEnv("JWT_SECRET_KEY")
	claims := jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(access_token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_SECRET_KEY), nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": customError.InvalidTokenError.Error()})
		return
	}

	expired := claims.VerifyExpiresAt(time.Now().Unix(), true)
	if !expired {
		c.JSON(http.StatusBadRequest, gin.H{"error": customError.ExpiredTokenError.Error()})
		return
	}

	err, userData := handler.service.GetUserData(queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": userData})
}
