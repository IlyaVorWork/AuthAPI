package repositories

import (
	"AuthAPI/internal/core/domain/models"
	"AuthAPI/pkg/customError"
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type UsersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (repository *UsersRepository) GetUserByLogin(login string) (models.User, error) {
	var dbData models.User

	err := repository.db.QueryRow("SELECT * FROM profile WHERE profile_login = $1", login).Scan(&dbData.Id, &dbData.Login, &dbData.Password)
	if err != nil {
		return models.User{}, err
	}

	roles, err := repository.GetUserRolesByLogin(login)

	dbData.Roles = roles

	return dbData, err
}

func (repository *UsersRepository) GetRolesList() ([]string, error) {
	var rolesNamesList []string

	rows, err := repository.db.Query("SELECT * FROM role")
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var role models.Role
		err = rows.Scan(&role.Id, &role.Name)
		if err != nil {
			return nil, err
		}
		rolesNamesList = append(rolesNamesList, role.Name)
	}

	return rolesNamesList, nil
}

func (repository *UsersRepository) GetRolesListAsMap() (map[string]bool, error) {
	var roles = make(map[string]bool)

	rows, err := repository.db.Query("SELECT * FROM role")
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var role models.Role
		err = rows.Scan(&role.Id, &role.Name)

		if err != nil {
			return nil, err
		}
		fmt.Println(role.Name)
		roles[role.Name] = true
	}

	return roles, nil
}

func (repository *UsersRepository) GetUserRolesByLogin(login string) ([]string, error) {
	var oldRoles []string

	query := "SELECT role.role_id, role_name FROM profile INNER JOIN profile_role ON profile.profile_id=profile_role.profile_id INNER JOIN role ON profile_role.role_id = role.role_id WHERE profile_login = $1"

	rows, err := repository.db.Query(query, login)
	defer rows.Close()
	if err != nil {
		return oldRoles, err
	}

	for rows.Next() {
		var role models.Role
		err = rows.Scan(&role.Id, &role.Name)
		if err != nil {
			return oldRoles, err
		}
		oldRoles = append(oldRoles, role.Name)
	}

	return oldRoles, nil
}

func (repository *UsersRepository) GetRoleIdByName(name string) (string, error) {
	var dbData models.Role

	err := repository.db.QueryRow("SELECT * FROM role WHERE role_name = $1", name).Scan(&dbData.Id, &dbData.Name)

	return dbData.Id, err
}

func (repository *UsersRepository) Register(login string, pass []byte) error {

	_, err := repository.db.Exec("INSERT INTO profile (profile_login, profile_password) VALUES ($1, $2)", login, pass)
	if err != nil {
		return err
	}

	return nil
}

func (repository *UsersRepository) Login(pass, dbPass string) error {

	err := bcrypt.CompareHashAndPassword([]byte(dbPass), []byte(pass))
	if err != nil {
		return customError.IncorrectPasswordError
	}

	return nil
}

func (repository *UsersRepository) Unregister(login string) error {
	_, err := repository.db.Exec("DELETE FROM profile WHERE profile_login = $1", login)
	if err != nil {
		return err
	}

	return nil
}

func (repository *UsersRepository) AddRole(profileId, newRoleId string) error {

	_, err := repository.db.Exec("INSERT INTO profile_role (profile_id, role_id) VALUES ($1, $2)", profileId, newRoleId)
	if err != nil {
		return err
	}

	return nil
}
