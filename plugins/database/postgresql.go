package database

// Postgres
type Postgres struct {
	InstanceMetadata
}

// CreateDatabase
func (p *Postgres) CreateDatabase(dbName string, username string, password string) (*DatabaseMetadata, error) {

	//TODO: logic for actually provisioning a database within the instance

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
	return nil
}
