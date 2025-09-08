package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"example.com/crud-api-hashing/config"
	"example.com/crud-api-hashing/handlers"
	"example.com/crud-api-hashing/middlewares"
	"example.com/crud-api-hashing/models"
)

func main() {
    config.InitDB()

    // Auto migrate models
    if err := config.DB.AutoMigrate(&models.User{}, &models.Role{},&models.Permission{}, &models.AuditLog{}); err != nil {
        log.Fatalf("failed to migrate: %v", err)
    }

    // Populate roles and permissions
    config.InitializePermissions()
    config.InitializeRoles()

    router := gin.Default()
    // User routes

    protected := router.Group("/users")
	protected.Use(middlewares.AuthMiddleware())
	{
        protected.GET("/", handlers.GetUsers(config.DB))
        protected.GET("/:id", handlers.GetUserById(config.DB))
        protected.PUT("/:id", handlers.UpdateUser(config.DB))
        protected.DELETE("/:id", handlers.DeleteUser(config.DB))
		
        protected.GET("/role/:id",middlewares.AdminOnlyMiddleware(config.DB) , handlers.GetUserRoleById(config.DB))
        protected.PUT("/role",middlewares.AdminOnlyMiddleware(config.DB) , handlers.UpdateUserRole(config.DB))

        protected.POST("/login/privileged", handlers.PrivilegeSession(config.DB))
	}
    router.POST("/users", handlers.CreateUser(config.DB))

    // Auth
    router.POST("/login", handlers.Login(config.DB))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("listening on :%s", port)
    router.Run(":" + port)
}
