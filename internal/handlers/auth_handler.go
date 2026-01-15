package handlers

import (
	"e_commerce/internal/config"
	"e_commerce/internal/models"
	"os"

	//"encoding/json"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)


func Register(c *gin.Context){
	var reg_req models.RegisterRequest
	err:=c.BindJSON(&reg_req)
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		})
	}
	//hash the password
	hashed_b,err:=bcrypt.GenerateFromPassword([]byte(reg_req.Password),10)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"message":"password hashed failed",
		})
	}
	hashed:=string(hashed_b)
	
	//Save it to database
	user:=models.User{
		Name:reg_req.Name,
		Email: reg_req.Email,
		Password: hashed,
		Admin: false,
	}
	result:=config.DB.Create(&user)
	if result.Error!=nil{
		c.JSON(http.StatusBadRequest,
		gin.H{
			"message":result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated,
	gin.H{
		"message":"User Created Successfully",
	})
	
}

func Login(c *gin.Context){
	var login_request models.LoginRequest
	
	err:=c.BindJSON(&login_request)
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		})
		return
	}
	
	//get the user from database with the requested email
	var user models.User
	config.DB.First(&user,"Email=?",login_request.Email)
	if user.ID==0{
		c.JSON(http.StatusNotFound,gin.H{
			"message":"User not found with this email",
		})
		return
	}
	
	//hases the input pass and compare it with db saved one
	err=bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(login_request.Password))
	//not matched
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":"wrong password",
		})
		return
	}
	//matched

	//create token struct containing signing meteod sub and expire time
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"sub":user.ID,
		"exp":time.Now().Add(time.Hour*24*30).Unix(),

	})
	//sample token= token&{ 0xc000012888 map[alg:HS256 typ:JWT] map[exp:1770898285 sub:4] [] false}

	//will generate tokenString xxx.xxx.xxx from token and secret key
	SECRET_KEY:=os.Getenv("AUTH_SECRET_KEY")
	//generate token string (header.payload.signature) from token
	tokenString,err:=token.SignedString([]byte(SECRET_KEY))
	if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"message":err.Error(),
		
		})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization",tokenString,3600*24*7,"","",false,true)

	c.JSON(http.StatusOK,gin.H{
		"message":"login successful",
	})
}
