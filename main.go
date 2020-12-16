package main

import (
	"time"

	"github.com/deltacat/dbstress/cmd"
)

// set by the compiler
var project, timestamp, version, revision string

func main() {
	if timestamp == "" {
		timestamp = time.Now().String()
	}
	if revision == "" {
		revision = "unknown"
	}
	if version == "" {
		version = "dev"
	}
	if project == "" {
		project = "dbstress"
	}
	cmd.Execute(cmd.VersionInfo{
		Project:   project,
		Version:   version,
		Timestamp: timestamp,
		Revision:  revision,
	})
}
