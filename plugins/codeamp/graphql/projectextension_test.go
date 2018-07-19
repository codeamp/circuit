package graphql_resolver_test

import (
	"context"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectExtensionTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *ProjectExtensionTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectEnvironment{},
		&model.ProjectBookmark{},
		&model.UserPermission{},
		&model.ProjectSettings{},
		&model.Environment{},
		&model.Extension{},
		&model.ProjectExtension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestProjectExtension")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ProjectExtensionTestSuite) TestCreateProjectExtensionSuccess() {
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
	ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
}

func (ts *ProjectExtensionTestSuite) TestCreateProjectExtensionFailure() {
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
	var ctx context.Context
	ts.helper.SetContext(ctx)

	_, err = ts.helper.CreateProjectExtensionWithError(ts.T(), extensionResolver, projectResolver)
	assert.NotNil(ts.T(), err)

	ts.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ProjectExtensionTestSuite) TestCreateProjectExtensionEnvIDFailure() {
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
	projectResolver.DBProjectResolver.Environment.Model.ID, _ = uuid.FromString("123e4567-e89b-12d3-a456-426655440000")
	_, err = ts.helper.CreateProjectExtensionWithError(ts.T(), extensionResolver, projectResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestProjectExtensionInterface() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Project Extension Interface
	_ = projectExtensionResolver.ID()

	assert.Equal(ts.T(), projectResolver.ID(), projectExtensionResolver.Project().ID())
	assert.Equal(ts.T(), extensionResolver.ID(), projectExtensionResolver.Extension().ID())

	_ = projectExtensionResolver.Artifacts()

	// assert.Equal(ts.T(), model.JSON{[]byte("[]")}, projectExtensionResolver.Config())
	// assert.Equal(ts.T(), model.JSON{[]byte("{}")}, projectExtensionResolver.CustomConfig())

	_ = projectExtensionResolver.State()
	_ = projectExtensionResolver.StateMessage()

	environment, err := projectExtensionResolver.Environment()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), environment)
	assert.Equal(ts.T(), environmentResolver.ID(), environment.ID())

	created_at_diff := time.Now().Sub(projectExtensionResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	data, err := projectExtensionResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = projectExtensionResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestProjectExtensionExtractArtifacts() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	graphql_resolver.ExtractArtifacts(projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension,
		extensionResolver.DBExtensionResolver.Extension, ts.Resolver.DB)
}

func (ts *ProjectExtensionTestSuite) TestProjectExtensionQuery() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Test ProjectExtension Query Interface
	var ctx context.Context
	_, err = ts.Resolver.ProjectExtensions(ctx)
	assert.NotNil(ts.T(), err)

	projectExtensionResolvers, err := ts.Resolver.ProjectExtensions(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), projectExtensionResolvers)
	assert.NotEmpty(ts.T(), projectExtensionResolvers)
}

func (ts *ProjectExtensionTestSuite) TestUpdateProjectExtensionSuccess() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Update Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID: &projectExtensionID,
	}
	_, err = ts.Resolver.UpdateProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ProjectExtensionTestSuite) TestUpdateProjectExtensionFailureProjectExtensionNotFound() {
	// Update Project Extension
	projectExtensionID := test.ValidUUID
	projectExtensionInput := model.ProjectExtensionInput{
		ID: &projectExtensionID,
	}

	_, err := ts.Resolver.UpdateProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestUpdateProjectExtensionFailureExtensionNotFound() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Update Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID:          &projectExtensionID,
		ExtensionID: test.ValidUUID,
	}
	_, err = ts.Resolver.UpdateProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestUpdateProjectExtensionFailureEnvironmentNotFound() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Update Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID:            &projectExtensionID,
		ExtensionID:   string(extensionResolver.ID()),
		EnvironmentID: test.ValidUUID,
	}
	_, err = ts.Resolver.UpdateProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestUpdateProjectExtensionFailureProjectNotFound() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Update Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID:            &projectExtensionID,
		ExtensionID:   string(extensionResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
		ProjectID:     test.ValidUUID,
	}
	_, err = ts.Resolver.UpdateProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestDeleteProjectExtensionFailureNoProjectID() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Update Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID:          &projectExtensionID,
		ExtensionID: string(extensionResolver.ID()),
		ProjectID:   test.ValidUUID,
	}
	_, err = ts.Resolver.DeleteProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestDeleteProjectExtensionSuccess() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Delete Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID: &projectExtensionID,
	}
	_, err = ts.Resolver.DeleteProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ProjectExtensionTestSuite) TestDeleteProjectExtensionFailureBadEnvironmentID() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Delete Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID:            &projectExtensionID,
		ExtensionID:   string(extensionResolver.ID()),
		EnvironmentID: test.ValidUUID,
	}

	_, err = ts.Resolver.DeleteProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestDeleteProjectExtensionFailureBadExtensionID() {
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
	projectExtension := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Delete Project Extension
	projectExtensionID := string(projectExtension.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ID:          &projectExtensionID,
		ExtensionID: test.ValidUUID,
	}

	_, err = ts.Resolver.DeleteProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TestDeleteProjectExtensionFailureNoProjectExtensionID() {
	// Delete Project Extension
	projectExtensionID := test.ValidUUID
	projectExtensionInput := model.ProjectExtensionInput{
		ID: &projectExtensionID,
	}
	_, err := ts.Resolver.DeleteProjectExtension(&struct{ ProjectExtension *model.ProjectExtensionInput }{&projectExtensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteProjectExtensionResolver(t *testing.T) {
	suite.Run(t, new(ProjectExtensionTestSuite))
}
