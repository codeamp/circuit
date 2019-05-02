package mongo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/Clever/atlas-api-client/digestauth"
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	atlas "github.com/Clever/atlas-api-client/gen-go/client"
	atlas_models "github.com/Clever/atlas-api-client/gen-go/models"
)

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

// How does this work?
// The artifacts that should be returned after a Mongo
// extension is successfuly created should be
// the credentials for which the project will
// be using to access the bucket, as well as the prefix
// that has been assigned for this application to use
// in addition to the region the bucket is in
//
// Accepts:
//
//	mongo_atlas_endpoint
//	mongo_atlas_api_public
//	mongo_atlas_api_private
//	mongo_atlas_project_id
//  mongo_atlas_api_timeout
//
//  mongo_hostname
//  mongo_credentials_check_timeout
func (x *MongoExtension) Process(e transistor.Event) error {
	var err error
	if e.Matches("project:mongo") {
		log.InfoWithFields(fmt.Sprintf("Process mongo event: %s", e.Event()), log.Fields{})
		switch e.Action {
		case transistor.GetAction("create"):
			err = x.createMongoExtension(e)
		case transistor.GetAction("update"):
			err = x.updateMongoExtension(e)
		case transistor.GetAction("delete"):
			err = x.deleteMongoExtension(e)
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

func (x *MongoExtension) listMongoUsers(atlasAPI atlas.Client, data *MongoData) (*atlas_models.GetDatabaseUsersResponse, error) {
	log.Error("listMongoUsers")

	resp, err := atlasAPI.GetDatabaseUsers(context.Background(), data.Atlas.ProjectID)
	if err == nil {
		for _, databaseUser := range resp.Results {
			log.Warn("User: ", databaseUser.Username)
			log.Warn("DB: ", databaseUser.DatabaseName)
			log.Warn("Roles (ct): ", len(databaseUser.Roles))
		}
	} else {
		log.Error(err.Error())
		return nil, err
	}

	return resp, nil
}

func (x *MongoExtension) getMongoUser(atlasAPI atlas.Client, data *MongoData, userName string) (*atlas_models.DatabaseUser, error) {
	log.Error("getMongoUser")

	getUserInput := &atlas_models.GetDatabaseUserInput{
		GroupID:  data.Atlas.ProjectID,
		Username: userName,
	}
	resp, err := atlasAPI.GetDatabaseUser(context.Background(), getUserInput)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return resp, nil
}

func (x *MongoExtension) createMongoUser(atlasAPI atlas.Client, data *MongoData, databaseName string, userName string) (*Credentials, error) {
	log.Error("createMongoUser")

	generatedPassword, err := x.genRandomAlpha(16)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	credentials := Credentials{
		Username: userName,
		Password: *generatedPassword,
	}

	createUserInput := &atlas_models.CreateDatabaseUserInput{
		GroupID: data.Atlas.ProjectID,
		CreateDatabaseUserRequest: &atlas_models.CreateDatabaseUserRequest{
			DatabaseUser: atlas_models.DatabaseUser{
				DatabaseName: "admin",
				GroupID:      data.Atlas.ProjectID,
				Links:        []*atlas_models.Link{},
				Roles: []*atlas_models.Role{
					&atlas_models.Role{
						CollectionName: "",
						DatabaseName:   databaseName,
						RoleName:       "readWrite",
					},
				},
				Username: userName,
			},
			Password: *generatedPassword,
		},
	}
	_, err = atlasAPI.CreateDatabaseUser(context.Background(), createUserInput)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &credentials, nil
}

func (x *MongoExtension) deleteMongoUser(atlasAPI atlas.Client, data *MongoData, userName string) error {
	log.Error("deleteMongoUser")

	_, err := x.getMongoUser(atlasAPI, data, userName)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "No user with username") {
			log.Error("User already deleted? ", userName)
		}
	} else {
		deleteUserInput := &atlas_models.DeleteDatabaseUserInput{
			GroupID:  data.Atlas.ProjectID,
			Username: userName,
		}
		err := atlasAPI.DeleteDatabaseUser(context.Background(), deleteUserInput)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}

	return nil
}

// mongoConnection := fmt.Sprintf("mongodb+srv://%s:%s@%s", data.Username, data.Password, data.Hostname)

// 	// Ensure we can construct a client interface with no issues
// 	log.Warn("Building a new client")
// 	client, err := mongo.NewClient(options.Client().ApplyURI(mongoConnection))
// 	if err != nil {
// 		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), "Failed to Build Connection to Mongo", nil)
// 		return err
// 	}

// 	// Provide timeout value and attempt to connect to the mongo database
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	err = client.Connect(ctx)

// 	if err != nil {
// 		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), "Failed to Build Connection to Mongo", nil)
// 		return err
// 	} else {
// 		log.Warn("Pinging client!")
// 		err = client.Ping(ctx, readpref.Primary())
// 		if err != nil {
// 			x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), "Failed to Ping Mongo", nil)
// 			return err
// 		}
// 	}
func (x *MongoExtension) genRandomAlpha(length int) (*string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	// Make a Regex to say we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}

	// a is an arbitrary char, no significance other than it being a placeholder
	randString := reg.ReplaceAllString(base64.RawStdEncoding.EncodeToString(b), "a")[:length]

	return &randString, nil
}

func (x *MongoExtension) getAtlasClient(data *MongoData) atlas.Client {
	atlasAPI := atlas.New(data.Atlas.APIEndpoint)
	digestT := digestauth.NewTransport(
		data.Atlas.PublicKey,
		data.Atlas.APIKey,
	)
	atlasAPI.SetTransport(&digestT)

	return atlasAPI
}

func (x *MongoExtension) createMongoExtension(e transistor.Event) error {
	data, err := x.extractArtifacts(e)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	payload := e.Payload.(plugins.ProjectExtension)

	userName := payload.Project.Slug
	databaseName := payload.Project.Slug

	atlasAPI := x.getAtlasClient(data)

	// Check to see if user already exists
	createMongoUser := false
	databaseUser, err := x.getMongoUser(atlasAPI, data, userName)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "No user with username") {
			createMongoUser = true
		}
	} else {
		if databaseUser != nil {
			log.Warn("Found User! Do not Create!")
			x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), "User Exists - Duplicate Extension?", nil)
		} else {
			log.Warn("Did not find database user. Create now!")
			createMongoUser = true
		}
	}

	var credentials *Credentials
	if createMongoUser == true {
		credentials, err = x.createMongoUser(atlasAPI, data, databaseName, userName)
		if err != nil {
			log.Error(err.Error())
		}
	}

	x.sendMongoResponse(e, transistor.GetAction("status"),
		transistor.GetState("complete"),
		"Mongo Provisioning Complete.\nRemoving this extension does not delete any data.",
		x.buildResultArtifacts(data, payload.Project.Slug, credentials))

	x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("complete"), "Successfully Installed", nil)
	return nil
}

// Check credentials provided from the previous success message
// If they still work, everything is great
// During the update process, try to perform a Ping! operation
// to ensure that we can connect using the credentials provided to the user
func (x *MongoExtension) updateMongoExtension(e transistor.Event) error {
	data, err := x.extractArtifacts(e)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	atlasAPI := x.getAtlasClient(data)

	_, err = x.listMongoUsers(atlasAPI, data)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// Report back success
	ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Nothing to Update. Removing this extension does not delete any data.")

	// Grab the artifacts from the previous run
	ev.Artifacts = []transistor.Artifact{
		x.getArtifactIgnoreError(&e, "mongo_host"),
		x.getArtifactIgnoreError(&e, "mongo_username"),
		x.getArtifactIgnoreError(&e, "mongo_password"),
		x.getArtifactIgnoreError(&e, "mongo_database_name"),
	}

	x.events <- ev

	return nil
}

// This is a helper function for x.updateS3 so that we don't have to deal with the
// multi-value return statement when shoving the data back into another array of artifacts
func (x *MongoExtension) getArtifactIgnoreError(e *transistor.Event, artifactName string) transistor.Artifact {
	artifact, err := e.GetArtifact(artifactName)
	if err != nil {
		return transistor.Artifact{}
	}

	return artifact
}

// Clean up the user credentials
// Leave the data in case it's needed at a later date, or its
// just inappropriate to remove the data at the same time
func (x *MongoExtension) deleteMongoExtension(e transistor.Event) error {
	log.Error("deleteMongoExtension")
	if len(e.Artifacts) <= 5 {
		log.Error("Do not do anything as we do not have any valid artifacts to act on")
		return nil
	}

	data, err := x.extractArtifacts(e)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	payload := e.Payload.(plugins.ProjectExtension)
	userName := payload.Project.Slug

	atlasAPI := x.getAtlasClient(data)
	err = x.deleteMongoUser(atlasAPI, data, userName)
	if err != nil {
		log.Error(err.Error())
	}

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

	// MongoEndpoint
	mongoEndpoint, err := e.GetArtifact("mongo_api_endpoint")
	if err != nil {
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.Atlas.APIEndpoint = mongoEndpoint.String()

	// Mongo Public API Key
	mongoAPIPublicKey, err := e.GetArtifact("mongo_api_public_key")
	if err != nil {
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.Atlas.PublicKey = mongoAPIPublicKey.String()

	// Mongo Private API Key
	mongoAPIPrivateKey, err := e.GetArtifact("mongo_api_private_key")
	if err != nil {
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.Atlas.APIKey = mongoAPIPrivateKey.String()

	// Mongo Project ID (The slug from the url in mongo atlas)
	mongoProjectID, err := e.GetArtifact("mongo_project_id")
	if err != nil {
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.Atlas.ProjectID = mongoProjectID.String()

	mongoHostname, err := e.GetArtifact("mongo_hostname")
	if err != nil {
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return nil, err
	}
	data.Hostname = mongoHostname.String()

	return &data, nil
}

// Prepare artifacts array to pass important information back to the front-end
func (x *MongoExtension) buildResultArtifacts(data *MongoData, payloadSlug string, credentials *Credentials) []transistor.Artifact {
	// Stuff new credentials into artifacts to be used by the front-end
	return []transistor.Artifact{
		transistor.Artifact{Key: "mongo_username", Value: credentials.Username, Secret: false},
		transistor.Artifact{Key: "mongo_password", Value: credentials.Password, Secret: false},
		transistor.Artifact{Key: "mongo_host", Value: data.Hostname, Secret: false},
		transistor.Artifact{Key: "mongo_database_name", Value: payloadSlug, Secret: false},
	}
}
