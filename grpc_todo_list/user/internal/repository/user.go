package repository

import (
	"errors"
	"user/internal/service"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UserId         uint   `gorm:"primarykey"`
	UserName       string `gorm:"unique"`
	NickName       string
	PasswordDigest string
}

const (
	PasswordCost = 12 // 密码加密难度
)

func (user *User) CheckUserExist(req *service.UserRequest) bool {
	if err := DB.Where("user_name=?", req.UserName).First(user).Error; err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}

// 获取用户信息
func (user *User) ShowUserInfo(req *service.UserRequest) (err error) {
	if exist := user.CheckUserExist(req); exist {
		return nil
	}
	return errors.New("UserName Not Exist")
}

// 用户注册
func (user *User) UserCreate(req *service.UserRequest) error {
	if exist := user.CheckUserExist(req); exist {
		return errors.New("UserName Exist")
	}
	newuser := User{
		UserName: req.UserName,
		NickName: req.NickName,
	}
	// 密码加密
	if err := newuser.SetPassword(req.Password); err != nil {
		return err
	}
	return DB.Create(&newuser).Error
}

// 加密
func (user *User) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(bytes)
	return nil
}

// 检查密码
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(password))
	return err == nil
}

func BuildUser(item User) *service.UserModel {
	userModel := service.UserModel{
		UserID:   uint32(item.UserId),
		UserName: item.UserName,
		NickName: item.NickName,
	}
	return &userModel
}
