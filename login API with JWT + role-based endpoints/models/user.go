package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:255;not null" json:"username"`
	PasswordHash string    `gorm:"size:512;not null" json:"-"`
	Roles        []Role    `gorm:"many2many:user_roles;"`
	Otp          string    `json:"otp"`
	OtpExpiry    time.Time `json:"otp_expiry"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
