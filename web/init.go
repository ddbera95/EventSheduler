package web

import (
	utils "EventTrigger/util"
	"EventTrigger/web/api"
	"github.com/gin-gonic/gin"
	"os"
)

func Init(router *gin.Engine) {
	utils.JWTSecretKey = []byte(os.Getenv("SECRET_KEY"))
	api.SetupAuthRouters(router.Group("/auth"))
	api.SetupTriggerRouters(router.Group("/trigger", utils.JWTAuthMiddleware()))
	api.SetupEventRouters(router.Group("/events", utils.JWTAuthMiddleware()))
}
