package mongo

import (
	"github.com/codeamp/transistor"
)

type MongoExtension struct {
	events chan transistor.Event
	data   MongoData

	MongoAtlasClientBuilder
}

type MongoCloudProvider struct {
	APIEndpoint string
	APIKey      string
	PublicKey   string
	ProjectID   string
	Timeout     int
}

type Credentials struct {
	Username string
	Password string
}

type MongoData struct {
	Hostname                string
	Atlas                   MongoCloudProvider
	CredentialsCheckTimeout int
}
