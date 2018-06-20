package graphql_resolver_test

import (
	"encoding/json"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	log "github.com/codeamp/logger"

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
	EnvironmentResolver      *graphql_resolver.EnvironmentResolver

	cleanupProjectIDs          []uuid.UUID
	cleanupProjectExtensionIDs []uuid.UUID
	cleanupEnvironmentIDs      []uuid.UUID
	cleanupExtensionIDs        []uuid.UUID
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
	suite.EnvironmentResolver = &graphql_resolver.EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: db}}
}

func (ts *ProjectExtensionTestSuite) createTestEnvironment(useAuth bool) (*graphql_resolver.EnvironmentResolver, error) {
	// setup
	envInput := model.EnvironmentInput{
		Name:      "projectextensiontest",
		Color:     "purple",
		Key:       "projectextensiontest",
		IsDefault: true,
	}

	ctx := test.ResolverAuthContext()
	if useAuth == false {
		ctx = nil
	}
	envResolver, err := ts.Resolver.CreateEnvironment(ctx, &struct{ Environment *model.EnvironmentInput }{Environment: &envInput})
	if err == nil {
		ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)
	}

	return envResolver, err
}

func (ts *ProjectExtensionTestSuite) createTestProject(useAuth bool, envResolver *graphql_resolver.EnvironmentResolver) (*graphql_resolver.ProjectResolver, error) {
	environmentID := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: &environmentID,
	}

	ctx := test.ResolverAuthContext()
	if useAuth == false {
		ctx = nil
	}
	createProjectResolver, err := ts.Resolver.CreateProject(ctx, &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})
	if err == nil {
		log.Warn("Adding ID: ", createProjectResolver.DBProjectResolver.Project.ID)
		ts.cleanupProjectIDs = append(ts.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.ID)
	}
	return createProjectResolver, err
}

func (ts *ProjectExtensionTestSuite) createTestExtension(useAuth bool, envResolver *graphql_resolver.EnvironmentResolver) (*graphql_resolver.ExtensionResolver, error) {
	// ctx := test.ResolverAuthContext()
	// if useAuth == false {
	// 	ctx = nil
	// }

	envID := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	extensionInput := &model.ExtensionInput{
		Name:          "Extension Test",
		Key:           "extensiontest",
		Type:          "once",
		EnvironmentID: envID,
		Config:        model.JSON{json.RawMessage("[{\"key\": \"value\"}]")},
	}
	extensionResolver, err := ts.Resolver.CreateExtension(&struct{ Extension *model.ExtensionInput }{Extension: extensionInput})
	ts.cleanupExtensionIDs = append(ts.cleanupExtensionIDs, extensionResolver.DBExtensionResolver.Model.ID)
	log.Warn(ts.cleanupExtensionIDs)

	return extensionResolver, err
}

func (ts *ProjectExtensionTestSuite) Test1CreateProjectExtensionSuccess() {
	envResolver, err := ts.createTestEnvironment(true)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.createTestProject(true, envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.createTestExtension(true, envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// projectID := projectResolver.DBProjectResolver.Project.Model.ID.String()
	// envID := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	// extensionID := extensionResolver.DBExtensionResolver.Extension.Model.ID.String()

	// projectExtensionInput := model.ProjectExtensionInput{
	// 	ProjectID:   projectID,
	// 	ExtensionID: extensionID,
	// 	Config: model.JSON{
	// 		RawMessage: json.RawMessage(""),
	// 	},
	// 	CustomConfig: model.JSON{
	// 		RawMessage: json.RawMessage(""),
	// 	},
	// 	EnvironmentID: envID,
	// }

	// spew.Dump(projectExtensionInput)
	// projectExtension, err := ts.Resolver.CreateProjectExtension(test.ResolverAuthContext(), &struct{ ProjectExtension *model.ProjectExtensionInput }{ProjectExtension: &projectExtensionInput})
	// if err != nil {
	// 	assert.FailNow(ts.T(), err.Error())
	// }

	// ts.cleanupProjectExtensionIDs = append(ts.cleanupProjectExtensionIDs, projectExtension.DBProjectExtensionResolver.ProjectExtension.ID)

	ts.TearDownTest()
}

func (ts *ProjectExtensionTestSuite) Test2CreateProjectExtensionFailure() {
	envResolver, err := ts.createTestEnvironment(true)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.createTestProject(true, envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.createTestProject(true, envResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *ProjectExtensionTestSuite) Test3CreateProjectNoAuth() {
	envResolver, err := ts.createTestEnvironment(true)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.createTestProject(false, envResolver)
	assert.NotNil(ts.T(), err)
}

// func (ts *ProjectExtensionTestSuite) Test2GormCreateProjectExtension() {

// }

func (ts *ProjectExtensionTestSuite) TearDownTest() {
	for _, id := range ts.cleanupProjectExtensionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.ProjectExtension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupProjectExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupProjectIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Project{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupProjectIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupExtensionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Extension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupEnvironmentIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupEnvironmentIDs = make([]uuid.UUID, 0)
}

func TestSuiteProjectExtensionResolver(t *testing.T) {
	ts := new(ProjectExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
