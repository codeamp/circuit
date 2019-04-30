package database

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Postgres
type Postgres struct {
	BaseDatabaseInstance
	db *gorm.DB
}

// initPostgresInstance opens a postgresql connection to the host
// and returns a DatabaseInstance object, holding the connection object
func initPostgresInstance(host string, username string, password string, sslmode string, port string) (DatabaseInstance, error) {
	db, err := gorm.Open("postgres", fmt.Sprintf("user=%s host=%s sslmode=%s password=%s port=%s", username, host, sslmode, password, port))
	if err != nil {
		return nil, err
	}

	return &Postgres{
		BaseDatabaseInstance: BaseDatabaseInstance{
			instanceMetadata: InstanceMetadata{
				ConnectionInformation: ConnectionInformation{
					Credentials: Credentials{
						Username: username,
						Password: password,
					},
					Endpoint: host,
					Port:     port,
				},
			},
		},
		db: db,
	}, nil
}

// CreateDatabase creates a db within the DB instance
// and creates and grants all read/write privileges
// to the requested userame
func (p *Postgres) CreateDatabaseAndUser(dbName string, username string, password string) (*DatabaseMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf(NilDBInstanceErr, "postgres")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	if err := p.db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName)).Error; err != nil {
		return nil, err
	}

	// create user
	if err := p.db.Exec(fmt.Sprintf("CREATE USER %s with PASSWORD '%s';", username, password)).Error; err != nil {
		return nil, err
	}

	// create user permissions
	if err := p.db.Exec(fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s;", dbName, username)).Error; err != nil {
		return nil, err
	}

	return &DatabaseMetadata{
		Name: dbName,
		Credentials: Credentials{
			Username: username,
			Password: password,
		},
	}, nil
}

// DeleteDatabaseAndUser deletes the requested database
// on the shared Postgres instance and also any roles
// associated with the database
func (p *Postgres) DeleteDatabaseAndUser(dbName string, dbUser string) error {
	if p.db == nil {
		return fmt.Errorf(NilDBInstanceErr, "postgres")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	if err := p.db.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName)).Error; err != nil {
		return err
	}

	if err := p.db.Exec(fmt.Sprintf("DROP USER %s;", dbUser)).Error; err != nil {
		return err
	}

	return nil
}
