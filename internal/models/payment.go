
package models
type payment struct{
	ID uint `gorm:"primaryKey"`
	StripePaymentID string `gorm:"not null"`

	OrderID uint 
	Order `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	//foreign key created and db enforced relation

	UserID uint 
	User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`


	Amount int `gorm:"not null"`
	Currency string `gorm:"not null"` 
	Status string `gorm:"not null"` 
	
}