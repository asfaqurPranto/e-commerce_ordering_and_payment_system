package config

import(
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"e_commerce/internal/models"
)
var DB *gorm.DB

func Connect_MySQL(){

	dsn := "root:root@tcp(127.0.0.1:3306)/e_commerce_db?parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err!=nil{
		panic(err.Error())
		
	}
	DB=db
}

func Create_Schemas(){
	//if there is a foreign key relation then parent table have to be created(automigrate) first
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Product{})
	DB.AutoMigrate(&models.Cart{},&models.CartItem{})
	DB.AutoMigrate(&models.Order{},&models.OrderItem{})

}