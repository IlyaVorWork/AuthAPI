package models

import (
	"github.com/dgrijalva/jwt-go"
)

type RegisterDTO struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginDTO struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UnregisterDTO struct {
	Login string `json:"login" binding:"required"`
}

type AddRolesDTO struct {
	Login string `json:"login" binding:"required"`
	Roles string `json:"roles" binding:"required"`
}

type GetUserDataDTO struct {
	Login string `json:"login" binding:"required"`
}

type User struct {
	Id       string
	Login    string
	Password string
	Roles    []string
}

type Role struct {
	Id   string
	Name string
}

type TokenClaims struct {
	Login string   `json:"login"`
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

type DownloadFileDTO struct {
	Login    string `json:"login" binding:"required"`
	FileName string `json:"file-name" binding:"required"`
	Path     string `json:"path" binding:"required"`
}

type DeleteFileDTO struct {
	Login    string `json:"login" binding:"required"`
	FileName string `json:"file-name" binding:"required"`
}

type GetFileListDTO struct {
	Login string `json:"login" binding:"required"`
}
