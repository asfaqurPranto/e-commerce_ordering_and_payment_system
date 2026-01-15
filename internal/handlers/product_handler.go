package handlers

import (
	"e_commerce/internal/config"
	"e_commerce/internal/middleware"
	"e_commerce/internal/models"
	"errors"

	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	
	"gorm.io/gorm"
)

//admin/create-product
func Create_Product(c *gin.Context) {

	type CreateProductRequest struct {
		Name        string `binding:"required"`
		Description string `binding:"required"`
		Category    string `binding:"required"`
		Price       int    `binding:"required,gt=0"` //in prod_req price must be greater than 0
		Stock       int    `binding:"required,gte=0"`
	}

	var create_prod CreateProductRequest

	err := c.BindJSON(&create_prod)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	new_prod := models.Product{
		Name:        create_prod.Name,
		Description: create_prod.Description,
		Category:    create_prod.Category,
		Price:       create_prod.Price,
		Stock:       create_prod.Stock,
	}
	result := config.DB.Create(&new_prod)
	if result.Error != nil || new_prod.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "can not create product",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "product created",
		"product": new_prod,
	})

}

// admin/update-product/:id
func Update_Product(c *gin.Context) {
	idStr := c.Param("id") //param value always will be in string

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id format",
		})
		return
	}

	type Update_Req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		//binding:omitempty = if the field not in the request
		//then make it null
		//if we define binding in the fields all it automaticly
		//become null, just to handle gt we did this
		Price *int `json:"price" binding:"omitempty,gt=0"`
		Stock *int `json:"stock" binding:"omitempty,gte=0"`
	}
	var req Update_Req

	err = c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["Category"] = *req.Category
	}
	if req.Price != nil {
		updates["Price"] = *req.Price
	}
	if req.Stock != nil {
		updates["Stock"] = *req.Stock
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no valid fields to update",
		})
		return
	}

	//perform the update
	result := config.DB.Model(&models.Product{}).Where("id=?", id).Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "database error occurred",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "product not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "product updated",
	})

}

// product/:id
func Get_Product_By_Id(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid id format",
		})
		return
	}
	var product models.Product
	result := config.DB.First(&product, id)
	if result.RowsAffected == 0 || product.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "product not found",
		})
		return
	}
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "database error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product": product,
	})

}

// product/add-cart
func Add_Product_To_Cart(c *gin.Context) {
	
	type Cart_Item_Req struct{
		ProductID uint `json:"product_id" binding:"required"`
		Quantity int `json:"quantity" binding:"required,gt=0"`
	}

	var req_item Cart_Item_Req
	err:=c.BindJSON(&req_item)
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		})
		return
	}

	//first we will look into the db to check..
	//if the same item exists in the user cart or not

	user,_:=middleware.User_Info(c)
	userID:=user.ID

	//find the user cart in db
	//if not found then create and return(will get new cart with userID)
	var cart models.Cart
	config.DB.FirstOrCreate(&cart,models.Cart{UserID:userID})

	//now derive same cartItem product from the cart
	//check if same user having same product in the cart
	var cartItem models.CartItem
	result:=config.DB.Where("cart_id=? AND product_id=?",cart.ID,req_item.ProductID).First(&cartItem) 
	
	if result.Error==nil{  //same product exists in user cart
		//just update kore quantity barabo
		config.DB.Model(&cartItem).Update("quantity",cartItem.Quantity+req_item.Quantity)
		// if result.Error!=nil{
		// 	c.JSON(http.StatusInternalServerError,gin.H{
		// 		"message":"database error creating new cart item",
		// 	})
		// 	return
		// }
	}else{
		//same product does not exists on user cart
		//just create new cartItem with previous cart
		
		newItem:=models.CartItem{
			CartID: cart.ID,
			ProductID: req_item.ProductID,
			Quantity: req_item.Quantity,
		}
		config.DB.Create(&newItem)

		// if result.Error!=nil{
		// 	c.JSON(http.StatusInternalServerError,gin.H{
		// 		"message":"database error creating new cart item",
		// 	})
		// return
		// }
	} 
	c.JSON(http.StatusOK,gin.H{
		"message":"item added to your cart",
	})
	
}

// product/direct-order
func Direct_Order(c *gin.Context){


	type Order_Req struct{
		ProductID uint `json:"product_id" binding:"required"`
		Quantity int `json:"quantity" binding:"required,gt=0"`
		PaymentMethod string `json:"payment_method" binding:"required"`
	}

	var req Order_Req
	err:=c.BindJSON(&req)
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		})
		return
	}
	user,err:=middleware.User_Info(c)
	if err!=nil{
		c.JSON(http.StatusUnauthorized,gin.H{
			"message":"you must login with the user id",
		})
		return;
	}

	userID:=user.ID

	err=config.DB.Transaction(func(tx *gorm.DB) error {

		var product models.Product

		//find product and lock the row for update
		err:=tx.First(&product,req.ProductID).Error
		if err!=nil{
			return errors.New("product not found")
		}
		//check product stock
		if product.Stock<req.Quantity{
			return fmt.Errorf("insufficiant stock.only %d left",product.Stock)
		}

		//update product stock
		product.Stock-=req.Quantity
		err=tx.Save(&product).Error
		if err!=nil{
			return err
		}
		
		//Create Order record
		totalPrice:=product.Price*req.Quantity

		//create order first 
		//after creating order we will get order.id from this order
		//and using this order item we will create order items
		order:=models.Order{
			UserID:userID,
			TotalAmount: totalPrice,
			PaymentMehtod: req.PaymentMethod,
			PaymentStatus: "pending",
			OrderStatus:"processing",
			

		}
		err=tx.Create(&order).Error
		if err!=nil{
			return err
		}


		orderItem:=models.OrderItem{
			OrderID: order.ID,
			ProductID: req.ProductID,
			Quantity: req.Quantity,
			Price:product.Price*req.Quantity,
		}
		err=tx.Create(&orderItem).Error
		if err!=nil{
			return err
		}

		return nil

	})
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"message":err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"message":"order created",
		
	})

}

// product/show-cart
func Show_Cart(c *gin.Context){

	user,err:=middleware.User_Info(c)

	if err!=nil{
		c.JSON(http.StatusUnauthorized,gin.H{
			"message":"please login with your account first",
		})
		return
	}
	
	// var cartItems []models.CartItem
	// if err:=config.DB.Preload("Product").Where("user_id=?",user.ID).Find(&cartItems)
	var cart models.Cart
	err=config.DB.Preload("CartItems.Product").Where("user_id=?",user.ID).First(&cart).Error

	if err!=nil{
		c.JSON(http.StatusNotFound,gin.H{
			"message":err.Error(),
		})
		return
	}
	totalPrice:=0
	for _,item :=range cart.CartItems{
		totalPrice+=item.Product.Price*item.Quantity
	}
	c.JSON(http.StatusOK,gin.H{
		"Cart Items":cart.CartItems,
		"total price":totalPrice,
	})

}


func Show_Orders(c *gin.Context){

	user,err:=middleware.User_Info(c)
	if err!=nil{
		c.JSON(http.StatusUnauthorized,gin.H{
			"message":"you must login first",
		})
		return
	}

	var orders []models.Order
	err=config.DB.Preload("OrderItems.Product").Where("user_id=?",user.ID).Find(&orders).Error
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"message":"database error",
		})
		return
	}
	if len(orders)==0{
		c.JSON(http.StatusOK,gin.H{
			"message":"you do not have any order",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"orders":orders,
	})


}