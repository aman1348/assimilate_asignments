package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"example.com/crud-api-hashing/models"
	"example.com/crud-api-hashing/utils"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
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

		// check if the user has the right role it is appliying for
		// user_roles := make(map[string]int)
		// for _, role := range user.Roles {
		//     user_roles[role.Name] = user_roles[role.Name] + 1
		// }
		// for _, role := range req.Role {
		//     _, ok := user_roles[string(role)]
		//     if !ok {
		//         c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		//     return
		//     }
		// }

		ok, err := utils.ComparePassword(user.PasswordHash, req.Password)
		if err != nil || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		var expiryTimeInSeconds uint = 60 * 60
		// create jwt token
		token, err := utils.GenerateAuthJWT(map[string]interface{}{
            "id": user.ID,
			"username": user.Username,
            "roles": user.Roles,
			"isPrivileged": false,
        }, expiryTimeInSeconds)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "login success", "jwt_token": token})
	}
}

// creates a 15min privilaged session
func PrivilegeSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		username := c.GetString("username")

		var user models.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session creation"})
			return
		}

		
		var expiryTimeInSeconds uint = 15 * 60
		// create jwt token
		token, err := utils.GenerateAuthJWT(map[string]interface{}{
            "id": user.ID,
			"username": user.Username,
            "roles": user.Roles,
			"isPrivileged": true,
        }, expiryTimeInSeconds)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "session creation success", "jwt_token": token})
	}
}
