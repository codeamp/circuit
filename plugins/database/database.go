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
	return "The plugin to install databases"
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

	postgresDBInstance := Postgres{
		InstanceMetadata: InstanceMetadata{
			ConnectionInformation: ConnectionInformation{
				Credentials: Credentials{
					Username: instanceUsername.String(),
					Password: instancePassword.String(),
				},
				Port:     instancePort.String(),
				Endpoint: instanceEndpoint.String(),
			},
		},
	}

	dbUsername, err := genDBUser(projectExtensionEvent)
	if err != nil {
		return err
	}

	dbPassword := ""
	dbName, err := genDBName(projectExtensionEvent)
	if err != nil {
		return err
	}

	dbMetadata, err := postgresDBInstance.CreateDatabase(*dbName, *dbUsername, dbPassword)
	if err != nil {
		return err
	}

	// store db metadata into instance
	respEvent := transistor.NewEvent(e.Name, transistor.GetAction("status"), nil)
	respEvent.State = transistor.GetState("complete")

	respEvent.AddArtifact("DB_USER", dbMetadata.Credentials.Username, false)
	respEvent.AddArtifact("DB_PASSWORD", dbMetadata.Credentials.Password, false)
	respEvent.AddArtifact("DB_NAME", dbMetadata.Name, false)
	respEvent.AddArtifact("DB_ENDPOINT", postgresDBInstance.InstanceMetadata.Endpoint, false)
	respEvent.AddArtifact("DB_PORT", postgresDBInstance.InstanceMetadata.Port, false)

	x.events <- respEvent
	return nil
}
