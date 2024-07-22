package core

import (
	"auth/internal/core/domain/models"
	"auth/pkg/customError"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
	"io"
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

type FileStorage interface {
	CreateBucket(ctx context.Context, bucketName string) error
	RemoveBucket(ctx context.Context, bucketName string) error
	RemoveObjects(ctx context.Context, bucketName string) error
	UploadFile(ctx context.Context, bucketName, fileName string, file io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, bucketName, fileName, path string) error
	DeleteFile(ctx context.Context, bucketName, fileName string) error
	GetFile(ctx context.Context, bucketName, fileName string) (minio.ObjectInfo, error)
	GetFileList(ctx context.Context, bucketName string) []string
}

type UserService struct {
	repo        UsersRepository
	fileStorage FileStorage
}

func NewUserService(repo UsersRepository, fileStorage FileStorage) *UserService {
	return &UserService{
		repo:        repo,
		fileStorage: fileStorage,
	}
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

func (service *UserService) LoginUser(login, pass string) (string, error) {

	dbData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return "", customError.UnexistingLoginError
	}

	err = service.repo.Login(pass, dbData.Password)
	if err != nil {
		return "", err
	}

	payload := jwt.MapClaims{
		"exp":   time.Now().Add(time.Minute * 60).Unix(),
		"login": dbData.Login,
		"roles": dbData.Roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	JWT_SECRET_KEY, _ := os.LookupEnv("JWT_SECRET_KEY")

	t, err := token.SignedString([]byte(JWT_SECRET_KEY))
	if err != nil {
		return "", err
	}

	return t, err
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

func (service *UserService) AddRoles(login, newRolesString string) (map[string]string, error) {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return nil, customError.UnexistingLoginError
	}

	oldRoles, err := service.repo.GetUserRolesByLogin(login)
	if err != nil {
		return nil, customError.UnexistingLoginError
	}

	existingRoles, err := service.repo.GetRolesListAsMap()
	if err != nil {
		return nil, err
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
			return newRolesStatus, err
		}

		newRolesStatus[newRoles[i]] = "role was successfully added"
	}

	return newRolesStatus, nil
}

func (service *UserService) GetUserData(login string) (models.User, error) {
	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return models.User{}, customError.UnexistingLoginError
	}

	return profileData, nil
}

func (service *UserService) CreateBucket(ctx context.Context, login string) error {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError
	}

	bucketName := fmt.Sprintf("%s-%s", strings.ToLower(profileData.Login), profileData.Id)

	return service.fileStorage.CreateBucket(ctx, bucketName)
}

func (service *UserService) RemoveBucket(ctx context.Context, login string) error {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError
	}

	bucketName := fmt.Sprintf("%s-%s", strings.ToLower(profileData.Login), profileData.Id)

	err = service.fileStorage.RemoveObjects(ctx, bucketName)
	if err != nil {
		return err
	}

	err = service.fileStorage.RemoveBucket(ctx, bucketName)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) UploadFile(ctx context.Context, login, fileName string, file io.Reader, size int64, contentType string) error {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError
	}

	bucketName := fmt.Sprintf("%s-%s", strings.ToLower(profileData.Login), profileData.Id)

	_, err = service.fileStorage.GetFile(ctx, bucketName, fileName)
	if err == nil {
		return customError.ExistingFileError
	}

	return service.fileStorage.UploadFile(ctx, bucketName, fileName, file, size, contentType)
}

func (service *UserService) DownloadFile(ctx context.Context, login, fileName, path string) error {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError
	}

	bucketName := fmt.Sprintf("%s-%s", strings.ToLower(profileData.Login), profileData.Id)

	_, err = service.fileStorage.GetFile(ctx, bucketName, fileName)
	if err != nil {
		return customError.UnexistingFileError
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", path, fileName)); err == nil {
		return customError.ExistingFileError
	}

	err = service.fileStorage.DownloadFile(ctx, bucketName, fileName, path)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) DeleteFile(ctx context.Context, login, fileName string) error {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return customError.UnexistingLoginError
	}

	bucketName := fmt.Sprintf("%s-%s", strings.ToLower(profileData.Login), profileData.Id)

	obj, err := service.fileStorage.GetFile(ctx, bucketName, fileName)
	fmt.Println(obj, err)
	if err != nil {
		return customError.UnexistingFileError
	}

	err = service.fileStorage.DeleteFile(ctx, bucketName, fileName)
	if err != nil {
		return err
	}

	return nil
}

func (service *UserService) GetFileList(ctx context.Context, login string) ([]string, error) {

	profileData, err := service.repo.GetUserByLogin(login)
	if err != nil {
		return make([]string, 0), customError.UnexistingLoginError
	}

	bucketName := fmt.Sprintf("%s-%s", strings.ToLower(profileData.Login), profileData.Id)

	list := service.fileStorage.GetFileList(ctx, bucketName)

	return list, nil
}
