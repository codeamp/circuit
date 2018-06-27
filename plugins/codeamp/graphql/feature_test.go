package graphql_resolver_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	Resolver        *graphql_resolver.Resolver
	FeatureResolver *graphql_resolver.FeatureResolver

	cleanupProjectIDs     []uuid.UUID
	cleanupFeatureIDs     []uuid.UUID
	cleanupEnvironmentIDs []uuid.UUID
}

func (suite *FeatureTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Feature{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.FeatureResolver = &graphql_resolver.FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: db}}
}

func (suite *FeatureTestSuite) TestCreateFeature() {
	// Environment
	envInput := model.EnvironmentInput{
		Name:      "TestCreateFeature",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

	// Project
	envId := fmt.Sprintf("%v", envResolver.DBEnvironmentResolver.Environment.Model.ID)
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: &envId,
	}

	createProjectResolver, err := suite.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		log.Fatal(err.Error())
	}
	suite.cleanupProjectIDs = append(suite.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.Model.ID)
	projectID := string(createProjectResolver.ID())

	projectIDUUID, err := uuid.FromString(strings.ToUpper(projectID))
	assert.Nil(suite.T(), err)

	feature := model.Feature{
		ProjectID:  projectIDUUID,
		Message:    "A test feature message",
		User:       "TestCreateFeature",
		Hash:       "42941a0900e952f7f78994d53b699aea23926804",
		ParentHash: "",
		Ref:        "refs/heads/master",
		Created:    time.Now(),
	}

	db := suite.Resolver.DB.Create(&feature)
	if db.Error != nil {
		assert.FailNow(suite.T(), db.Error.Error())
	}
	suite.cleanupFeatureIDs = append(suite.cleanupFeatureIDs, feature.Model.ID)

	var ctx context.Context
	_, err = suite.Resolver.Features(ctx)
	assert.NotNil(suite.T(), err)

	featureResolvers, err := suite.Resolver.Features(test.ResolverAuthContext())
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), featureResolvers)

	featureResolver := featureResolvers[0]
	assert.NotNil(suite.T(), featureResolver)

	_ = featureResolver.ID()
	_ = featureResolver.Project()
	message := featureResolver.Message()
	assert.Equal(suite.T(), feature.Message, message)

	user := featureResolver.User()
	assert.Equal(suite.T(), feature.User, user)

	hash := featureResolver.Hash()
	assert.Equal(suite.T(), feature.Hash, hash)

	parentHash := featureResolver.ParentHash()
	assert.Equal(suite.T(), feature.ParentHash, parentHash)

	ref := featureResolver.Ref()
	assert.Equal(suite.T(), feature.Ref, ref)

	created_at_diff := time.Now().Sub(featureResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(suite.T(), "Created at time is too old")
	}

	data, err := featureResolver.MarshalJSON()
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), data)

	err = featureResolver.UnmarshalJSON(data)
	assert.Nil(suite.T(), err)
}

func (suite *FeatureTestSuite) TearDownTest() {
	for _, id := range suite.cleanupFeatureIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Feature{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	suite.cleanupFeatureIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupProjectIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Project{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
	suite.cleanupProjectIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupEnvironmentIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
	suite.cleanupEnvironmentIDs = make([]uuid.UUID, 0)
}

func TestSuiteFeatureResolver(t *testing.T) {
	ts := new(FeatureTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
