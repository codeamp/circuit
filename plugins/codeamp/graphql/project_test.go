package graphql_resolver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	cleanupFeatureIDs          []uuid.UUID
	cleanupServiceIDs          []uuid.UUID
	cleanupServiceSpecIDs      []uuid.UUID
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

	extensionID := string(extensionResolver.ID())
	projExtensionInput := model.ProjectExtensionInput{
		ProjectID:     projectID,
		ExtensionID:   extensionID,
		Config:        model.JSON{extConfigJSON},
		CustomConfig:  model.JSON{extCustomConfigJSON},
		EnvironmentID: envId,
	}
	projectExtensionResolver, err := suite.Resolver.CreateProjectExtension(test.ResolverAuthContext(), &struct {
		ProjectExtension *model.ProjectExtensionInput
	}{ProjectExtension: &projExtensionInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupProjectExtensionIDs = append(suite.cleanupProjectExtensionIDs, projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.Model.ID)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	suite.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Features
	projectIDUUID, err := uuid.FromString(strings.ToUpper(projectID))
	assert.Nil(suite.T(), err)

	feature := model.Feature{
		ProjectID:  projectIDUUID,
		Message:    "A test feature message",
		User:       "TestProjectInterface",
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

	// Releases
	featureID := feature.Model.ID.String()
	releaseInput := model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envId,
		ForceRebuild:  false,
	}
	_, err = suite.Resolver.CreateRelease(test.ResolverAuthContext(), &struct{ Release *model.ReleaseInput }{Release: &releaseInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecInput := model.ServiceSpecInput{
		Name:                   "test",
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	serviceSpecResolver, err := suite.Resolver.CreateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupServiceSpecIDs = append(suite.cleanupServiceSpecIDs, serviceSpecResolver.DBServiceSpecResolver.ServiceSpec.Model.ID)

	// Services
	servicePortInputs := []model.ServicePortInput{}
	serviceInput := model.ServiceInput{
		ProjectID:     projectID,
		Command:       "echo \"hello\" && exit 0",
		Name:          "test-service",
		ServiceSpecID: string(serviceSpecResolver.ID()),
		Count:         "0",
		Ports:         &servicePortInputs,
		Type:          "general",
		EnvironmentID: envId,
	}

	serviceResolver, err := suite.Resolver.CreateService(&struct{ Service *model.ServiceInput }{Service: &serviceInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupServiceIDs = append(suite.cleanupServiceIDs, serviceResolver.DBServiceResolver.Service.Model.ID)

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
	featuresList := createProjectResolver.Features(&struct{ ShowDeployed *bool }{ShowDeployed: &showDeployed})
	assert.NotEmpty(suite.T(), featuresList, "Features List Empty")

	_, _ = createProjectResolver.CurrentRelease()
	releasesList := createProjectResolver.Releases()
	assert.NotEmpty(suite.T(), releasesList, "Releases List Empty")

	servicesList := createProjectResolver.Services()
	assert.NotEmpty(suite.T(), servicesList, "Services List Empty")

	var ctx context.Context
	_, err = createProjectResolver.Secrets(ctx)
	assert.NotNil(suite.T(), err)

	secretsList, err := createProjectResolver.Secrets(test.ResolverAuthContext())
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), secretsList, "Secrets List Empty")

	extensionsList, err := createProjectResolver.Extensions()
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), extensionsList, "Extensions List Empty")

	_ = createProjectResolver.GitBranch()
	_ = createProjectResolver.ContinuousDeploy()
	projectEnvironments := createProjectResolver.Environments()
	assert.NotEmpty(suite.T(), projectEnvironments, "Project Environments Empty")

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
	for _, id := range suite.cleanupFeatureIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Feature{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	suite.cleanupFeatureIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupServiceIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Service{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	suite.cleanupServiceIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupServiceSpecIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.ServiceSpec{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	suite.cleanupServiceSpecIDs = make([]uuid.UUID, 0)

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
