package processing

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

/*
	Набор констант для строки входа в БД
*/
const (
	HOST        = "database" //"database" - docker
	PORT        = 5432
	DB_USER     = "postgres"
	DB_PASSWORD = "12345"
	DB_NAME     = "base" //"base" - docker
)

/*
	Подключение к Базе данных
*/
func InitDatabase() *sql.DB {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", HOST, PORT, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	db.SetMaxOpenConns(100) //ограничить доступ к БД для клиентов
	return db
}

// Функция проверки на ошибку
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
