package db

import (
	"database/sql"
	"fmt"
	"retromanager/config"

	"github.com/xxxsen/errs"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbGameInfo  *sql.DB
	dbMediaInfo *sql.DB
)

func InitGameDB(c *config.DBConfig) error {
	client, err := sql.Open("mysql", buildSqlDataSource(c))
	if err != nil {
		return errs.Wrap(errs.ErrDatabase, "open db fail", err)
	}
	if err := client.Ping(); err != nil {
		return errs.Wrap(errs.ErrDatabase, "ping fail", err)
	}
	dbGameInfo = client
	return nil
}

func GetGameDB() *sql.DB {
	return dbGameInfo
}

func buildSqlDataSource(c *config.DBConfig) string {
	return fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8mb4", c.User, c.Pwd, "tcp", c.Host, c.Port, c.DB)
}

func InitFileDB(c *config.DBConfig) error {
	client, err := sql.Open("mysql", buildSqlDataSource(c))
	if err != nil {
		return errs.Wrap(errs.ErrDatabase, "open db fail", err)
	}
	if err := client.Ping(); err != nil {
		return errs.Wrap(errs.ErrDatabase, "ping fail", err)
	}
	dbMediaInfo = client
	return nil
}

func GetMediaDB() *sql.DB {
	return dbMediaInfo
}
