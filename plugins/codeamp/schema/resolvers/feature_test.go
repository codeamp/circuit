package resolvers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestFeatures struct {
	suite.Suite
	db      *gorm.DB
	context context.Context
	t       *transistor.Transistor
	actions *actions.Actions
}

func (suite *TestFeatures) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"15432",
		"postgres",
		"codeamp_test",
		"disable",
		"",
	))

	db.Exec("CREATE DATABASE codeamp_test")
	db.Exec("CREATE EXTENSION \"uuid-ossp\"")
	db.Exec("CREATE EXTENSION IF NOT EXISTS hstore")

	db.AutoMigrate(
		&models.Feature{},
		&models.Project{},
		&models.User{},
	)

	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	db.Save(&user)

	// initialize transistor
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

	suite.db = db
	suite.t = t
	suite.actions = actions.NewActions(t.TestEvents, db)
	suite.context = context.WithValue(context.TODO(), "jwt", utils.Claims{UserId: user.Model.ID.String()})
}

func (suite *TestFeatures) TearDownSuite() {
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from features;")
}

func (suite *TestFeatures) TestSuccessfulFeature() {

	testProject := models.Project{
		Name:          fmt.Sprintf("testname %s", time.Now().String()),
		Slug:          fmt.Sprintf("testslug %s", time.Now().String()),
		Repository:    "testrepository",
		Secret:        "testsecret",
		GitUrl:        "testgiturl",
		GitProtocol:   "testgitprotocol",
		RsaPrivateKey: "testrsaprivatekey",
		RsaPublicKey:  "testrsapublickey",
	}

	suite.db.Save(&testProject)

	testFeature1 := models.Feature{
		Message:    "test",
		User:       "testuser",
		Hash:       "testhash1",
		ParentHash: "testparenthash",
		Ref:        "testref",
		Created:    time.Now(),
		ProjectId:  testProject.Model.ID,
	}
	suite.db.Save(&testFeature1)

	testFeature2 := models.Feature{
		Message:    "test",
		User:       "testuser",
		Hash:       "testhash2",
		ParentHash: "testparenthash",
		Ref:        "testref",
		Created:    time.Now(),
		ProjectId:  testProject.Model.ID,
	}
	suite.db.Save(&testFeature2)

	testFeature3 := models.Feature{
		Message:    "test",
		User:       "testuser",
		Hash:       "testhash3",
		ParentHash: "testparenthash",
		Ref:        "testref",
		Created:    time.Now(),
		ProjectId:  testProject.Model.ID,
	}
	suite.db.Save(&testFeature3)

	// features() method is "created_at desc" so last created -> first
	testFeatures := []models.Feature{testFeature3, testFeature2, testFeature1}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	testFeatureResolvers, _ := resolver.Features(suite.context)

	assert.Equal(suite.T(), 3, len(testFeatureResolvers))
	for idx, featureResolver := range testFeatureResolvers {
		assert.Equal(suite.T(), graphql.ID(testFeatures[idx].Model.ID.String()), featureResolver.ID())
		assert.Equal(suite.T(), testFeatures[idx].Message, featureResolver.Message())
		assert.Equal(suite.T(), testFeatures[idx].User, featureResolver.User())
		assert.Equal(suite.T(), testFeatures[idx].Hash, featureResolver.Hash())
		assert.Equal(suite.T(), testFeatures[idx].ParentHash, featureResolver.ParentHash())
		assert.Equal(suite.T(), testFeatures[idx].Ref, featureResolver.Ref())
		assert.Equal(suite.T(), graphql.Time{Time: testFeatures[idx].Created}.Unix(), featureResolver.Created().Unix())

		projectResolver, _ := featureResolver.Project(suite.context)
		assert.Equal(suite.T(), testProject.Name, projectResolver.Name())

	}
}

func TestFeatureResolvers(t *testing.T) {
	suite.Run(t, new(TestFeatures))
}
