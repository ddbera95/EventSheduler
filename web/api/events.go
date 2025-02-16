package api

import (
	"EventTrigger/data"
	"EventTrigger/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func SetupEventRouters(router *gin.RouterGroup) {

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "events.tmpl", gin.H{})
	})

	router.GET("/events", GetAllEvents)

}

func GetAllEvents(c *gin.Context) {

	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))

	userClaims, _ := claims.(*util.UserClaims)

	var events []data.Event
	if err := data.DB.Where("user_id = ?", userClaims.ID).Limit(size).Offset((page - 1) * size).Find(&events).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching events", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}
