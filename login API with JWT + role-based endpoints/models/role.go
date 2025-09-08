package models

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique" json:"role_name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

/*

Name:
	admin
	editor
	user

Permissions:
	Admin:
		Read Users
		Delete Users

		Read role
		Update role
	
	Editor:
		Read Users
		Delete Users

	User:
		Read Users
*/


		// Create role
		// Delete role
