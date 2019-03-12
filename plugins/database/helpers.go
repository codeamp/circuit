package database

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/codeamp/circuit/plugins"
)

// genDBName creates a database name for the specified
// project extension with the format <project.slug>-<environment>
func genDBName(pe plugins.ProjectExtension) string {
	return fmt.Sprintf("%s_%s", pe.Project.Slug, pe.Environment)
}

// genDBUsername creates a database username for the specified
// project extension with the format <project.slug>-<environment>
func genDBUser(pe plugins.ProjectExtension) string {
	return fmt.Sprintf("%s_%s_user", pe.Project.Slug, pe.Environment)
}

func genDBPassword() string {
	var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	rand.Seed(time.Now().UnixNano())

	b := make([]rune, DB_PASSWORD_LENGTH)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
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
