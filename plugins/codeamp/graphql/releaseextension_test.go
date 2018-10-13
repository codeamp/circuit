package graphql_resolver_test

import (
	"context"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReleaseExtensionTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
	helper   Helper
}

func (suite *ReleaseExtensionTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectEnvironment{},
		&model.ProjectBookmark{},
		&model.UserPermission{},
		&model.ProjectSettings{},
		&model.Environment{},
		&model.Extension{},
		&model.ProjectExtension{},
		&model.Extension{},
		&model.ReleaseExtension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestReleaseExtension")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ReleaseExtensionTestSuite) TestReleaseExtensionInterface() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, _ := ts.helper.CreateProject(ts.T(), environmentResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Features
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Remove Autocreated ReleaseExtension from Release
	for _, re := range releaseResolver.ReleaseExtensions() {
		ts.Resolver.DB.Unscoped().Delete(&re.DBReleaseExtensionResolver.ReleaseExtension)
	}

	// Release Extension
	releaseExtensionResolver := ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	_ = releaseResolver.ID()
	_ = releaseResolver.Project()
	_ = releaseResolver.User()

	var ctx context.Context
	artifacts, err := releaseResolver.Artifacts(ctx)
	assert.NotNil(ts.T(), err)
	log.Error(err)

	artifacts, err = releaseResolver.Artifacts(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), artifacts)

	assert.Equal(ts.T(), "42941a0900e952f7f78994d53b699aea23926804", string(releaseResolver.HeadFeature().Hash()))

	_ = releaseResolver.TailFeature()
	assert.Equal(ts.T(), "waiting", releaseResolver.State())
	assert.Equal(ts.T(), "Release created", releaseResolver.StateMessage())

	environment, err := releaseResolver.Environment()
	assert.Nil(ts.T(), err)
	if environment == nil {
		assert.FailNow(ts.T(), "No Environment")
	}
	// assert.Equal(ts.T(), envId, environment.ID())

	created_at_diff := time.Now().Sub(releaseResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	data, err := releaseResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = releaseResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)

	// Test Release Extension Interface
	releaseExtensionResolvers := releaseResolver.ReleaseExtensions()
	assert.NotNil(ts.T(), releaseExtensionResolvers)
	assert.NotEmpty(ts.T(), releaseExtensionResolvers)

	releaseExtensionResolver = releaseExtensionResolvers[0]
	_ = releaseExtensionResolver.ID()
	_, err = releaseExtensionResolver.Release()
	_, err = releaseExtensionResolver.Extension()

	assert.Equal(ts.T(), "ServicesSignature", releaseExtensionResolver.ServicesSignature())
	assert.Equal(ts.T(), "SecretsSignature", releaseExtensionResolver.SecretsSignature())

	assert.Equal(ts.T(), "waiting", releaseExtensionResolver.State())
	assert.Equal(ts.T(), "workflow", releaseExtensionResolver.Type())
	assert.Equal(ts.T(), "TestReleaseExtension", releaseExtensionResolver.StateMessage())

	_ = releaseExtensionResolver.Artifacts()
	_ = releaseExtensionResolver.Finished()
	_ = releaseExtensionResolver.Started()

	created_at_diff = time.Now().Sub(releaseExtensionResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	data, err = releaseExtensionResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = releaseExtensionResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)

	// Test Query Interface
	// var ctx context.Context
	_, err = ts.Resolver.ReleaseExtensions(ctx)
	assert.NotNil(ts.T(), err)

	releaseExtensions, err := ts.Resolver.ReleaseExtensions(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), releaseExtensions)
	assert.NotEmpty(ts.T(), releaseExtensions)
}

func (suite *ReleaseExtensionTestSuite) TearDownTest() {
	suite.helper.TearDownTest(suite.T())
	suite.Resolver.DB.Close()
}

func TestSuiteReleaseExtensionResolver(t *testing.T) {
	suite.Run(t, new(ReleaseExtensionTestSuite))
}
