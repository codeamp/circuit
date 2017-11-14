package resolvers_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestFeatures struct {
	suite.Suite
	db *gorm.DB
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
	)

	suite.db = db
}

func (suite *TestFeatures) TearDownSuite() {
	spew.Dump("dropping test db")
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from features;")
}

func (suite *TestFeatures) TestSuccessfulFeature() {
	spew.Dump("TestSuccessfulGetFeature")

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

	testFeature := models.Feature{
		Message:    "test",
		User:       "testuser",
		Hash:       "testhash",
		ParentHash: "testparenthash",
		Ref:        "testref",
		Created:    time.Now(),
		ProjectId:  testProject.Model.ID,
	}
	suite.db.Save(&testFeature)

	testFeatureResolver := resolvers.NewFeatureResolver(testFeature, suite.db)
	projectResolver, _ := testFeatureResolver.Project(context.TODO())

	assert.Equal(suite.T(), graphql.ID(testFeature.ID.String()), testFeatureResolver.ID())
	assert.Equal(suite.T(), testFeature.Message, testFeatureResolver.Message())
	assert.Equal(suite.T(), testFeature.User, testFeatureResolver.User())
	assert.Equal(suite.T(), testFeature.Hash, testFeatureResolver.Hash())
	assert.Equal(suite.T(), testFeature.ParentHash, testFeatureResolver.ParentHash())
	assert.Equal(suite.T(), testFeature.Ref, testFeatureResolver.Ref())
	assert.Equal(suite.T(), graphql.Time{Time: testFeature.Created}, testFeatureResolver.Created())
	assert.Equal(suite.T(), reflect.TypeOf(&resolvers.ProjectResolver{}), reflect.TypeOf(projectResolver))
}

func TestFeatureResolvers(t *testing.T) {
	suite.Run(t, new(TestFeatures))
}
