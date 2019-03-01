package database

import "github.com/codeamp/circuit/plugins"

func genDBName(pe plugins.ProjectExtension) (*string, error) {
	dbName := "db"
	return &dbName, nil
}

func genDBUser(pe plugins.ProjectExtension) (*string, error) {
	user := "user"
	return &user, nil
}
