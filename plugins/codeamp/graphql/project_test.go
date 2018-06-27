package graphql_resolver_test

import (
	"context"
	"encoding/json"
	"fmt"

	"testing"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	graphql "github.com/graph-gophers/graphql-go"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	cleanupExtensionIDs        []uuid.UUID
	cleanupEnvironmentIDs      []uuid.UUID
	cleanupProjectIDs          []uuid.UUID
	cleanupSecretIDs           []uuid.UUID
	cleanupProjectBookmarkIDs  []uuid.UUID
	cleanupProjectExtensionIDs []uuid.UUID
}

func (suite *ProjectTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectBookmark{},
		&model.ProjectEnvironment{},
		&model.ProjectExtension{},
		&model.ProjectSettings{},
		&model.UserPermission{},
		&model.Environment{},
		&model.Extension{},
		&model.Secret{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
}

func (suite *ProjectTestSuite) TestProjectInterface() {
	// Environment
	envInput := model.EnvironmentInput{
		Name:      "TestProjectInterface",
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

	// TODO: ADB This should be happening in the CreateProject function!
	// If an ID for an Environment is supplied, Project should try to look that up and return resolver
	// that includes project AND environment
	createProjectResolver.DBProjectResolver.Environment = envResolver.DBEnvironmentResolver.Environment
	suite.cleanupProjectIDs = append(suite.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.Model.ID)

	// Secret
	projectID := string(createProjectResolver.ID())
	secretInput := model.SecretInput{
		Key:           "TestProjectInterface",
		Type:          "env",
		Scope:         "extension",
		EnvironmentID: string(envResolver.ID()),
		ProjectID:     &projectID,
		IsSecret:      false,
	}

	secretResolver, err := suite.Resolver.CreateSecret(test.ResolverAuthContext(), &struct {
		Secret *model.SecretInput
	}{Secret: &secretInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	suite.cleanupSecretIDs = append(suite.cleanupSecretIDs, secretResolver.DBSecretResolver.Secret.Model.ID)

	// Extension
	extensionInput := model.ExtensionInput{
		Name:          "TestProjectInterface",
		Key:           "test-project-interface",
		Component:     "",
		EnvironmentID: envId,
		Config:        model.JSON{[]byte("[]")},
		Type:          "workflow",
	}
	extensionResolver, err := suite.Resolver.CreateExtension(&struct {
		Extension *model.ExtensionInput
	}{Extension: &extensionInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// Move this to model namespace!
	type ExtConfig struct {
		Key           string `json:"key"`
		Value         string `json:"value"`
		Secret        bool   `json:"secret"`
		AllowOverride bool   `json:"allowOverride"`
	}

	// Project Extension
	extConfigMap := make([]ExtConfig, 0)
	extConfigJSON, err := json.Marshal(extConfigMap)
	assert.Nil(suite.T(), err)

	extCustomConfigMap := make(map[string]ExtConfig)
	extCustomConfigJSON, err := json.Marshal(extCustomConfigMap)
	assert.Nil(suite.T(), err)

	log.Warn("Creating project extension")
	extensionID := string(extensionResolver.ID())
	projExtensionInput := model.ProjectExtensionInput{
		ProjectID:     projectID,
		ExtensionID:   extensionID,
		Config:        model.JSON{extConfigJSON},
		CustomConfig:  model.JSON{extCustomConfigJSON},
		EnvironmentID: envId,
	}
	_, err = suite.Resolver.CreateProjectExtension(test.ResolverAuthContext(), &struct {
		ProjectExtension *model.ProjectExtensionInput
	}{ProjectExtension: &projExtensionInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	log.Warn("Beginning test of interface")
	// Test
	_ = createProjectResolver.ID()
	_ = createProjectResolver.Name()
	_ = createProjectResolver.Repository()
	_ = createProjectResolver.Secret()

	gitUrl := createProjectResolver.GitUrl()
	assert.Equal(suite.T(), projectInput.GitUrl, gitUrl)

	gitProtocol := createProjectResolver.GitProtocol()
	assert.Equal(suite.T(), projectInput.GitProtocol, gitProtocol)

	_ = createProjectResolver.RsaPrivateKey()
	_ = createProjectResolver.RsaPublicKey()

	showDeployed := false
	_ = createProjectResolver.Features(&struct{ ShowDeployed *bool }{ShowDeployed: &showDeployed})
	_, _ = createProjectResolver.CurrentRelease()
	_ = createProjectResolver.Releases()
	_ = createProjectResolver.Services()

	var ctx context.Context
	_, err = createProjectResolver.Secrets(ctx)
	assert.NotNil(suite.T(), err)

	secretsList, err := createProjectResolver.Secrets(test.ResolverAuthContext())
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), secretsList)

	_, err = createProjectResolver.Extensions()
	assert.Nil(suite.T(), err)

	_ = createProjectResolver.GitBranch()
	_ = createProjectResolver.ContinuousDeploy()
	projectEnvironments := createProjectResolver.Environments()
	assert.NotEmpty(suite.T(), projectEnvironments)

	_ = createProjectResolver.Bookmarked(ctx)
	_ = createProjectResolver.Bookmarked(test.ResolverAuthContext())

	_ = createProjectResolver.Created()

	data, err := createProjectResolver.MarshalJSON()
	assert.Nil(suite.T(), err)

	err = createProjectResolver.UnmarshalJSON(data)
	assert.Nil(suite.T(), err)
}

func (suite *ProjectTestSuite) TestCreateProject() {
	// setup
	env := model.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	envId := fmt.Sprintf("%v", env.Model.ID)
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

	// assert permissions exist for dev env
	//assert.Equal(suite.T(), createProjectResolver.Permissions(), []string{env.Model.ID.String()})
}

func (suite *ProjectTestSuite) TestQueryProject() {
	// setup
	env := model.Environment{
		Name:      "TestQueryProject",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	envId := fmt.Sprintf("%v", env.Model.ID)
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

	var ctx context.Context
	_, err = suite.Resolver.Projects(ctx, &struct {
		ProjectSearch *model.ProjectSearchInput
	}{
		ProjectSearch: nil,
	})
	assert.NotNil(suite.T(), err)

	// do a search for 'foo'
	searchQuery := "foo"
	projects, err := suite.Resolver.Projects(test.ResolverAuthContext(), &struct {
		ProjectSearch *model.ProjectSearchInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: false,
			Repository: &searchQuery,
		},
	})
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), projects)

	projectId := createProjectResolver.ID()
	_, err = suite.Resolver.Project(ctx, &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectId,
		EnvironmentID: &envId,
	})
	assert.NotNil(suite.T(), err)

	_, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectId,
		EnvironmentID: &envId,
	})
	assert.Nil(suite.T(), err)

}

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestUpdateProjectEnvironments() {
	// setup
	project := model.Project{
		Name:          "foo",
		Slug:          "foo",
		Repository:    "foo/foo",
		Secret:        "foo",
		GitUrl:        "foo",
		GitProtocol:   "foo",
		RsaPrivateKey: "foo",
		RsaPublicKey:  "foo",
	}
	suite.Resolver.DB.Create(&project)
	suite.cleanupProjectIDs = append(suite.cleanupProjectIDs, project.Model.ID)

	env := model.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)
	suite.cleanupEnvironmentIDs = append(suite.cleanupProjectIDs, env.Model.ID)

	projectEnvironmentsInput := model.ProjectEnvironmentsInput{
		ProjectID: project.Model.ID.String(),
		Permissions: []model.ProjectEnvironmentInput{
			{
				EnvironmentID: env.Model.ID.String(),
				Grant:         true,
			},
		},
	}

	updateProjectEnvironmentsResp, err := suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// check if env is found in response
	assert.Equal(suite.T(), 1, len(updateProjectEnvironmentsResp))
	assert.Equal(suite.T(), env.Model.ID, updateProjectEnvironmentsResp[0].DBEnvironmentResolver.Environment.Model.ID)

	projectEnvironments := []model.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 1, len(projectEnvironments))
	assert.Equal(suite.T(), env.Model.ID.String(), projectEnvironments[0].EnvironmentID.String())

	// take away access
	projectEnvironmentsInput.Permissions[0].Grant = false
	updateProjectEnvironmentsResp, err = suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 0, len(updateProjectEnvironmentsResp))

	projectEnvironments = []model.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 0, len(projectEnvironments))
}

func (suite *ProjectTestSuite) TestGetBookmarkedAndQueryProjects() {
	// init 3 projects into db
	projectNames := []string{"foo", "foobar", "boo"}
	userId := uuid.NewV1()

	for _, name := range projectNames {
		project := model.Project{
			Name:          name,
			Slug:          name,
			Repository:    fmt.Sprintf("test/%s", name),
			Secret:        "foo",
			GitUrl:        "foo",
			GitProtocol:   "foo",
			RsaPrivateKey: "foo",
			RsaPublicKey:  "foo",
		}

		suite.Resolver.DB.Create(&project)
		suite.cleanupProjectIDs = append(suite.cleanupProjectIDs, project.Model.ID)

		projectBookmark := model.ProjectBookmark{
			UserID:    userId,
			ProjectID: project.Model.ID,
		}

		suite.Resolver.DB.Create(&projectBookmark)
		suite.cleanupProjectBookmarkIDs = append(suite.cleanupProjectBookmarkIDs, projectBookmark.Model.ID)
	}

	adminContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      userId.String(),
		Email:       "codeamp",
		Permissions: []string{"admin"},
	})
	projects, err := suite.Resolver.Projects(adminContext, &struct {
		ProjectSearch *model.ProjectSearchInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: true,
		},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 3, len(projects))

	// do a search for 'foo'
	searchQuery := "foo"
	projects, err = suite.Resolver.Projects(adminContext, &struct {
		ProjectSearch *model.ProjectSearchInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: false,
			Repository: &searchQuery,
		},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 2, len(projects))
}

func (suite *ProjectTestSuite) TearDownTest() {
	for _, id := range suite.cleanupExtensionIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Extension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	suite.cleanupExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupProjectExtensionIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.ProjectExtension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	suite.cleanupProjectExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupProjectBookmarkIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.ProjectBookmark{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
	suite.cleanupProjectBookmarkIDs = make([]uuid.UUID, 0)

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

	for _, id := range suite.cleanupSecretIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Secret{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
	suite.cleanupSecretIDs = make([]uuid.UUID, 0)
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
