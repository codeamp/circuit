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
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestReleaseExtension struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
	user    models.User
	context context.Context
}

func (suite *TestReleaseExtension) SetupSuite() {

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

func (suite *TestReleaseExtension) SetupDBAndContext() {
	suite.db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.ReleaseExtension{},
		&models.Release{},
		&models.Feature{},
		&models.Extension{},
		&models.ExtensionSpec{},
		&models.Project{},
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

func (suite *TestReleaseExtension) TearDownSuite() {
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from user_permissions;")
	suite.db.Exec("delete from release_extensions;")
	suite.db.Exec("delete from extension_specs;")
	suite.db.Exec("delete from features;")
	suite.db.Exec("delete from extensions;")
	suite.db.Exec("delete from releases;")
}

func (suite *TestReleaseExtension) TestReleaseExtensions() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestReleaseExtensions")

	project := models.Project{
		Name:          fmt.Sprintf("testname %s", time.Now().String()),
		Slug:          fmt.Sprintf("testslug %s", time.Now().String()),
		Repository:    "testrepository",
		Secret:        "testsecret",
		GitUrl:        "testgiturl",
		GitProtocol:   "testgitprotocol",
		RsaPrivateKey: "testrsaprivatekey",
		RsaPublicKey:  "testrsapublickey",
	}
	suite.db.Save(&project)

	headFeature := models.Feature{
		Message:    "test",
		User:       "testuser",
		Hash:       "testhash1",
		ParentHash: "testparenthash",
		Ref:        "testref",
		Created:    time.Now(),
		ProjectId:  project.Model.ID,
	}
	suite.db.Save(&headFeature)

	tailFeature := models.Feature{
		Message:    "test",
		User:       "testuser",
		Hash:       "testhash2",
		ParentHash: "testparenthash",
		Ref:        "testref",
		Created:    time.Now(),
		ProjectId:  project.Model.ID,
	}
	suite.db.Save(&tailFeature)

	release := models.Release{
		ProjectId:     project.Model.ID,
		UserID:        suite.user.Model.ID,
		HeadFeatureID: headFeature.Model.ID,
		TailFeatureID: tailFeature.Model.ID,
		StateMessage:  "statemessage",
	}
	suite.db.Save(&release)

	extensionSpec := models.ExtensionSpec{
		Type: plugins.GetType("workflow"),
		Key:  fmt.Sprintf("releaseextension%s", stamp),
	}
	suite.db.Save(&extensionSpec)

	extension := models.Extension{
		ProjectId:       project.Model.ID,
		ExtensionSpecId: extensionSpec.Model.ID,
		State:           plugins.GetState("waiting"),
		Artifacts:       map[string]*string{},
		FormSpecValues:  postgres.Jsonb{},
	}
	suite.db.Save(&extension)

	re := models.ReleaseExtension{
		ReleaseId:         release.Model.ID,
		FeatureHash:       fmt.Sprintf("featurehash%s", stamp),
		ServicesSignature: fmt.Sprintf("servicessignature%s", stamp),
		SecretsSignature:  fmt.Sprintf("secretssignature%s", stamp),
		ExtensionId:       extension.Model.ID,
		State:             plugins.GetState("waiting"),
		StateMessage:      "testmessage",
		Type:              plugins.GetType("workflow"),
		Artifacts:         map[string]*string{},
		Finished:          time.Now(),
	}

	suite.db.Save(&re)

	re2 := models.ReleaseExtension{
		ReleaseId:         release.Model.ID,
		FeatureHash:       fmt.Sprintf("featurehash2%s", stamp),
		ServicesSignature: fmt.Sprintf("servicessignature2%s", stamp),
		SecretsSignature:  fmt.Sprintf("secretssignature2%s", stamp),
		ExtensionId:       extension.Model.ID,
		State:             plugins.GetState("waiting"),
		StateMessage:      "testmessage",
		Type:              plugins.GetType("workflow"),
		Artifacts:         map[string]*string{},
		Finished:          time.Now(),
	}

	suite.db.Save(&re2)

	res := []models.ReleaseExtension{
		re2, re,
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	reResolvers, _ := resolver.ReleaseExtensions(suite.context)
	assert.Equal(suite.T(), 2, len(reResolvers))

	for idx, reResolver := range reResolvers {
		assert.Equal(suite.T(), res[idx].FeatureHash, reResolver.FeatureHash())
		assert.Equal(suite.T(), res[idx].ServicesSignature, reResolver.ServicesSignature())
		assert.Equal(suite.T(), res[idx].SecretsSignature, reResolver.SecretsSignature())

		extensionResolver, _ := reResolver.Extension(suite.context)
		assert.Equal(suite.T(), res[idx].ExtensionId.String(), string(extensionResolver.ID()))
		assert.Equal(suite.T(), string(res[idx].State), reResolver.State())
	}

	suite.TearDownSuite()
}

func TestReleaseExtensionResolvers(t *testing.T) {
	suite.Run(t, new(TestReleaseExtension))
}
