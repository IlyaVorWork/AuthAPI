package main

import (
	docs "AuthAPI/docs"
	"AuthAPI/internal/core"
	"AuthAPI/internal/handlers"
	"AuthAPI/internal/repositories"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
)

// @title Auth API
// @version 1.0

// @host localhost:8080
// @BasePath /user
func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	r := gin.Default()

	userRepo := repositories.NewUsersRepository(repositories.AccessDataBase())
	userService := core.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	docs.SwaggerInfo.BasePath = "/user"
	user := r.Group("/user")
	{
		user.POST("/register", userHandler.Register)
		user.POST("/login", userHandler.Login)
		user.DELETE("/unregister", userHandler.Unregister)
		user.PUT("/addRoles", userHandler.AddRoles)
		user.POST("/getUserData", userHandler.GetUserData)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	err := r.Run()
	if err != nil {
		panic(err)
	}
}
