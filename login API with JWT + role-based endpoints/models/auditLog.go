package models

import "time"


type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	Details    string    `json:"details"`
	CreatedAt  time.Time `json:"created_at"`
}