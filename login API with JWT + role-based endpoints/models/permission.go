package models


type Permission struct {
    ID       uint   `gorm:"primaryKey"`
    Action   string
    Resource string
}
/*
action:
	Create
	Read
	Update
	Delete

Resource: 
	users
	role
*/
