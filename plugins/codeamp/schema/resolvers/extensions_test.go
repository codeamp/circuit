package resolvers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestExtensions struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
}

func (suite *TestExtensions) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"15432",
		"postgres",
		"codeamp_test",
		"disable",
		"",
	))

	db.Exec(fmt.Sprintf("CREATE DATABASE %s", "codeamp_test"))
	db.Exec("CREATE EXTENSION \"uuid-ossp\"")
	db.Exec("CREATE EXTENSION IF NOT EXISTS hstore")

	db.AutoMigrate(
		&models.Project{},
		&models.Feature{},
		&models.Release{},
		&models.Service{},
		&models.EnvironmentVariable{},
		&models.Extension{},
		&models.ExtensionSpec{},
	)

	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return codeamp.NewCodeAmp() })
	t, _ := transistor.NewTestTransistor(transistor.Config{
		Server:    "http://127.0.0.1:16379",
		Password:  "",
		Database:  "0",
		Namespace: "",
		Pool:      "30",
		Process:   "1",
		Plugins: map[string]interface{}{
			"codeamp": map[string]interface{}{
				"workers": 1,
				"postgres": map[string]interface{}{
					"host":     "0.0.0.0",
					"port":     "15432",
					"user":     "postgres",
					"dbname":   "codeamp_test",
					"sslmode":  "disable",
					"password": "",
				},
			},
		},
		EnabledPlugins: []string{},
		Queueing:       false,
	})

	actions := actions.NewActions(t.TestEvents, db)

	suite.db = db
	suite.t = t
	suite.actions = actions
}

func (suite *TestExtensions) TearDownSuite() {
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from features;")
	suite.db.Exec("delete from releases;")
	suite.db.Exec("delete from services;")
	suite.db.Exec("delete from environment_variables;")
	suite.db.Exec("delete from extensions;")
	suite.db.Exec("delete from extension_specs;")
}

func (suite *TestExtensions) TestSuccessfulCreateExtension() {
	stamp := strings.ToLower("TestSuccessfulCreateExtension")
	timestamp := time.Now()

	es := models.ExtensionSpec{
		Type:      plugins.Workflow,
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec:  map[string]*string{},
		Created:   timestamp,
	}
	suite.db.Save(&es)

	p := models.Project{
		Name:          fmt.Sprintf("test%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p)

	extInput := struct {
		Extension *resolvers.ExtensionInput
	}{
		Extension: &resolvers.ExtensionInput{
			ProjectId:       p.Model.ID.String(),
			ExtensionSpecId: es.Model.ID.String(),
			FormSpecValues: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key1",
					Value: "key2",
				},
			},
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	extResolver, _ := resolver.CreateExtension(context.TODO(), &extInput)

	projectResolver, _ := extResolver.Project(context.TODO())
	extensionSpecResolver, _ := extResolver.ExtensionSpec(context.TODO())
	formSpecValues, _ := extResolver.FormSpecValues(context.TODO())

	assert.Equal(suite.T(), fmt.Sprintf("test%s", stamp), projectResolver.Name())
	assert.Equal(suite.T(), fmt.Sprintf("test%s", stamp), extensionSpecResolver.Name())
	assert.Equal(suite.T(), "key1", formSpecValues[0].Key())
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidExtensionSpecId() {
	stamp := strings.ToLower("TestFailedCreateExtensionInvalidExtensionSpecId")

	p := models.Project{
		Name:          fmt.Sprintf("test%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p)

	extInput := struct {
		Extension *resolvers.ExtensionInput
	}{
		Extension: &resolvers.ExtensionInput{
			ProjectId:       p.Model.ID.String(),
			ExtensionSpecId: "invaliduuidfake",
			FormSpecValues: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key1",
					Value: "key2",
				},
			},
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Could not parse ExtensionSpecId. Invalid Format.", err.Error())
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidProjectId() {
	stamp := strings.ToLower("TestFailedCreateExtensionInvalidProjectId")

	timestamp := time.Now()

	es := models.ExtensionSpec{
		Type:      plugins.Workflow,
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec:  map[string]*string{},
		Created:   timestamp,
	}
	suite.db.Save(&es)

	extInput := struct {
		Extension *resolvers.ExtensionInput
	}{
		Extension: &resolvers.ExtensionInput{
			ProjectId:       "invalidfakeuuid",
			ExtensionSpecId: es.Model.ID.String(),
			FormSpecValues: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key1",
					Value: "key2",
				},
			},
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Could not parse ProjectId. Invalid format.", err.Error())
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidFormSpecValues() {
	stamp := strings.ToLower("TestFailedCreateExtensionInvalidFormSpecValues")
	timestamp := time.Now()

	requiredStringParam := "required|string"

	es := models.ExtensionSpec{
		Type:      plugins.Workflow,
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec: map[string]*string{
			"key1": &requiredStringParam,
			"key3": &requiredStringParam,
		},
		Created: timestamp,
	}
	suite.db.Save(&es)

	p := models.Project{
		Name:          fmt.Sprintf("test%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p)

	// missing key3
	extInput := struct {
		Extension *resolvers.ExtensionInput
	}{
		Extension: &resolvers.ExtensionInput{
			ProjectId:       p.Model.ID.String(),
			ExtensionSpecId: es.Model.ID.String(),
			FormSpecValues: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key1",
					Value: "val",
				},
				plugins.KeyValue{
					Key:   "key2",
					Value: "val2",
				},
			},
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Required keys not found within extension input: [key3]", err.Error())
}

func (suite *TestExtensions) TestFailedCreateExtensionExtensionSpecDoesntExist() {
	stamp := strings.ToLower("TestFailedCreateExtensionExtensionSpecDoesntExist")

	p := models.Project{
		Name:          fmt.Sprintf("test%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p)

	fakeExtensionSpecId := uuid.NewV1().String()

	extInput := struct {
		Extension *resolvers.ExtensionInput
	}{
		Extension: &resolvers.ExtensionInput{
			ProjectId:       p.Model.ID.String(),
			ExtensionSpecId: fakeExtensionSpecId,
			FormSpecValues: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key1",
					Value: "key2",
				},
			},
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Can't find corresponding extensionSpec.", err.Error())
}

func TestExtensionResolvers(t *testing.T) {
	suite.Run(t, new(TestExtensions))
}
