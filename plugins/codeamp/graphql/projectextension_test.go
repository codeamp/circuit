package graphql_resolver_test

import (
	"context"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	uuid "github.com/satori/go.uuid"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectExtensionTestSuite struct {
	suite.Suite
	Resolver                 *graphql_resolver.Resolver
	ProjectExtensionResolver *graphql_resolver.ProjectExtensionResolver

	cleanupProjectIDs     []uuid.UUID
	cleanupEnvironmentIDs []uuid.UUID
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

	_ = codeamp.CodeAmp{}
	_ = &graphql_resolver.Resolver{DB: db, Events: nil, Redis: nil}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.ProjectExtensionResolver = &graphql_resolver.ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: db}}
}

func (ts *ProjectExtensionTestSuite) Test1CreateProjectSuccess() {
	// setup
	env := model.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	ts.Resolver.DB.Create(&env)
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, env.Model.ID)

	modelID := env.Model.ID.String()
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: &modelID,
	}

	createProjectResolver, err := ts.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	ts.cleanupProjectIDs = append(ts.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.ID)
}

func (ts *ProjectExtensionTestSuite) Test2CreateProjectFailure() {
	// setup
	env := model.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	ts.Resolver.DB.Create(&env)
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, env.Model.ID)

	modelID := env.Model.ID.String()
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: &modelID,
	}

	createProjectResolver, err := ts.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	ts.cleanupProjectIDs = append(ts.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.ID)

	_, err = ts.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})

	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) Test3CreateProjectNoAuth() {
	// setup
	env := model.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	ts.Resolver.DB.Create(&env)
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, env.Model.ID)

	modelID := env.Model.ID.String()
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: &modelID,
	}

	var ctx context.Context
	_, err := ts.Resolver.CreateProject(ctx, &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})

	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) Test2GormCreateProjectExtension() {

}

func (ts *ProjectExtensionTestSuite) TearDownTest() {
	for _, id := range ts.cleanupProjectIDs {
		ts.Resolver.DB.Unscoped().Delete(&model.Project{Model: model.Model{ID: id}})
	}

	for _, id := range ts.cleanupEnvironmentIDs {
		ts.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}})
	}

}

func TestSuiteProjectExtensionResolver(t *testing.T) {
	ts := new(ProjectExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
