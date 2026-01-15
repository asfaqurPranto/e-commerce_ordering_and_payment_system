package models

import "time"

type User struct{
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"not null"`
	//Email string `gorm:"uniqueIndex;not null"` //enable indexing by email
	// this feature is not available in mysql i need postgres
	Email string `gorm:"unique"`
	Password string `gorm:"not null"`

	Admin bool `default:"false"`
	CreatedAt time.Time
	UpdatedAt time.Time

}

type LoginRequest struct{
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct{
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`

}