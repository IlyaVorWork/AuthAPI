package handlers

import (
	"auth/internal/core/domain/models"
	"auth/pkg/customError"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"time"
)

const maxUploadSize = 5 << 20

var fileTypes = map[string]interface{}{
	"image/jpeg": nil,
	"image/png":  nil,
}

type Service interface {
	RegisterUser(login, pass string) error
	LoginUser(login, pass string) (string, error)
	UnregisterUser(login string) error
	AddRoles(login, newRoles string) (map[string]string, error)
	GetUserData(login string) (models.User, error)
	CreateBucket(ctx context.Context, login string) error
	RemoveBucket(ctx context.Context, login string) error
	UploadFile(ctx context.Context, login, name string, file io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, login, fileName, path string) error
	DeleteFile(ctx context.Context, login, fileName string) error
	GetFileList(ctx context.Context, login string) ([]string, error)
}

type UserHandler struct {
	service Service
}

func NewUserHandler(service Service) *UserHandler {
	return &UserHandler{service: service}
}

func VerifyToken(c *gin.Context, login string) error {
	access_token := c.Request.Header.Get("Authorization")
	if access_token == "" {
		return customError.TokenNotProvidedError
	}

	JWT_SECRET_KEY, _ := os.LookupEnv("JWT_SECRET_KEY")
	claims := models.TokenClaims{}

	_, err := jwt.ParseWithClaims(access_token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_SECRET_KEY), nil
	})

	if err != nil {
		return customError.InvalidTokenError
	}

	fmt.Println(claims.Roles)

	if claims.Login != login && claims.Roles[0] != "Admin" {
		return customError.NoPermission
	}

	expired := claims.VerifyExpiresAt(time.Now().Unix(), true)
	if !expired {
		return customError.ExpiredTokenError
	}

	return nil
}

func VerifyAdmin(c *gin.Context) error {
	access_token := c.Request.Header.Get("Authorization")
	if access_token == "" {
		return customError.TokenNotProvidedError
	}

	JWT_SECRET_KEY, _ := os.LookupEnv("JWT_SECRET_KEY")
	claims := models.TokenClaims{}

	_, err := jwt.ParseWithClaims(access_token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWT_SECRET_KEY), nil
	})
	if err != nil {
		return customError.InvalidTokenError
	}

	if claims.Roles[0] != "Admin" {
		return customError.NoPermission
	}

	expired := claims.VerifyExpiresAt(time.Now().Unix(), true)
	if !expired {
		return customError.ExpiredTokenError
	}

	return nil
}

// Register	     godoc
// @Summary 	 Register new user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 RegisterDTO	body	models.RegisterDTO		true	"Data of new account"
// @Success 	 200 		"New profile was successfully registered"			string
// @Failure 	 400 		{object}		responses.Error
// @Router /register [post]
func (handler *UserHandler) Register(c *gin.Context) {
	var queryData models.RegisterDTO
	err := c.ShouldBindJSON(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = handler.service.RegisterUser(queryData.Login, queryData.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	_, err = handler.service.AddRoles(queryData.Login, "User")

	err = handler.service.CreateBucket(c, queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "New profile was successfully registered")
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
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	token, err := handler.service.LoginUser(queryData.Login, queryData.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Access_token": token})
	return
}

// Unregister  godoc
// @Summary 	 Unregister user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 UnregisterDTO	body	models.UnregisterDTO		true	"Data of account to delete"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		"Profile was successfully unregistered"			string
// @Failure 	 400 		{object}		responses.Error
// @Router /unregister [delete]
func (handler *UserHandler) Unregister(c *gin.Context) {

	var queryData models.UnregisterDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = VerifyToken(c, queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = handler.service.RemoveBucket(c.Request.Context(), queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = handler.service.UnregisterUser(queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "Profile was successfully unregistered")
	return
}

// AddRoles  	 godoc
// @Summary 	 AddRoles user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 AddRolesDTO	body	models.AddRolesDTO		true	"Login of an account and roles to add"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		{object}		responses.AddRolesSuccess
// @Failure 	 400 		{object}		responses.Error
// @Failure 	 400 		{object}		responses.AddRolesError
// @Router /addRoles [put]
func (handler *UserHandler) AddRoles(c *gin.Context) {

	var queryData models.AddRolesDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = VerifyAdmin(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	newRolesStatus, err := handler.service.AddRoles(queryData.Login, queryData.Roles)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error(), "Roles status": newRolesStatus})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Login": queryData.Login, "Roles status": newRolesStatus})
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
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = VerifyToken(c, queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	userData, err := handler.service.GetUserData(queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": userData})
}

// UploadFile  	 godoc
// @Summary 	 UploadFile user
// @Tags 		 File
// @Accept       mpfd
// @Produce      json
// @Param		 file	formData	file	true	"File to upload"
// @Param		 login	formData	string	true	"Login of a user"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		"File was successfully uploaded" string
// @Failure 	 400 		{object}		responses.Error
// @Router /uploadFile [post]
func (handler *UserHandler) UploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	form, err := reader.ReadForm(c.Request.ContentLength)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	fileHeader := form.File["file"][0]
	login := form.Value["login"][0]

	file, err := fileHeader.Open()

	fmt.Println(file)
	fmt.Println(login)

	err = VerifyToken(c, login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	buffer := make([]byte, fileHeader.Size)
	_, err = file.Read(buffer)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	fileType := http.DetectContentType(buffer)

	if _, value := fileTypes[fileType]; !value {
		c.JSON(http.StatusBadRequest, gin.H{"Error": customError.TypeNotAllowed.Error()})
		return
	}
	file.Seek(0, io.SeekStart)

	err = handler.service.UploadFile(c.Request.Context(), login, fileHeader.Filename, file, fileHeader.Size, fileType)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "File was successfully uploaded")
	return
}

// DownloadFile  	 godoc
// @Summary 	 DownloadFile user
// @Tags 		 File
// @Accept       json
// @Produce      json
// @Param		 DownloadFileDTO	body	models.DownloadFileDTO		true	"Login of an owner, name of file and path of downloading"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		"File was successfully downloaded" string
// @Failure 	 400 		{object}		responses.Error
// @Router /downloadFile [post]
func (handler *UserHandler) DownloadFile(c *gin.Context) {

	var queryData models.DownloadFileDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = VerifyToken(c, queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = handler.service.DownloadFile(c.Request.Context(), queryData.Login, queryData.FileName, queryData.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "File was successfully downloaded")
	return
}

// DeleteFile  	 godoc
// @Summary 	 DeleteFile user
// @Tags 		 File
// @Accept       json
// @Produce      json
// @Param		 DeleteFileDTO	body	models.DeleteFileDTO		true	"Login of an owner and name of file to delete"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		"File was successfully deleted" string
// @Failure 	 400 		{object}		responses.Error
// @Router /deleteFile [delete]
func (handler *UserHandler) DeleteFile(c *gin.Context) {

	var queryData models.DeleteFileDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = VerifyToken(c, queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = handler.service.DeleteFile(c.Request.Context(), queryData.Login, queryData.FileName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "File was successfully deleted")
	return
}

// GetFileList   godoc
// @Summary 	 GetFileList user
// @Tags 		 File
// @Accept       json
// @Produce      json
// @Param		 GetFileListDTO	body	models.GetFileListDTO		true	"Login of an owner of files"
// @Param		 Authorization	header	string		true	"Access token"
// @Success 	 200 		{object}		responses.GetFileListSuccess
// @Failure 	 400 		{object}		responses.Error
// @Router /getFileList [post]
func (handler *UserHandler) GetFileList(c *gin.Context) {

	var queryData models.GetFileListDTO
	err := c.ShouldBind(&queryData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = VerifyToken(c, queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	list, err := handler.service.GetFileList(c.Request.Context(), queryData.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Files list": list})
	return
}
