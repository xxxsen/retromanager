package db

import (
	"database/sql"
	"retromanager/config"
	"retromanager/constants"
	"retromanager/errs"
)

var (
	dbGameInfo *sql.DB
)

func InitGameDB(c *config.DBConfig) error {
	client, err := sql.Open("", "")
	if err != nil {
		return errs.Wrap(constants.ErrDatabase, "open db fail", err)
	}
	if err := client.Ping(); err != nil {
		return errs.Wrap(constants.ErrDatabase, "ping fail", err)
	}
	dbGameInfo = client
	return nil
}

func GetGameDB() *sql.DB {
	return dbGameInfo
}
