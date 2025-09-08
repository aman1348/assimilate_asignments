package middlewares

import (
	"fmt"
	"net/http"

	"example.com/crud-api-hashing/models"
	"example.com/crud-api-hashing/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		fmt.Println("auth header : ", authHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		token, err := utils.ValidateJWT(authHeader)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Optionally, set username in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("username", claims["username"])
			fmt.Println(" claims: ", claims["username"])
		}

		fmt.Println(" in middlewaretoken Authorized")

		c.Next()
	}
}

func AdminOnlyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// id := c.Param("id")
		username := c.GetString("username")
		// c.GetHeader("username")
		var user models.User

		if err := db.Preload("Roles").First(&user, "username = ?", username).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		// fmt.Println(username, "  == ", user.Username)
		// if username != user.Username {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Request"})
		// 	c.Abort()
		// 	return
		// }
		isAdmin := false
		for _, role := range user.Roles {
			if role.Name == "admin"{
				isAdmin = isAdmin || true
			}
		}

		if !isAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authorized"})
			c.Abort()
			return
		}

		fmt.Println(" user is admin")

		c.Next()
	}
}

// func AdminOnlyMiddleware(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// id := c.Param("id")
// 		username := c.GetString("username")
// 		// c.GetHeader("username")
// 		var user models.User

// 		if err := db.Preload("Roles").First(&user, "username = ?", username).Error; err != nil {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
// 			c.Abort()
// 			return
// 		}

// 		// fmt.Println(username, "  == ", user.Username)
// 		// if username != user.Username {
// 		// 	c.JSON(http.StatusForbidden, gin.H{"error": "Invalid Request"})
// 		// 	c.Abort()
// 		// 	return
// 		// }
// 		isAdmin := false
// 		for _, role := range user.Roles {
// 			if role.Name == "admin"{
// 				isAdmin = isAdmin || true
// 			}
// 		}

// 		if !isAdmin {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authorized"})
// 			c.Abort()
// 			return
// 		}

// 		fmt.Println(" user is admin")

// 		c.Next()
// 	}
// }
