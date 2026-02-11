package handlers

import (
	"e_commerce/internal/config"
	"e_commerce/internal/middleware"
	"e_commerce/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	//"io"
	"log"
	"net/http"

	//"os"
	"strconv"
	//"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/webhook"
	"gorm.io/gorm"
	//"github.com/stripe/stripe-go/v78/webhook"
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
	//err=config.DB.Preload("OrderItems.Product").First(&order,req.OrderID).Error
	err=config.DB.First(&order,req.OrderID).Error
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
			"message":"wrong payment id",
		})
		return
	}

	if order.PaymentMethod!="stripe" && order.OrderStatus!="waiting for payment"{
		c.JSON(http.StatusBadRequest,gin.H{
			"message:":"please choose correct order",
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
			"UserID": strconv.Itoa(int(user.ID)),
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

		"clientSecret": intent.ClientSecret,
		"paymentIntent":intent.ID,
	})


}


func Stripe_Webhook(c *gin.Context){

	log.Println("webhook triggered")

	payload,err:=io.ReadAll(c.Request.Body)
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"failed to read request body",
		})
		return
	}
	sigHeader:=c.GetHeader("Stripe-Signature")
	endpointSecret:=os.Getenv("STRIPE_WEBHOOK_SECRET")

	event, err := webhook.ConstructEventWithOptions(
		payload,
		sigHeader,
		endpointSecret,
		webhook.ConstructEventOptions{			//have to do this for version mismatch
			IgnoreAPIVersionMismatch: true,
		},
	)	

	if err != nil {
		log.Println(" Signature verification failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid signature",
		})
		return
	}

	// Derive informations from metadata and payment intent
	var pi stripe.PaymentIntent

	json.Unmarshal(event.Data.Raw,&pi)

	orderIDstr:=pi.Metadata["OrderID"]
	orderID,_:=strconv.ParseUint(orderIDstr,10,4)

	userIDstr:=pi.Metadata["UserID"]
	userID,_:=strconv.ParseUint(userIDstr,10,4)


	payment :=models.Payment{
		StripePaymentID: pi.ID,
		OrderID: orderID,
		UserID: userID,
		Amount: int(pi.Amount),
		Currency:string(pi.Currency),
		//set status and did db transaction inside events
	}

	


	if event.Type == "payment_intent.succeeded" {
		
		//Save the payment inside db
		payment.Status="succeed"
		err=config.DB.Create(&payment).Error
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{
				"message":"payment not saved inside db",
			})
			return
		}           
		
		//make oder[id].status=payment done
		 config.DB.Transaction(func(tx *gorm.DB) error {

			//Inside Order table make the order status as payment done
			var order models.Order
			err=tx.Preload("OrderItems.Product").First(&order,orderID).Error
			if err!=nil{
				return err
			}
			var stockFinished bool

    		//Decrease Product Stock by quantity
			tx.Transaction(func(tx2 *gorm.DB)error{

				for _,item :=range order.OrderItems{

					product_id:=item.ProductID
					quantity:=item.Quantity

					var product models.Product
					err:=tx2.Set("gorm:query_option","FOR_UPDATE").First(&product,product_id).Error
					if err!=nil{
						return err;
					}
					newStock:=product.Stock-quantity
					if newStock<0 {
						
						stockFinished=true
						return errors.New("in sufficiant stock")
					}
					product.Stock-=quantity
					err=tx2.Save(&product).Error
					
					if err!=nil{
						return err
					}

				}

				return nil
			})
			if stockFinished{
				order.OrderStatus="(need refund)Payment_Done_Insufficiant_Stock"
				tx.Save(&order)
			}else{
				order.OrderStatus="Payment Done"
				tx.Save(&order)
			}
			
			
     		return nil                  
		})

		
	}else if event.Type == "payment_intent.payment_failed" {
		fmt.Println("failed payment triggred","userID :" ,userID)
		payment.Status="Failed"
		err:=config.DB.Create(&payment)
		if err!=nil{
			fmt.Println("failed payment not created")
		}
		
	}

}