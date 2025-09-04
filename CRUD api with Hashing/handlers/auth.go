package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
	
    "example.com/crud-api-hashing/models"
    "example.com/crud-api-hashing/utils"
)

type LoginReq struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func Login(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req LoginReq
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var user models.User
        if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }

        ok, err := utils.ComparePassword(user.PasswordHash, req.Password)
        if err != nil || !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "login success", "user": gin.H{"id": user.ID, "username": user.Username}})
    }
}
