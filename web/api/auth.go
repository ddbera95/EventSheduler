package api

import (
	"EventTrigger/data"
	"EventTrigger/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Credentials struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required"`
}

func SetupAuthRouters(router *gin.RouterGroup) {
	router.POST("/signup", SignUp)
	router.POST("/login", Login)

	router.GET("/signup", func(context *gin.Context) {
		context.HTML(http.StatusOK, "signup.tmpl", gin.H{})
	})
	router.GET("/login", func(context *gin.Context) {
		context.HTML(http.StatusOK, "login.tmpl", gin.H{})
	})
}

func SignUp(c *gin.Context) {

	var newUser data.User

	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser.Password, _ = util.HashPassword(newUser.Password)

	if err := data.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newUser)

}

func Login(c *gin.Context) {
	var credential Credentials

	if err := c.ShouldBindJSON(&credential); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user data.User

	if err := data.DB.Where("email = ?", credential.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if util.CheckPassword(user.Password, credential.Password) != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Credentials are invalid"})
		return
	}

	token, err := util.GenerateJWT(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set JWT token in a cookie
	c.SetCookie("token", token, int((time.Hour * 24).Seconds()), "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})
}
