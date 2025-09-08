package config

import "example.com/crud-api-hashing/models"

func InitializePermissions() {
	permissions := []models.Permission{
		{Action: "Create", Resource: "users"},
		{Action: "Read", Resource: "users"},
		{Action: "Update", Resource: "users"},
		{Action: "Delete", Resource: "users"},
		{Action: "Create", Resource: "role"},
		{Action: "Read", Resource: "role"},
		{Action: "Update", Resource: "role"},
		{Action: "Delete", Resource: "role"},
	}

	for _, p := range permissions {
		DB.FirstOrCreate(&p, models.Permission{Action: p.Action, Resource: p.Resource})
	}
}

func InitializeRoles() {
	roles := []models.Role{
		{Name: "admin"},
		{Name: "editor"},
		{Name: "user"},
	}

	for _, r := range roles {
		DB.FirstOrCreate(&r, models.Role{Name: r.Name})
	}
	attachPermissionsToRoles()
}

func attachPermissionsToRoles() {
	var (
		readUsers   models.Permission
		updateUsers models.Permission
		deleteUsers models.Permission
		readRole    models.Permission
		updateRole  models.Permission
	)

	// fetch permissions by action/resource
	DB.First(&readUsers, models.Permission{Action: "Read", Resource: "users"})
	DB.First(&updateUsers, models.Permission{Action: "Update", Resource: "users"})
	DB.First(&deleteUsers, models.Permission{Action: "Delete", Resource: "users"})
	DB.First(&readRole, models.Permission{Action: "Read", Resource: "role"})
	DB.First(&updateRole, models.Permission{Action: "Update", Resource: "role"})

	// Admin: Read/Delete users + Read/Update role
	var admin models.Role
	DB.First(&admin, models.Role{Name: "admin"})
	DB.Model(&admin).Association("Permissions").Replace([]models.Permission{readUsers, deleteUsers, updateUsers, readRole, updateRole})

	// Editor: Read/Delete users
	var editor models.Role
	DB.First(&editor, models.Role{Name: "editor"})
	DB.Model(&editor).Association("Permissions").Replace([]models.Permission{readUsers, deleteUsers})

	// User: Read users
	var user models.Role
	DB.First(&user, models.Role{Name: "user"})
	DB.Model(&user).Association("Permissions").Replace([]models.Permission{readUsers})

}
