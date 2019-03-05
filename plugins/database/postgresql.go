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

// CreateDatabase creates a db within the DB instance
// and creates and grants all read/write privileges
// to the requested userame
func (p *Postgres) CreateDatabaseAndUser(dbName string, username string, password string) (*DatabaseMetadata, error) {
	if p.db == nil {
		return nil, fmt.Errorf(NilDBInstanceErr, "postgres")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	_, err := p.db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	if err != nil {
		return nil, err
	}

	// create user
	_, err = p.db.Exec(fmt.Sprintf("CREATE USER %s with PASSWORD '%s';", username, password))
	if err != nil {
		return nil, err
	}

	// create user permissions
	_, err = p.db.Exec(fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO %s;", dbName, username))
	if err != nil {
		return nil, err
	}

	// _, err = p.db.Exec("FLUSH PRIVILEGES;")
	// if err != nil {
	// 	return nil, err
	// }

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
	_, err := p.db.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))
	if err != nil {
		return err
	}

	_, err = p.db.Exec(fmt.Sprintf("DROP USER %s;", dbUser))
	if err != nil {
		return err
	}

	return nil
}
