package database

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	uuid "github.com/satori/go.uuid"
)

// genDBName creates a database name for the specified
// project extension with the format <project.slug>-<environment>
func genDBName(pe plugins.ProjectExtension) string {
	projectSlugWithUnderscores := strings.Replace(pe.Project.Slug, "-", "_", -1)
	envWithUnderscores := strings.Replace(pe.Environment, "-", "_", -1)
	if len(projectSlugWithUnderscores) > 20 {
		projectSlugWithUnderscores = projectSlugWithUnderscores[:20]
	}

	if len(envWithUnderscores) > 20 {
		envWithUnderscores = envWithUnderscores[:20]
	}

	uniqueID := uuid.NewV4()

	return fmt.Sprintf("%s_%s_%s", projectSlugWithUnderscores, envWithUnderscores, strings.Replace(uniqueID.String()[:15], "-", "_", -1))
}

// genDBUsername creates a database username for the specified
// project extension with the format <project.slug>-<environment>
func genDBUser(pe plugins.ProjectExtension) string {
	projectSlugWithUnderscores := strings.Replace(pe.Project.Slug, "-", "_", -1)
	envWithUnderscores := strings.Replace(pe.Environment, "-", "_", -1)
	if len(projectSlugWithUnderscores) > 20 {
		projectSlugWithUnderscores = projectSlugWithUnderscores[:20]
	}

	if len(pe.Environment) > 20 {
		envWithUnderscores = envWithUnderscores[:20]
	}

	uniqueID := uuid.NewV4()

	return fmt.Sprintf("%s_%s_%s_user", projectSlugWithUnderscores, envWithUnderscores, strings.Replace(uniqueID.String()[:11], "-", "_", -1))
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
func initDBInstance(dbType string, host string, username string, sslmode string, password string, port string) (*DatabaseInstance, error) {
	var dbInstance DatabaseInstance
	var err error

	switch dbType {
	case POSTGRESQL:
		dbInstance, err = initPostgresInstance(host, username, password, sslmode, port)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(UnsupportedDBTypeErr, dbType)
	}

	return &dbInstance, nil
}
