package resolvers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
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
	user    models.User
	context context.Context
}

func (suite *TestExtensions) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"4500",
		"postgres",
		"codeamp_test",
		"disable",
		"",
	))

	db.Exec(fmt.Sprintf("CREATE DATABASE %s", "codeamp_test"))
	db.Exec("CREATE EXTENSION \"uuid-ossp\"")
	db.Exec("CREATE EXTENSION IF NOT EXISTS hstore")

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
					"port":     "5432",
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

func (suite *TestExtensions) SetupDBAndContext() {
	suite.db.AutoMigrate(
		&models.Project{},
		&models.Feature{},
		&models.Release{},
		&models.Service{},
		&models.EnvironmentVariable{},
		&models.Extension{},
		&models.ExtensionSpec{},
		&models.User{},
		&models.Environment{},
	)

	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	suite.db.Save(&user)

	suite.context = context.WithValue(suite.context, "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
}

func (suite *TestExtensions) TearDownSuite() {
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from features;")
	suite.db.Exec("delete from releases;")
	suite.db.Exec("delete from services;")
	suite.db.Exec("delete from environment_variables;")
	suite.db.Exec("delete from extensions;")
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from extension_specs;")
}

func (suite *TestExtensions) TestSuccessfulCreateExtension() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulCreateExtension")

	es := models.ExtensionSpec{
		Type:      plugins.GetType("workflow"),
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec:  map[string]*string{},
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

	env := models.Environment{
		Name: fmt.Sprintf("env%s", stamp),
	}
	suite.db.Save(&env)

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
			EnvironmentId: env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	extResolver, _ := resolver.CreateExtension(context.TODO(), &extInput)

	envResolver, _ := extResolver.Environment(context.TODO())
	projectResolver, _ := extResolver.Project(context.TODO())
	extensionSpecResolver, _ := extResolver.ExtensionSpec(context.TODO())
	formSpecValues, _ := extResolver.FormSpecValues(context.TODO())

	assert.Equal(suite.T(), fmt.Sprintf("env%s", stamp), envResolver.Name(context.TODO()))
	assert.Equal(suite.T(), fmt.Sprintf("test%s", stamp), projectResolver.Name())
	assert.Equal(suite.T(), fmt.Sprintf("test%s", stamp), extensionSpecResolver.Name())
	assert.Equal(suite.T(), "key1", formSpecValues[0].Key())

	suite.TearDownSuite()
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidExtensionSpecId() {
	suite.SetupDBAndContext()
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
	suite.TearDownSuite()
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidProjectId() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestFailedCreateExtensionInvalidProjectId")

	es := models.ExtensionSpec{
		Type:      plugins.GetType("workflow"),
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec:  map[string]*string{},
	}
	suite.db.Save(&es)

	env := models.Environment{
		Name: fmt.Sprintf("env%s", stamp),
	}
	suite.db.Save(&env)

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
			EnvironmentId: env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Could not parse ProjectId. Invalid format.", err.Error())
	suite.TearDownSuite()
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidFormSpecValues() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestFailedCreateExtensionInvalidFormSpecValues")

	requiredStringParam := "required|string"

	es := models.ExtensionSpec{
		Type:      plugins.GetType("workflow"),
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec: map[string]*string{
			"key1": &requiredStringParam,
			"key3": &requiredStringParam,
		},
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
	env := models.Environment{
		Name: fmt.Sprintf("env%s", stamp),
	}
	suite.db.Save(&env)

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
			EnvironmentId: env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Required keys not found within extension input: [key3]", err.Error())
	suite.TearDownSuite()
}

func (suite *TestExtensions) TestFailedCreateExtensionExtensionSpecDoesntExist() {
	suite.SetupDBAndContext()
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

	env := models.Environment{
		Name: fmt.Sprintf("env%s", stamp),
	}
	suite.db.Save(&env)

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
			EnvironmentId: env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Can't find corresponding extensionSpec.", err.Error())
	suite.TearDownSuite()
}

func (suite *TestExtensions) TestFailedCreateExtensionInvalidEnvironmentId() {
	suite.SetupDBAndContext()
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

	requiredStringParam := "required|string"

	es := models.ExtensionSpec{
		Type:      plugins.GetType("workflow"),
		Key:       fmt.Sprintf("testkey%s", stamp),
		Name:      fmt.Sprintf("test%s", stamp),
		Component: fmt.Sprintf("testcomponent%s", stamp),
		FormSpec: map[string]*string{
			"key1": &requiredStringParam,
			"key3": &requiredStringParam,
		},
	}
	suite.db.Save(&es)

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
			EnvironmentId: "invalidenvironmentid",
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtension(context.TODO(), &extInput)

	assert.Equal(suite.T(), "Could not parse EnvironmentId. Invalid format.", err.Error())
	suite.TearDownSuite()
}

func TestExtensionResolvers(t *testing.T) {
	suite.Run(t, new(TestExtensions))
}
