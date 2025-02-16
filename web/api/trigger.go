package api

import (
	"EventTrigger/data"
	"EventTrigger/event"
	. "EventTrigger/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

const (
	API    = "api"
	Timer  = "timer"
	Ticker = "ticker"
)

func SetupTriggerRouters(router *gin.RouterGroup) {
	router.POST("/create", CreateTrigger)
	router.GET("/create", func(context *gin.Context) {
		context.HTML(http.StatusOK, "trigger.tmpl", gin.H{})
	})
	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "triggers.tmpl", gin.H{})
	})
	router.GET("/triggers", GetAllTriggers)
	router.GET("/trigger/:api", TriggerEvent)
	router.GET("/triggers/:api", GetTriggerByAPI)
	router.GET("/get/:id", GetTriggerByID)
	router.PUT("/triggers/:id", UpdateTrigger)
	router.DELETE("/triggers/:id", DeleteTrigger)
}

func GetAllTriggers(c *gin.Context) {

	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))

	userClaims, _ := claims.(*UserClaims)

	var triggers []data.Trigger
	if err := data.DB.Where("user_id = ?", userClaims.ID).Limit(size).Offset((page - 1) * size).Find(&triggers).Find(&triggers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching triggers", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"triggers": triggers,
	})
}

func GetTriggerByAPI(c *gin.Context) {
	api := c.Param("api")
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}
	userClaims, _ := claims.(*UserClaims)
	var trigger data.Trigger
	if err := data.DB.Where("user_id = ? ,  api = ?", userClaims.ID, api).First(&trigger).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Trigger not found", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trigger)
}

func TriggerEvent(c *gin.Context) {
	api := c.Param("api")
	triggerType := c.Param("type")

	if triggerType != "" {
		triggerType = "api"
	}

	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}

	userClaims, _ := claims.(*UserClaims)

	var trigger data.Trigger
	if err := data.DB.Where("user_id = ? ,  api = ?", userClaims.ID, api).First(&trigger).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Trigger not found", "error": err.Error()})
		return
	}

	request := event.Request{
		TriggerId:     trigger.ID,
		TriggerType:   API,
		Payload:       trigger.Payload,
		ExecutionType: "triggered",
		UserId:        userClaims.ID,
		API:           api,
	}

	request.HandleRequest()

	return
}

func GetTriggerByID(c *gin.Context) {
	id := c.Param("id")
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}
	userClaims, _ := claims.(*UserClaims)
	var trigger data.Trigger
	if err := data.DB.Where("user_id = ? and  id = ?", userClaims.ID, id).First(&trigger).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Trigger not found", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trigger)
}

func CreateTrigger(c *gin.Context) {
	var trigger data.Trigger
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}
	userClaims, _ := claims.(*UserClaims)

	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data", "error": err.Error()})
		return
	}

	trigger.UserID = userClaims.ID

	if err := data.DB.Create(&trigger).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating trigger", "error": err.Error()})
		return
	}

	switch trigger.Type {
	case Timer:
		event.TriggerScheduler.AddEvent(event.Trigger{
			TriggerId: trigger.ID,
			UserId:    userClaims.ID,
			Duration:  trigger.Duration.Duration,
			Ticker:    false,
		})

	case Ticker:
		event.TriggerScheduler.AddEvent(event.Trigger{
			TriggerId: trigger.ID,
			UserId:    userClaims.ID,
			Duration:  trigger.Duration.Duration,
			Ticker:    true,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"trigger": trigger,
		"message": "Trigger has been created successfully",
	})
	return
}

func UpdateTrigger(c *gin.Context) {
	id := c.Param("id")
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
		return
	}
	userClaims, _ := claims.(*UserClaims)
	var trigger data.Trigger
	if err := data.DB.Where("user_id = ? and id = ?", userClaims.ID, id).First(&trigger, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Trigger not found", "error": err.Error()})
		return
	}

	if err := data.DB.Delete(&trigger).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "error": err.Error()})
		return
	}

	event.TriggerScheduler.DeleteEvent(trigger.ID)

	trigger.ID = userClaims.ID

	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Trigger", "error": err.Error()})
		return
	}

	trigger.ID = 0

	if err := data.DB.Create(&trigger).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating trigger", "error": err.Error()})
		return
	}

	switch trigger.Type {
	case Timer:
		event.TriggerScheduler.AddEvent(event.Trigger{
			TriggerId: trigger.ID,
			UserId:    userClaims.ID,
			Duration:  trigger.Duration.Duration,
			Ticker:    false,
		})

	case Ticker:
		event.TriggerScheduler.AddEvent(event.Trigger{
			TriggerId: trigger.ID,
			UserId:    userClaims.ID,
			Duration:  trigger.Duration.Duration,
			Ticker:    true,
		})
	}

	if err := data.DB.Save(&trigger).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating trigger", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trigger)
}

func DeleteTrigger(c *gin.Context) {
	id := c.Param("id")
	var trigger data.Trigger
	if err := data.DB.Where("user_id = ? and id = ? ").First(&trigger, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Trigger not found", "error": err.Error()})
		return
	}

	if err := data.DB.Delete(&trigger).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting trigger", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trigger deleted successfully"})
}
