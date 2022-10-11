package handler

import (
	"context"
	"user/internal/repository"
	"user/internal/service"
	"user/pkg/e"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

// 用户登录
func (s *UserService) UserLogin(ctx context.Context, req *service.UserRequest) (resp *service.UserDetalResponse, err error) {
	var user repository.User
	resp = new(service.UserDetalResponse)
	resp.Code = e.Success
	err = user.ShowUserInfo(req)
	if err != nil {
		resp.Code = e.Error
		return resp, err
	}
	resp.UserDetail = repository.BuildUser(user)
	return resp, err
}

// 用户注册
func (s *UserService) UserRegister(ctx context.Context, req *service.UserRequest) (resp *service.UserDetalResponse, err error) {
	var user repository.User
	resp = new(service.UserDetalResponse)
	resp.Code = e.Success
	err = user.UserCreate(req)
	if err != nil {
		resp.Code = e.Error
		return resp, err
	}
	resp.UserDetail = repository.BuildUser(user)
	return resp, err
}

// func (s *UserService) mustEmbedUnimplementedUserServiceServer() {}
