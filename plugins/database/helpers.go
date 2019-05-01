package database

import (
	"fmt"
	"log"
	"regexp"
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

func genRandomAlphabetStringWithLength(length int) (*string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		log.Fatal(err)
	}

	randString := reg.ReplaceAllString(base64.RawStdEncoding.EncodeToString(b), "")
	if len(randString) > length {
		randString = randString[:length]
	}

	return &randString, nil
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
