package mongo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"

	atlas_models "github.com/Clever/atlas-api-client/gen-go/models"
)

func init() {
	transistor.RegisterPlugin("mongo", func() transistor.Plugin {
		return &MongoExtension{MongoAtlasClientNamespacer: &MongoAtlasClientNamespace{}, MongoClientNamespacer: &MongoClientNamespace{}}
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
//	mongo_atlas_api_public_key
//	mongo_atlas_api_private_key
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

// For (mostly) debug purposes, dump the list of mongo users when the 'update' function is called
func (x *MongoExtension) listMongoUsers(atlasAPI MongoAtlasClienter, data *MongoData) (*atlas_models.GetDatabaseUsersResponse, error) {
	resp, err := atlasAPI.GetDatabaseUsers(context.Background(), data.Atlas.ProjectID)
	if err == nil {
		for _, databaseUser := range resp.Results {
			log.Info("User: ", databaseUser.Username)
			log.Info("DB: ", databaseUser.DatabaseName)
			log.Info("Roles (ct): ", len(databaseUser.Roles))
		}
	} else {
		log.Error(err.Error())
		return nil, err
	}

	return resp, nil
}

// Used to query the Mongo Atlas API to determine if the user is already there
// This is used later to verify the user doesn't exist before creating it (and possibly overriding the password)
// it's also used when an extension is created an additional time to prevent it from installing successfully if the user already exits
func (x *MongoExtension) getMongoUser(atlasAPI MongoAtlasClienter, data *MongoData, userName string) (*atlas_models.DatabaseUser, error) {
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

// Used to create a Mongo Atlas user via their API because they do not allow
// users to create other users through the mongo console.
// It creates a user that has write access only to the associated database
// it has read access to the cluster at large to determine the list of databases on the cluster
func (x *MongoExtension) createMongoUser(atlasAPI MongoAtlasClienter, data *MongoData, databaseName string, userName string) (*Credentials, error) {
	generatedPassword, err := x.genRandomAlpha(16)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	credentials := Credentials{
		Username: userName,
		Password: *generatedPassword,
	}

	const DATABASE_ALL_USERS_ARE_REQUIRED_IN = "admin"
	createUserInput := &atlas_models.CreateDatabaseUserInput{
		GroupID: data.Atlas.ProjectID,
		CreateDatabaseUserRequest: &atlas_models.CreateDatabaseUserRequest{
			DatabaseUser: atlas_models.DatabaseUser{
				DatabaseName: DATABASE_ALL_USERS_ARE_REQUIRED_IN,
				GroupID:      data.Atlas.ProjectID,
				Links:        []*atlas_models.Link{},
				Roles: []*atlas_models.Role{
					// Give them readWrite access to the database corresponding to the project
					// the extension has been created for
					&atlas_models.Role{
						CollectionName: "",
						DatabaseName:   databaseName,
						RoleName:       "readWrite",
					},
					// This permission is necessary to list all the database names
					&atlas_models.Role{
						CollectionName: "",
						DatabaseName:   "admin",
						RoleName:       "clusterMonitor",
					},
					// This permission is necessary to list all the database names
					&atlas_models.Role{
						CollectionName: "",
						DatabaseName:   "admin",
						RoleName:       "read",
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

// Delete a mongo user from Mongo Atlas
// This operation does not remove any data, it merely removes
// the credentials that were creataed to access the data
func (x *MongoExtension) deleteMongoUser(atlasAPI MongoAtlasClienter, data *MongoData, userName string) error {
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

// The purpose of this function is to verify that the credentials
// that we have generated are valid and have a modicum of access
// to the requested resources. This is purely to ensure that the
// credentials changes have been proliferated throughout the cluster
func (x *MongoExtension) verifyCredentials(e transistor.Event, data *MongoData, credentials *Credentials, databaseName string) error {
	mongoConnection := fmt.Sprintf("mongodb+srv://%s:%s@%s/%s?authMechanism=SCRAM-SHA-1", credentials.Username, credentials.Password, data.Hostname, databaseName)

	// Ensure we can construct a client interface with no issues
	client, err := x.MongoClientNamespacer.NewClient(options.Client().ApplyURI(mongoConnection))
	if err != nil {
		return err
	}

	startedTime := time.Now()
	currentTime := time.Now()
	waitInterval := 5

	// Provide timeout value and attempt to connect to the mongo database
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(data.CredentialsCheckTimeout)*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Error(err.Error())
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	for {
		log.Debug("Pinging client!")
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Error(err.Error())
		} else {
			if _, err := client.ListDatabaseNames(ctx, bsonx.Doc{}); err != nil {
				log.Error(err.Error())
			} else {
				return nil
			}
		}

		time.Sleep(time.Duration(waitInterval) * time.Second)
		currentTime = time.Now()

		if currentTime.Sub(startedTime) >= (time.Duration(data.CredentialsCheckTimeout) * time.Second) {
			x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("Timed out when verifying permissions (%ds)", data.CredentialsCheckTimeout), nil)
			return errors.New("Timed out when verifying permissions")
		}

		if err := ctx.Err(); err != nil {
			x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
			return err
		}

		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("running"), fmt.Sprintf("Testing Permissions For Mongo (%v elapsed)", currentTime.Sub(startedTime)), nil)
	}

	return nil
}

// Used to construct "random" credentials for created users
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
		return nil, err
	}

	// a is an arbitrary char, no significance other than it being a placeholder
	randString := reg.ReplaceAllString(base64.RawStdEncoding.EncodeToString(b), "a")[:length]

	return &randString, nil
}

// Wrapper for grabbing a new atlas client from the ClientBuilder interface
func (x *MongoExtension) getAtlasClient(data *MongoData) MongoAtlasClienter {
	if x.MongoAtlasClientNamespacer == nil {
		log.Panic("MongoAtlasClientBuilder should NOT be nil!")
	}

	return x.MongoAtlasClientNamespacer.New(data.Atlas.APIEndpoint, data.Atlas.PublicKey, data.Atlas.APIKey)
}

// Creates a mongo extension
// This interacts with the Mongo Atlas API (because thats how you create users)
// It creates the user, then tries to verify the credentials, then reports the extension
// as either success or failure depending on the result of the creds test
func (x *MongoExtension) createMongoExtension(e transistor.Event) error {
	data, err := x.extractArtifacts(e)
	if err != nil {
		log.Error(err.Error())
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	payload := e.Payload.(plugins.ProjectExtension)

	userName := payload.Project.Slug
	databaseName := payload.Project.Slug

	atlasAPI := x.getAtlasClient(data)

	// Check to see if user already exists
	createMongoUser := false
	_, err = x.getMongoUser(atlasAPI, data, userName)
	if err != nil {
		log.Error(err.Error())
		if strings.Contains(err.Error(), "No user with username") {
			createMongoUser = true
		} else {
			log.Error(err.Error())
			x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
			return err
		}
	}

	// Create the mongo user if it doesn't exist
	// otherwise report an error message for possible duplicate extension
	var credentials *Credentials
	if createMongoUser == true {
		credentials, err = x.createMongoUser(atlasAPI, data, databaseName, userName)
		if err != nil {
			log.Error(err.Error())
			x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)

			return err
		}
	} else {
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), "User Exists - Duplicate Extension?", nil)
		return err
	}

	// Take the credentials we just received and try to verify they are actually valid
	// before we send off a message declaring that they are ready for use
	if err := x.verifyCredentials(e, data, credentials, payload.Project.Slug); err != nil {
		log.Error(err.Error())
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
		return err
	}

	x.sendMongoResponse(e, transistor.GetAction("status"),
		transistor.GetState("complete"),
		"Mongo Provisioning Complete.\nRemoving this extension does not delete any data.",
		x.buildResultArtifacts(data, payload.Project.Slug, credentials))

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

	payload := e.Payload.(plugins.ProjectExtension)
	atlasAPI := x.getAtlasClient(data)

	_, err = x.listMongoUsers(atlasAPI, data)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	userName := x.getArtifactIgnoreError(&e, "mongo_username")
	password := x.getArtifactIgnoreError(&e, "mongo_password")
	credentials := Credentials{
		Username: userName.String(),
		Password: password.String(),
	}
	if err := x.verifyCredentials(e, data, &credentials, payload.Project.Slug); err != nil {
		log.Error(err.Error())
		x.sendMongoResponse(e, transistor.GetAction("status"), transistor.GetState("failed"), err.Error(), nil)
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

// Wrapper for sending an event back thruogh the messaging system for standardization and brevity
func (x *MongoExtension) sendMongoResponse(e transistor.Event, action transistor.Action, state transistor.State, stateMessage string, artifacts []transistor.Artifact) {
	event := e.NewEvent(action, state, stateMessage)
	event.Artifacts = artifacts

	x.events <- event
}

// Pull all the artifacts out from the event that we will need
// in order to service these requests. Stuff them into a local storage object.
func (x *MongoExtension) extractArtifacts(e transistor.Event) (*MongoData, error) {
	var data MongoData

	if e.Payload == nil {
		return nil, errors.New("Missing payload!")
	}

	// MongoEndpoint
	mongoEndpoint, err := e.GetArtifact("mongo_atlas_endpoint")
	if err != nil {
		return nil, err
	}
	data.Atlas.APIEndpoint = mongoEndpoint.String()

	// Mongo Public API Key
	mongoAPIPublicKey, err := e.GetArtifact("mongo_atlas_api_public_key")
	if err != nil {
		return nil, err
	}
	data.Atlas.PublicKey = mongoAPIPublicKey.String()

	// Mongo Private API Key
	mongoAPIPrivateKey, err := e.GetArtifact("mongo_atlas_api_private_key")
	if err != nil {
		return nil, err
	}
	data.Atlas.APIKey = mongoAPIPrivateKey.String()

	// Mongo Project ID (The slug from the url in mongo atlas)
	mongoProjectID, err := e.GetArtifact("mongo_atlas_project_id")
	if err != nil {
		return nil, err
	}
	data.Atlas.ProjectID = mongoProjectID.String()

	// The hostname of the actual cluster (not the atlas api endpoint)
	mongoHostname, err := e.GetArtifact("mongo_hostname")
	if err != nil {
		return nil, err
	}
	data.Hostname = mongoHostname.String()

	// Credentials check timeout
	{
		mongoCredentialsCheckTimeout, err := e.GetArtifact("mongo_credentials_check_timeout")
		if err != nil {
			return nil, err
		}

		credentialsCheckTimeout, err := strconv.Atoi(mongoCredentialsCheckTimeout.String())
		if err != nil {
			return nil, err
		}
		data.CredentialsCheckTimeout = credentialsCheckTimeout
	}

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
