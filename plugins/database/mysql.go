package database

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// MySQL
type MySQL struct {
	BaseDatabaseInstance
	db *gorm.DB
}

// initMySQLInstance opens a MySQL connection to the host
// and returns a DatabaseInstance object, holding the connection object
func initMySQLInstance(host string, username string, password string, sslmode string, port string) (DatabaseInstance, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=True&loc=Local", username, password, host, port))
	if err != nil {
		spew.Dump(err.Error())
		return nil, err
	}

	return &MySQL{
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
func (m *MySQL) CreateDatabaseAndUser(dbName string, username string, password string) (*DatabaseMetadata, error) {
	spew.Dump("CreateDatabaseAndUser")
	if m.db == nil {
		return nil, fmt.Errorf(NilDBInstanceErr, "mysql")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	if err := m.db.LogMode(true).Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName)).Error; err != nil {
		return nil, err
	}

	// create user
	if err := m.db.LogMode(true).Exec(fmt.Sprintf("CREATE USER '%s' IDENTIFIED BY '%s';", username, password)).Error; err != nil {
		return nil, err
	}

	// create user permissions
	if err := m.db.LogMode(true).Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s';", dbName, username)).Error; err != nil {
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
// on the shared MySQL instance and also any roles
// associated with the database
func (m *MySQL) DeleteDatabaseAndUser(dbName string, dbUser string) error {
	if m.db == nil {
		return fmt.Errorf(NilDBInstanceErr, "mysql")
	}

	// we have to use fmt.Sprintf here because of an issue with the underlying driver
	// and properly sanitizing create database statements
	if err := m.db.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName)).Error; err != nil {
		return err
	}

	if err := m.db.Exec(fmt.Sprintf("DROP USER %s;", dbUser)).Error; err != nil {
		return err
	}

	return nil
}
