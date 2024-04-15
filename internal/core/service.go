package core

import (
	"AuthAPI/internal/core/domain/models"
	"AuthAPI/pkg/customError"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"os"
	"slices"
	"strings"
	"time"
)

type UsersRepository interface {
	GetUserByLogin(login string) (models.User, error)
	GetUserRolesByLogin(login string) ([]string, error)
	GetRolesListAsMap() (map[string]bool, error)
	GetRoleIdByName(role string) (string, error)

	Register(login string, hashPassword []byte) error
	Login(login, password string) error
	Unregister(login string) error
	AddRole(profileId, newRoleId string) error
}

type UserService struct {
	repo UsersRepository
}

func NewUserService(repo UsersRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) RegisterUser(login, pass string) error {

	_, err := service.repo.GetUserByLogin(login)
	if err == nil {
		return customError.ExistingLoginError
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(pass), 14)
	if err != nil {
		return err
	}

	err = service.repo.Register(login, hashPassword)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) LoginUser(login, pass string) (error, string) {

	dbData, err := service.repo.GetUserByLogin(login)
	fmt.Println(dbData, err)
	if err != nil {
		return customError.UnexistingLoginError, ""
	}

	err = service.repo.Login(pass, dbData.Password)
	if err != nil {
		return err, ""
	}

	payload := jwt.MapClaims{
		"login": dbData.Login,
		"exp":   time.Now().Add(time.Minute * 60).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	JWT_SECRET_KEY, _ := os.LookupEnv("JWT_SECRET_KEY")

	t, err := token.SignedString([]byte(JWT_SECRET_KEY))
	if err != nil {
		return err, ""
	}

	return nil, t
}

func (service *UserService) UnregisterUser(login string) error {

	_, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError
	}

	err = service.repo.Unregister(login)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) AddRoles(login, newRolesString string) (error, map[string]string) {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError, nil
	}

	oldRoles, err := service.repo.GetUserRolesByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError, nil
	}

	existingRoles, err := service.repo.GetRolesListAsMap()
	if err != nil {
		return err, nil
	}

	newRoles := strings.Split(newRolesString, " ")
	for i, el := range newRoles {
		newRoles[i] = strings.Title(strings.ToLower(el))
	}

	newRolesStatus := make(map[string]string)

	for i := 0; i < len(newRoles); i++ {
		if !existingRoles[newRoles[i]] {
			newRolesStatus[newRoles[i]] = "this role does not exist"
			continue
		}

		if slices.Contains(oldRoles, newRoles[i]) {
			newRolesStatus[newRoles[i]] = "user already has this role"
			continue
		}

		id, err := service.repo.GetRoleIdByName(newRoles[i])
		if err != nil {
			newRolesStatus[newRoles[i]] = err.Error()
			continue
		}

		err = service.repo.AddRole(profileData.Id, id)
		if err != nil {
			return err, newRolesStatus
		}

		newRolesStatus[newRoles[i]] = "role was successfully added"
	}

	return nil, newRolesStatus
}

func (service *UserService) GetUserData(login string) (error, models.User) {
	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError, models.User{}
	}

	return nil, profileData
}
