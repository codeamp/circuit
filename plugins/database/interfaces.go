package database

/*
	The main variable to account for in the development of this extension
	is the database instance type. If we all of a sudden want MySQL instead of Postgres,
	the only behavior that should change is how the specific database is created and deleted
	when requested. Thus, the DatabaseInstance interface is necessary, so that we can program to
	it rather than specific database types within database.go. A higher-level constructor will be necessary
	to interpret which database-type object (Postgres, MySQL, etc) to return back.
*/

// DatabaseInstace interface
type DatabaseInstance interface {
	CreateDatabase(username string, password string, dbName string) (*DatabaseMetadata, error)
	DeleteDatabase(string) error
	GetInstanceMetadata() InstanceMetadata
}

// InstanceMetadata contains metadata about the database
// and should be inherited by any struct implementing DatabaseInstance
type InstanceMetadata struct {
	ConnectionInformation
}

type BaseDatabaseInstance struct {
	DatabaseInstance
	instanceMetadata InstanceMetadata
}

func (b *BaseDatabaseInstance) GetInstanceMetadata() InstanceMetadata {
	return b.instanceMetadata
}

// ConnectionInformation contains the information
// required for connecting to the instance
type ConnectionInformation struct {
	Credentials
	Endpoint string
	Port     string
}

// DatabaseMetadata contains the information
// required for connecting to a specific database
// within a database instance
type DatabaseMetadata struct {
	Credentials
	Name string
}

// Credentials contains authorization information
type Credentials struct {
	Username string
	Password string
}
