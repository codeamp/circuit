package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Postgres
type Postgres struct {
	BaseDatabaseInstance
	db *sql.DB
}

// initPostgresInstance opens a postgresql connection to the host
// and returns a DatabaseInstance object, holding the connection object
func initPostgresInstance(host string, username string, password string, port string) DatabaseInstance {
	db, _ := sql.Open("postgres", fmt.Sprintf("user=%s host=%s sslmode=%s password=%s port=%s", username, host, "disable", password, port))
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
	}
}

// CreateDatabase
func (p *Postgres) CreateDatabase(dbName string, username string, password string) (*DatabaseMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf(NilDBInstanceErr, "postgres")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	_, err := p.db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if err != nil {
		return nil, err
	}

	// create user permissions
	_, err = p.db.Exec(fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s;", dbName, username))
	if err != nil {
		return nil, err
	}

	_, err = p.db.Exec("FLUSH PRIVILEGES;")
	if err != nil {
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

// DeleteDatabase
func (p *Postgres) DeleteDatabase(dbName string) error {
	if p.db == nil {
		return fmt.Errorf(NilDBInstanceErr, "postgres")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	_, err := p.db.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
	if err != nil {
		return err
	}

	return nil
}
