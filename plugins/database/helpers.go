package database

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
)

func genDBName(pe plugins.ProjectExtension) (*string, error) {
	dbName := "db"
	return &dbName, nil
}

func genDBUser(pe plugins.ProjectExtension) (*string, error) {
	user := "user"
	return &user, nil
}

// initDBInstance finds the correct db instance type to initialize
// and returns a corresponding DatabaseInstance object
func initDBInstance(dbType string, host string, username string, password string, port string) (*DatabaseInstance, error) {
	var dbInstance DatabaseInstance
	switch dbType {
	case POSTGRESQL:
		dbInstance = initPostgresInstance(host, username, password, port)
	default:
		return nil, fmt.Errorf(UnsupportedDBTypeErr, dbType)
	}

	return &dbInstance, nil
}
