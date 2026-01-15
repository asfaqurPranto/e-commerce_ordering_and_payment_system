package handlers

import (
	"e_commerce/internal/config"
	"e_commerce/internal/middleware"
	"e_commerce/internal/models"
	"strconv"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	//"github.com/stripe/stripe-go/v78"
	//"github.com/stripe/stripe-go/v78/paymentintent"
	//"github.com/joho/godotenv"
)


func Stripe_Payment(c *gin.Context){
	type Payment_Req  struct{
		OrderID uint `json:"order_id" binding:"required"`
	}
	var req Payment_Req

	err:=c.BindJSON(&req)
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		})
		return
	}
	//getting the user and the order
	user,err:=middleware.User_Info(c)
	if err!=nil{
		c.JSON(http.StatusUnauthorized,gin.H{
			"message":"you must login first",
		})
		return
	}

	var order models.Order

	//checking order exists or not
	err=config.DB.Preload("OrderItems.Product").First(&order,req.OrderID).Error
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"message":"database error",
		})
		return
	}
	if order.ID==0{
		c.JSON(http.StatusNotFound,gin.H{
			"message":"order not found with given order_id",
		})
		return
	}


	//if the order placed by the same user or not
	if user.ID !=order.UserID{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"sala jomidar onnor order e tui payment diccis!",
		})
		return
	}

	//
	if order.OrderStatus=="canceled"{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"your order was cancled reorder the product again",
		})
		return
	}
	// if order.PaymentMehtod=="cod"{
	// 	c.JSON(http.StatusBadRequest,gin.H{
	// 		"message":"you choose cash on delivery",
	// 	})
	// 	return
	// }
	if order.PaymentStatus=="paid"{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"you already paid for this order",
		})
		return
	}
	
	//stripe part
	amount_in_usd_cent:=float64(order.TotalAmount)*100/122.28
	
	intent,err:=paymentintent.New(&stripe.PaymentIntentParams{
		Amount: stripe.Int64(int64(amount_in_usd_cent)), //will need to convert taka to usd cent
		Currency: stripe.String("usd"),

		PaymentMethodTypes: []*string{
        stripe.String("card"),
    	},
		// optional metadata
		Metadata: map[string]string{
			"userId": strconv.Itoa(int(user.ID)),
			"OrderID":strconv.Itoa(int(order.ID)),
			
		},
	})

	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"order":order,
		"message":"you are trying to pay for this order",
		"clientSecret": intent.ClientSecret,
		"paymentIntent":intent.ID,
	})

	//now check order status and payment method

}


func Stripe_Webhook(c *gin.Context){

	c.JSON(http.StatusOK,gin.H{
		"message":"not finished yet. working on this",
	})
}