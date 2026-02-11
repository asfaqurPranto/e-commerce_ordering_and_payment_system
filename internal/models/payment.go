
package models
type Payment struct{
	ID uint `gorm:"primaryKey"`
	StripePaymentID string `gorm:"not null"`

	OrderID uint64 
	Order Order `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	//foreign key created and db enforced relation

	UserID uint64 
	User User`gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`


	Amount int `gorm:"not null"`
	Currency string `gorm:"not null"` 
	Status string `gorm:"not null"` 
	
}