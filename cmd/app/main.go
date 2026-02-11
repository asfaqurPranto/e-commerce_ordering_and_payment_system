package main
import(
	"github.com/gin-gonic/gin"
	"e_commerce/internal/handlers"
	"e_commerce/internal/config"
	"e_commerce/internal/middleware"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v78"
	"os"
	
)
func init(){
	//loading .env
	godotenv.Load()
	stripe.Key=os.Getenv("STRIPE_SECRET_KEY")   

	config.Connect_MySQL()
	config.Create_Schemas()
	
}
func main(){
	router:=gin.Default()
	
	public_route:=router.Group("")
	{
		public_route.POST("/login",handlers.Login)
		public_route.POST("/register",handlers.Register)
		//will add logout
	}
	product_route:=router.Group("/product")
	{
		
		product_route.GET("/:id",middleware.Login_Required,handlers.Get_Product_By_Id)

		product_route.POST("/add-cart",middleware.Login_Required,handlers.Add_Product_To_Cart)

		product_route.POST("/direct-order",middleware.Login_Required,handlers.Direct_Order)

		product_route.GET("/show-cart",middleware.Login_Required,handlers.Show_Cart)
		product_route.GET("/show-orders",middleware.Login_Required,handlers.Show_Orders)

	}

	admin_route:=router.Group("/admin")
	{
	// 	admin_route.GET("/show-all",middleware.Login_required,middleware.admin_required,handlers.Show_all_product)
	 	admin_route.POST("/create-product",middleware.Login_Required,middleware.Admin_Required,handlers.Create_Product)
	 	admin_route.POST("/update-product/:id",middleware.Login_Required,middleware.Admin_Required,handlers.Update_Product)
	// 	//will make it unavailable by making stock=0 rather than deleting it 
	// 	admin_route.POST("/delete-product",middleware.Login_required,moddieware.admin_required,handlers.Delete_product)

	}
///stripe/webhook
	payment_route:=router.Group("/stripe")
	{
		payment_route.POST("/create-payment-intent",middleware.Login_Required,handlers.Stripe_Payment)
		payment_route.POST("/webhook",handlers.Stripe_Webhook)
	}
	router.Run()

}