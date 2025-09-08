package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"example.com/crud-api-hashing/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// ------------------
// Argon2 helper (encode/decode)
// ------------------

// Parameters -- tune for your environment
var (
	ArgonTime    uint32 = 1
	ArgonMemory  uint32 = 64 * 1024 // 64 MB
	ArgonThreads uint8  = 4
	ArgonKeyLen  uint32 = 32
	SaltLen             = 16
)

func CheckUserPermissions(user models.User, resource string, action string) (bool) {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Action == action && perm.Resource == resource {
				return true
			}
		}
	}
	return false
}

func GetuserDetailsWithPermissionsByUsername(db *gorm.DB, username string) (models.User, error) {
	var user models.User
	err := db.Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
func GetuserDetailsWithPermissionsById(db *gorm.DB, userID string) (models.User, error) {
	var user models.User
	err := db.Preload("Roles.Permissions").First(&user, userID).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// get Role Details of the given role from db
func GetRoleDetails(db *gorm.DB, role_name string) (models.Role, error) {
	var role models.Role

	err := db.Where("name = ?", role_name).First(&role).Error
	if err != nil {
		return models.Role{}, err
	}
	return role, nil
}

func LogAudit(db *gorm.DB, actorName, action, resource string, details string) error {
	log := models.AuditLog{
		Username: actorName,
		Action:   action,
		Resource: resource,
		Details:  details,
	}
	return db.Create(&log).Error
}

// create a JWT token for auth
func GenerateAuthJWT(userDetails map[string]interface{}, expiryTimeInSeconds uint) (string, error) {
	claims := jwt.MapClaims{
		"username": userDetails["username"],
		"role":     userDetails["role"],
		"exp":      time.Now().Add(time.Second * time.Duration(expiryTimeInSeconds)).Unix(), // expires in 1 hour
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// fmt.Println("jwt secret : ", os.Getenv("JWT_SECRET"))
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// validate the JWT Token
func ValidateJWT(jwt_token string) (*jwt.Token, error) {
	token, err := jwt.Parse(jwt_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	return token, nil
}

// // get claims from the JWT
// func GetJwtClaims(jwt_token string) (bool, error) {
// 	token, err := jwt.Parse(jwt_token, func(token *jwt.Token) (interface{}, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, jwt.ErrSignatureInvalid
// 		}
// 		return []byte(os.Getenv("JWT_SECRET")), nil
// 	})

// 	if err != nil || !token.Valid {
// 		return false, err
// 	}

// 	return true, nil
// }

// generate a salt
func generateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// HashPassword returns an encoded hash that contains the parameters, salt and derived key
func HashPassword(password string) (string, error) {
	salt, err := generateSalt(SaltLen)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", ArgonMemory, ArgonTime, ArgonThreads, b64Salt, b64Hash)
	return encoded, nil
}

// ComparePassword verifies password against encoded hash
func ComparePassword(encodedHash, password string) (bool, error) {
	// Expected format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	paramsPart := parts[3]
	saltB64 := parts[4]
	hashB64 := parts[5]

	var memory uint32
	var timeCost uint32
	var threads uint8
	_, err := fmt.Sscanf(paramsPart, "m=%d,t=%d,p=%d", &memory, &timeCost, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, err
	}

	calcHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(hash)))

	if subtleConstantTimeCompare(calcHash, hash) {
		return true, nil
	}
	return false, nil
}

// Constant-time comparison for passwords
func subtleConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte = 0
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
