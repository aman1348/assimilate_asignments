package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"

    "example.com/crud-api-hashing/config"
    "example.com/crud-api-hashing/handlers"
    "example.com/crud-api-hashing/models"
)

func main() {
    config.InitDB()

    // Auto migrate models
    if err := config.DB.AutoMigrate(&models.User{}); err != nil {
        log.Fatalf("failed to migrate: %v", err)
    }

    router := gin.Default()

    // User routes
    router.POST("/users", handlers.CreateUser(config.DB))
    router.PUT("/users/:id", handlers.UpdateUser(config.DB))
    router.DELETE("/users/:id", handlers.DeleteUser(config.DB))
    router.GET("/users/:id", handlers.GetUserById(config.DB))
    router.GET("/users", handlers.GetUsers(config.DB))
    // add: r.GET, r.PUT, r.DELETE...

    // Auth
    router.POST("/login", handlers.Login(config.DB))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("listening on :%s", port)
    router.Run(":" + port)
}
