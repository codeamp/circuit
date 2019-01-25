package database

import (
	"errors"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

// Database
type Database struct {
	events chan transistor.Event
}

// init
func init() {
	transistor.RegisterPlugin("database", func() transistor.Plugin {
		return &Database{}
	}, plugins.ProjectExtension{})
}

// Description
func (x *Database) Description() string {
	return ""
}

// SampleConfig
func (x *Database) SampleConfig() string {
	return ` `
}

// Start
func (x *Database) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Database")
	return nil
}

// Stop
func (x *Database) Stop() {
	log.Info("Stopping Database")
}

// Subscribe
func (x *Database) Subscribe() []string {
	return []string{
		"project:database",
	}
}

func getFailedStatusEvent(err error) transistor.Event {
	ev := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), nil)
	ev.State = transistor.GetState("failed")
	ev.StateMessage = err.Error()
	return ev
}

// Process
func (x *Database) Process(e transistor.Event) error {
	log.Debug("Processing database event")

	// confirm all required structured inputs (Project slug, Environment)
	if e.PayloadModel != "plugins.ProjectExtension" {
		err := errors.New("invalid payload model")
		log.Error(err.Error())
		x.events <- getFailedStatusEvent(err)
		return err
	}

	projectExtensionEvent := e.Payload.(plugins.ProjectExtension)

	// Create DB within shared instance of the correct db variant (postgres/mysql)
	instanceEndpoint, err := e.GetArtifact("SHARED_DATABASE_ENDPOINT")
	if err != nil {
		x.events <- getFailedStatusEvent(err)
		return err
	}

	instanceUsername, err := e.GetArtifact("SHARED_DATABASE_ADMIN_USERNAME")
	if err != nil {
		x.events <- getFailedStatusEvent(err)
		return err
	}

	instancePassword, err := e.GetArtifact("SHARED_DATABASE_ADMIN_PASSWORD")
	if err != nil {
		x.events <- getFailedStatusEvent(err)
		return err
	}

	instancePort, err := e.GetArtifact("SHARED_DATABASE_PORT")
	if err != nil {
		x.events <- getFailedStatusEvent(err)
		return err
	}

	psql := Postgres{
		Endpoint: instanceEndpoint.String(),
		Username: instanceUsername.String(),
		Password: instancePassword.String(),
		Port:     instancePort.String(),
	}

	dbName, err := genDBName(&projectExtensionEvent)
	if err != nil {
		x.events <- getFailedStatusEvent(err)
	}

	dbUser, err := genDBUser(&projectExtensionEvent)
	if err != nil {
		x.events <- getFailedStatusEvent(err)
	}

	dbInfo, err := psql.CreateDatabase(*dbName, *dbUser)
	if err != nil {
		x.events <- getFailedStatusEvent(err)
	}

	ev := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), nil)
	ev.State = transistor.GetState("complete")
	ev.AddArtifact("DB_USER", dbInfo.Username, false)
	ev.AddArtifact("DB_PASS", dbInfo.Password, false)
	ev.AddArtifact("DB_ENDPOINT", dbInfo.Endpoint, false)
	ev.AddArtifact("DB_NAME", dbInfo.DBName, false)

	x.events <- ev

	return nil
}

func genDBName(pe *plugins.ProjectExtension) (*string, error) {
	dbName := "db"
	return &dbName, nil
}

func genDBUser(pe *plugins.ProjectExtension) (*string, error) {
	user := "user"
	return &user, nil
}

// DBInfo for the databases within the instance itself
type DBInfo struct {
	Username string
	Password string
	Endpoint string
	DBName   string
}

// Databaser interface
type Databaser interface {
	CreateDatabase(string, string) (DBInfo, error)
	DeleteDatabase(string) error
}

// Postgres
type Postgres struct {
	Endpoint string
	Username string
	Password string
	Port     string
}

// CreateDatabase
func (p *Postgres) CreateDatabase(dbName string, user string) (*DBInfo, error) {
	return &DBInfo{
		Endpoint: "",
		Username: "",
		Password: "",
		DBName:   "",
	}, nil
}

// DeleteDatabase
func (p *Postgres) DeleteDatabase(dbName string) error {
	return nil
}
