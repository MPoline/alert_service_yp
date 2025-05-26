package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"go.uber.org/zap"
)

const (
	host     = "localhost"
	port     = 9876
	user     = "alertserviceUser"
	password = "alertserviceUser"
	dbname   = "alertservicedb"
)

func DBInit() (db *sql.DB, err error) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		zap.L().Error("Error opening database: ", zap.Error(err))
		return
	}
	zap.L().Info("Successful open to the database")
	return
}

func DBClose(db *sql.DB) {
	if err := db.Close(); err != nil {
		zap.L().Error("Error closing database: ", zap.Error(err))
	} else {
		zap.L().Info("The database connection was closed")
	}
}
