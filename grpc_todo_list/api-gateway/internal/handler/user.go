package handler

import (
	"apigateway/internal/service"
	"apigateway/pkg/res"
	"apigateway/pkg/util"
	"context"
	"user/pkg/e"

	"github.com/gin-gonic/gin"
)

// 注册
func UserRegister(c *gin.Context) {
	var userReq service.UserRequest
	PanicIfUserError(c.Bind(&userReq))
	// gin.Key 中取出服务实例
	userService := c.Keys["user"].(service.UserServiceClient)
	userResp, err := userService.UserRegister(context.Background(), &userReq)
	PanicIfUserError(err)
	r := res.Response{
		Data:   userResp,
		Status: uint(userResp.Code),
		Msg:    e.GetMsg(uint(userResp.Code)),
		Error:  err.Error(),
	}
	c.JSON(200, r)
}

// 登录
func UserLogin(c *gin.Context) {
	var userReq service.UserRequest
	PanicIfUserError(c.Bind(&userReq))
	// gin.Key 中取出服务实例(client)
	userService := c.Keys["user"].(service.UserServiceClient)
	userResp, err := userService.UserLogin(context.Background(), &userReq)
	PanicIfUserError(err)
	token, err := util.GenerateToken(uint(userResp.UserDetail.UserID))
	r := res.Response{
		Data: res.TokenData{
			User:  userResp.UserDetail,
			Token: token,
		},
		Status: uint(userResp.Code),
		Msg:    e.GetMsg(uint(userResp.Code)),
		Error:  err.Error(),
	}
	c.JSON(200, r)
}
