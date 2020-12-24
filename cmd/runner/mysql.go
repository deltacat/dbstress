package runner

import (
	"github.com/deltacat/dbstress/client"
	"github.com/deltacat/dbstress/config"
	"github.com/deltacat/dbstress/data/mysql"
	"github.com/sirupsen/logrus"
)

// MySQLRunner mysql runner
type MySQLRunner struct {
	caseRunner
	layout mysql.Layout
}

// NewMySQLRunner create a new mysql runner instance
func NewMySQLRunner(cli client.Client, cs config.CaseConfig, layout mysql.Layout) MySQLRunner {
	return MySQLRunner{
		caseRunner: caseRunner{
			cli:         cli,
			cfg:         cs,
			concurrency: cs.Concurrent,
		},
		layout: layout,
	}
}

// Run run the case
func (r *MySQLRunner) Run() error {
	logrus.WithField("name", r.cfg.Name).Info("running case")

	if err := r.cli.Create(r.layout.GetCreateStmt()); err != nil {
		return err
	}

	return r.doInsert(r.doWriteMysql)
}
