package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"example.com/crud-api-hashing/models"
	"example.com/crud-api-hashing/utils"
)

type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

func CreateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		// check if role == admin ( there can only be 1 admin)
		if req.Role == "admin" {
			var admin_user models.User
			// err := db.Where("role = ?", "admin").First(&admin_user).Error
			err := db.Preload("Roles", "name = ?", "admin").First(&admin_user).Error
			if err == nil {
				fmt.Println(admin_user)
				c.JSON(http.StatusBadRequest, gin.H{"error": "an admin already exists"})
				return
			}
		}

		// get role based on req, if not found, set role to user
		role, err := utils.GetRoleDetails(db, req.Role)
		if err != nil {
			user_role, err := utils.GetRoleDetails(db, "user")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user role"})
				return
			}
			role = user_role
		}
		// create a new user
		user := models.User{Username: req.Username, PasswordHash: hash, Roles: []models.Role{role}}
		fmt.Println("new user : ", user)
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		utils.LogAudit(db, req.Username, "Create", "users", fmt.Sprintf("created new user - %v with role - %v", req.Username, role.Name))

		c.JSON(http.StatusCreated, gin.H{"id": user.ID, "username": user.Username, "created_at": user.CreatedAt})
	}
}

func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		username := c.GetString("username")
		var req CreateUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := utils.GetuserDetailsWithPermissionsByUsername(db, username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		idUint, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
			return
		}
		if idUint != uint64(user.ID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username & userid mismatch"})
			return
		}
		// check user permissions
		if !utils.CheckUserPermissions(user, "users", "Update") {
			c.JSON(http.StatusForbidden, gin.H{"error": "resource not authorized"})
			return
		}

		updates := make(map[string]interface{})
		// update username if username is passed
		if req.Username != "" {
			updates["username"] = req.Username
		}
		// update password if password is passed
		if req.Password != "" {
			hash, err := utils.HashPassword(req.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
				return
			}
			updates["password_hash"] = hash
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		var updated_user models.User
		if err := db.Model(&updated_user).Where("id = ?", id).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Fetch updated record to return fresh timestamps
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		utils.LogAudit(db, username, "Update", "users", fmt.Sprintf("updated user details for userid - %v", rune(user.ID)))
		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

func DeleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Convert string param to uint
		id := c.Param("id")
		username := c.GetString("username")
		fmt.Println("username : ", username)
		// var action_user models.User
		action_user, err := utils.GetuserDetailsWithPermissionsByUsername(db, username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		// idUint, err := strconv.ParseUint(id, 10, 32)
		// if err != nil {
		// 	c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
		// 	return
		// }
		// if idUint != uint64(user.ID) {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "username & userid mismatch"})
		// 	return
		// }
		// check user permissions
		if !utils.CheckUserPermissions(action_user, "users", "Delete") {
			c.JSON(http.StatusForbidden, gin.H{"error": "resource not authorized"})
			return
		}

		// Parse ID properly
		// var id uint
		// if _, err := fmt.Sscan(idParam, &id); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		// 	return
		// }

		// Find user
		var user models.User
		if err := db.Preload("Roles").First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// Clear associations (important for many2many)
		if err := db.Model(&user).Association("Roles").Clear(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear roles"})
			return
		}

		// Delete user
		if err := db.Delete(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot delete the user"})
			return
		}
		utils.LogAudit(db, username, "Delete", "users", fmt.Sprintf("username - %v, deleted..", username))

		c.JSON(http.StatusOK, gin.H{
			"message":    "user deleted successfully",
			"id":         user.ID,
			"username":   user.Username,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

func GetUserById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("inside getuserbyid id : ", id)
		var user models.User
		// Fetch updated record to return fresh timestamps
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
}

func GetUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

		// Fetch updated record to return fresh timestamps
		if err := db.Find(&users).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var usersWithoutPass []models.User
		for _, user := range users {
			userWithoutPass := models.User{
				ID:        user.ID,
				Username:  user.Username,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
			usersWithoutPass = append(usersWithoutPass, userWithoutPass)
		}

		c.JSON(http.StatusOK, gin.H{
			"users": usersWithoutPass,
		})
	}
}

func GetUserRoleById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("inside getUserRoleById id : ", id)
		username := c.GetString("username")
		fmt.Println("username : ", username)
		// var action_user models.User
		action_user, err := utils.GetuserDetailsWithPermissionsByUsername(db, username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// check user permissions
		if !utils.CheckUserPermissions(action_user, "role", "Read") {
			c.JSON(http.StatusForbidden, gin.H{"error": "resource not authorized"})
			return
		}

		var user models.User
		// Fetch updated record to return fresh timestamps
		// Define a variable to hold the user data

		// Get the user with the given ID, including the associated roles
		err = db.Preload("Roles").First(&user, "id = ?", id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Println("User not found")
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			} else {
				fmt.Println("Error retrieving user:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		utils.LogAudit(db, username, "Read", "role", "get role by id")
		c.JSON(http.StatusOK, gin.H{
			"user_Details": user,
		})
	}
}

type AssignRoleByNameRequest struct {
	Username string   `json:"username" binding:"required"`
	Roles    []string `json:"roles" binding:"required"` // list of role names
}

func UpdateUserRole(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if user has permissions
		username := c.GetString("username")
		fmt.Println("username : ", username)
		// var action_user models.User
		action_user, err := utils.GetuserDetailsWithPermissionsByUsername(db, username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// check user permissions
		if !utils.CheckUserPermissions(action_user, "Update", "role") {
			c.JSON(http.StatusForbidden, gin.H{"error": "resource not authorized"})
			return
		}

		// update role
		var req AssignRoleByNameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// an admin cannot assign anyone as admin(single admin policy)
		for _, new_role := range req.Roles {
			if new_role == "admin" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "cannot assign admin"})
				return
			}
		}

		// Find user by username
		var user models.User
		if err := db.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// Find roles by names
		var roles []models.Role
		if err := db.Where("name IN ?", req.Roles).Find(&roles).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role names"})
			return
		}

		if len(roles) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no valid roles found"})
			return
		}

		// Replace user's roles with given ones
		if err := db.Model(&user).Association("Roles").Replace(&roles); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update roles"})
			return
		}

		utils.LogAudit(db, username, "Update", "role", fmt.Sprintf("update user - %v role", req.Username))
		c.JSON(http.StatusOK, gin.H{
			"message": "roles updated successfully",
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"roles":    roles,
			},
		})

	}
}
