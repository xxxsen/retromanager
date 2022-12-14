package db

import (
	"database/sql"

	"github.com/xxxsen/common/database"
	"github.com/xxxsen/common/errs"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbGameInfo *sql.DB
)

func InitGameDB(c *database.DBConfig) error {
	client, err := database.InitDatabase(c)
	if err != nil {
		return errs.Wrap(errs.ErrDatabase, "open db fail", err)
	}
	dbGameInfo = client
	return nil
}

func GetGameDB() *sql.DB {
	return dbGameInfo
}
