package database

// Databaser interface
type DatabaseInstance interface {
	CreateDatabase(string, string) (InstanceMetadata, error)
	DeleteDatabase(string) error
}

// InstanceMetadata contains metadata about the database
// and should be inherited by any struct implementing DatabaseInstance
type InstanceMetadata struct {
	ConnectionInformation
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
