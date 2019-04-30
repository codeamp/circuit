package mongo

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoExtension struct {
	events chan transistor.Event
	data   MongoData

	mongo Mongoer
}

type MongoData struct {
	Hostname   string
	Username   string
	Password   string
	Collection []string
	Prefix     string
}

func init() {
	transistor.RegisterPlugin("mongo", func() transistor.Plugin {
		return &MongoExtension{}
	}, plugins.ProjectExtension{})
}

func (x *MongoExtension) Description() string {
	return "Provision Mongo Assets for Project Use"
}

func (x *MongoExtension) SampleConfig() string {
	return ` `
}

func (x *MongoExtension) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Mongo")
	return nil
}

func (x *MongoExtension) Stop() {
	log.Info("Stopping Mongo")
}

func (x *MongoExtension) Subscribe() []string {
	return []string{
		"project:mongo:create",
		"project:mongo:update",
		"project:mongo:delete",
	}
}

func (x *MongoExtension) Process(e transistor.Event) error {
	var err error
	if e.Matches("project:mongo") {
		log.InfoWithFields(fmt.Sprintf("Process mongo event: %s", e.Event()), log.Fields{})
		switch e.Action {
		case transistor.GetAction("create"):
			err = x.createMongo(e)
		case transistor.GetAction("update"):
			err = x.updateMongo(e)
		case transistor.GetAction("delete"):
			err = x.deleteMongo(e)
		default:
			log.Warn(fmt.Sprintf("Unhandled mongo event: %s", e.Event()))
		}

		if err != nil {
			log.Error(err.Error())
			return err
		}
	}

	return nil
}

func (x *MongoExtension) createMongo(e transistor.Event) error {
	var data MongoData
	data.Hostname = "mongodb+srv://staging-lglzx.mongodb.net"
	data.Username = "checkr"
	data.Password = "dcEpYFRHygrhhLiT"

	client, err := mongo.NewClient(options.Client().ApplyURI(data.Hostname))
	if err != nil {
		log.Error(err.Error())
	} else {
		spew.Dump(client)
	}

	x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("complete"), "Successfully Installed", nil)
	return nil
}

func (x *MongoExtension) updateMongo(e transistor.Event) error {
	x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("complete"), "Successfully Updated", nil)
	return nil
}

func (x *MongoExtension) deleteMongo(e transistor.Event) error {
	x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("complete"), "Successfully Deleted", nil)
	return nil
}

// Wraper for sending an event back thruogh the messaging system for standardization and brevity
func (x *MongoExtension) sendMongoResponse(e transistor.Event, action transistor.Action, state transistor.State, stateMessage string, artifacts []transistor.Artifact) {
	event := e.NewEvent(action, state, stateMessage)
	event.Artifacts = artifacts

	x.events <- event
}

// Pull all the artifacts out from the event that we will need
// in order to service these requests. Stuff them into a local storage object.
func (x *MongoExtension) extractArtifacts(e transistor.Event) (*MongoData, error) {
	var data MongoData

	// // Access Key ID
	// awsAccessKeyID, err := e.GetArtifact("aws_access_key_id")
	// if err != nil {
	// 	x.sendMongoExtensionResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
	// 	return nil, err
	// }
	// data.AWSAccessKeyID = awsAccessKeyID.String()

	return &data, nil
}

// Prepare artifacts array to pass important information back to the front-end
func (x *MongoExtension) buildResultArtifacts(data *MongoData) []transistor.Artifact {
	// Stuff new credentials into artifacts to be used by the front-end
	return []transistor.Artifact{
		// transistor.Artifact{Key: "MongoExtension_aws_access_key_id", Value: accessKey.AccessKeyId, Secret: false},
		// transistor.Artifact{Key: "MongoExtension_aws_secret_key", Value: accessKey.SecretAccessKey, Secret: false},
		// transistor.Artifact{Key: "MongoExtension_aws_region", Value: data.AWSRegion, Secret: false},
		// transistor.Artifact{Key: "MongoExtension_aws_bucket", Value: data.AWSBucket, Secret: false},
		// transistor.Artifact{Key: "MongoExtension_aws_prefix", Value: fmt.Sprintf("%s/", prefix), Secret: false},
	}
}
