package database

import (
	"fmt"
	"strings"

	"crypto/rand"
	"encoding/base64"

	"github.com/codeamp/circuit/plugins"
	uuid "github.com/satori/go.uuid"
)

// genDBName creates a database name for the specified
// project extension with the format <project.slug>-<environment>
func genDBName(pe plugins.ProjectExtension) string {
	projectSlugWithUnderscores := strings.Replace(pe.Project.Slug, "-", "_", -1)
	envWithUnderscores := strings.Replace(pe.Environment, "-", "_", -1)
	if len(projectSlugWithUnderscores) > 10 {
		projectSlugWithUnderscores = projectSlugWithUnderscores[:10]
	}

	if len(envWithUnderscores) > 10 {
		envWithUnderscores = envWithUnderscores[:10]
	}

	uniqueID := uuid.NewV4()

	return fmt.Sprintf("%s_%s_%s", projectSlugWithUnderscores, envWithUnderscores, strings.Replace(uniqueID.String()[:12], "-", "_", -1))
}

// genDBUser creates a database username for the specified
// project extension with the format <project.slug>-<environment>
func genDBUser() string {
	uniqueID := uuid.NewV4()
	return strings.Replace(uniqueID.String()[:14], "-", "_", -1)
}

func genDBPassword() (*string, error) {
	b := make([]byte, DB_PASSWORD_LENGTH)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	randString := base64.URLEncoding.EncodeToString(b)

	return &randString, err
}

// initDBInstance finds the correct db instance type to initialize
// and returns a corresponding DatabaseInstance object
func initDBInstance(dbType string, host string, username string, sslmode string, password string, port string) (*DatabaseInstance, error) {
	var dbInstance DatabaseInstance
	var err error

	switch dbType {
	case POSTGRESQL:
		dbInstance, err = initPostgresInstance(host, username, password, sslmode, port)
		if err != nil {
			return nil, err
		}
	case MYSQL:
		dbInstance, err = initMySQLInstance(host, username, password, sslmode, port)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(UnsupportedDBTypeErr, dbType)
	}

	return &dbInstance, nil
}
