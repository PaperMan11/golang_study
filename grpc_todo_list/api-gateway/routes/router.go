package routes

import (
	"apigateway/internal/handler"
	"apigateway/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(service ...interface{}) *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.Use(middleware.Cors(), middleware.Init(service))
	v1 := ginRouter.Group("/api/v1")
	{
		v1.GET("ping", func(ctx *gin.Context) { ctx.String(200, "pong") })
		// 用户服务
		v1.POST("/user/register", handler.UserRegister)
		v1.POST("/user/login", handler.UserLogin)
	}
	return ginRouter
}
