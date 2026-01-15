package models

import "time"

//simply child always have to define ParentID in thier attribute
// and also declare parent parent and ondelete and oncascade

//If parant dont need to care about child ,(if parent is independent or
//parent value will not be empty) then parent need to do nothing
//else parent have to define //gorm foreign key inside it
//[] thakle preload korle cole asbe , na thakle double query kora lagto



type Product struct{
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
	Description string `gorm:"type:text"`
	Category string `gorm:"not null"`
	Price int `gorm:"not null"`
	Stock int `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

}

type CartItem struct{
	ID uint `gorm:"primaryKey"`
	CartID uint `gorm:"not null"`

	ProductID uint `gorm:"not null"`
	Product Product `gorm:"foreignKey:ProductID"`

	Quantity int `gorm:"not null;default:1"`


}

type Cart struct{
	ID uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null"`
	User User `gorm:"foreignKey:UserID"` 

	CartItems []CartItem `gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

}

type OrderItem struct{
	ID uint `gorm:"primaryKey" json:"-"`
	OrderID uint  `gorm:"not null" json:"-"`

	ProductID uint `gorm:"not null"`
	Product Product `gorm:"foreignKey:ProductID"`

	Quantity int `gorm:"not null"`
	Price int `gorm:"not null"`
}


type Order struct{
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"not null"`
	User User `gorm:"foreignKey:UserID" json:"-"`

	TotalAmount int `gorm:"not null"`

	PaymentMehtod string `gorm:"not null"`
	PaymentStatus string `gorm:"default:'pending'"`


	OrderStatus string `gorm:"default:'processing'"`
	CreatedAt time.Time

	OrderItems []OrderItem `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

}