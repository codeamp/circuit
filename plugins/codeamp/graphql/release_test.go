package graphql_resolver_test

import (
	"context"
	"testing"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	graphql "github.com/graph-gophers/graphql-go"
)

type ReleaseTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *ReleaseTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestRelease")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ReleaseTestSuite) TestCreateReleaseSuccess() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackSuccess() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Reset state back to blank
	releaseResolver.DBReleaseResolver.Release.State = ""
	releaseResolver.DBReleaseResolver.Release.StateMessage = "Forced Empty via Test"
	ts.Resolver.DB.Save(&releaseResolver.DBReleaseResolver.Release)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	}
	ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, &releaseInput)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureReleaseInProgress() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureBadProjectID() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID: &releaseID,
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseSuccess() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureNoAuth() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Failure Case - No Authorization
	var ctx context.Context
	_, err = ts.Resolver.StopRelease(ctx, &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
}

func TestSuiteReleaseResolver(t *testing.T) {
	suite.Run(t, new(ReleaseTestSuite))
}
