package db

import (
	"database/sql"
	"retromanager/config"
	"retromanager/constants"
	"retromanager/errs"
)

var (
	dbGameInfo  *sql.DB
	dbMediaInfo *sql.DB
)

func InitGameDB(c *config.DBConfig) error {
	client, err := sql.Open("mysql", buildSqlDataSource(c))
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

func buildSqlDataSource(c *config.DBConfig) string {
	//TODO: impl it
	panic(1)
}

func InitMediaDB(c *config.DBConfig) error {
	client, err := sql.Open("mysql", buildSqlDataSource(c))
	if err != nil {
		return errs.Wrap(constants.ErrDatabase, "open db fail", err)
	}
	if err := client.Ping(); err != nil {
		return errs.Wrap(constants.ErrDatabase, "ping fail", err)
	}
	dbMediaInfo = client
	return nil
}

func GetMediaDB() *sql.DB {
	return dbMediaInfo
}
