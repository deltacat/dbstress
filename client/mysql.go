package client

import (
	"database/sql"
	"log"

	"github.com/sirupsen/logrus"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

type mysqlClient struct {
	db *sql.DB
}

// NewMySQLClient create new mysql client
func NewMySQLClient(cfg MySQLConfig) (Client, error) {
	db, err := sql.Open("mysql", cfg.Dsn)
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return nil, err
	}

	logrus.WithField("dsn", cfg.Dsn).Info("ping mysql succeed")

	return &mysqlClient{
		db: db,
	}, nil

}

func (c *mysqlClient) Create(command string) error {
	return nil
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
