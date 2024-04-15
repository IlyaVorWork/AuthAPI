package repositories

import (
	"database/sql"
	"fmt"
	"os"
)

func AccessDataBase() *sql.DB {

	DB_USERNAME, _ := os.LookupEnv("DB_USERNAME")
	DB_PASSWORD, _ := os.LookupEnv("DB_PASSWORD")
	DB_NAME, _ := os.LookupEnv("DB_NAME")

	//connStr := fmt.Sprintf("postgresql://%s:%s@%s:5432?sslmode=disable", DB_USERNAME, DB_PASSWORD, DB_NAME)
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USERNAME, DB_PASSWORD, DB_NAME)

	db, err := sql.Open("postgres", connStr)

	fmt.Println(db)

	if err != nil {
		panic(err)
	}
	return db
}
