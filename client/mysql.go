package client

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

type mysqlClient struct {
	db  *sql.DB
	cfg MySQLConfig
}

// NewMySQLClient create new mysql client
func NewMySQLClient(cfg MySQLConfig) (Client, error) {
	db, err := connect(cfg.Host, cfg.User, cfg.Pass, "")
	if err != nil {
		return nil, err
	}
	return &mysqlClient{
		db:  db,
		cfg: cfg,
	}, nil

}

func connect(host, user, pass, database string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (c *mysqlClient) Create() error {
	createDbStmt := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", c.cfg.Database)
	_, err := c.db.Exec(createDbStmt)
	if err != nil {
		return err
	}

	db, err := connect(c.cfg.Host, c.cfg.User, c.cfg.Pass, c.cfg.Database)
	if err != nil {
		return err
	}
	c.db.Close()
	c.db = db

	logrus.Info("create mysql connection succeed")

	return err
}

func (c *mysqlClient) Send([]byte) (latNs int64, statusCode int, body string, err error) {
	return 0, 0, "", nil
}

func (c *mysqlClient) Close() error {
	if c.db != nil {
		c.db.Close()
		logrus.Info("mysql client closed")
	}
	return nil
}

func (c *mysqlClient) Reset() error {
	dropDbStmt := fmt.Sprintf("DROP DATABASE %s;", c.cfg.Database)
	_, err := c.db.Exec(dropDbStmt)
	return err
}
