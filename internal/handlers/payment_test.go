package handlers

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
)
func TestPayment(t *testing.T){
	
	//confirming the payment using test card (this is usually done inside front end ,for testing webhook we did it here)
	err:=godotenv.Load("../../.env")
	if err!=nil{
		panic("could not load env file")
	}
	stripe.Key=os.Getenv("STRIPE_SECRET_KEY")
	//For testing purpose will copy payment indent id here
	INTENT_ID:="pi_3SzlioLCyp0h0Tke0WL5YAHp"
	confirmed, err := paymentintent.Confirm(INTENT_ID, &stripe.PaymentIntentConfirmParams{
	 	//PaymentMethod: stripe.String("pm_card_visa_chargeDeclinedExpiredCard"),
		PaymentMethod: stripe.String("pm_card_visa"),
	})


	if err != nil {
		fmt.Println("Payment Failed : ",err.Error())
		return
	}
	fmt.Println("Payment Successful : ", confirmed.ID) 
	
}