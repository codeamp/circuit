package graphql_resolver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ReleaseExtensionTestSuite struct {
	suite.Suite
	Resolver                 *graphql_resolver.Resolver
	ReleaseExtensionResolver *graphql_resolver.ReleaseExtensionResolver

	cleanupExtensionIDs        []uuid.UUID
	cleanupEnvironmentIDs      []uuid.UUID
	cleanupProjectIDs          []uuid.UUID
	cleanupSecretIDs           []uuid.UUID
	cleanupProjectBookmarkIDs  []uuid.UUID
	cleanupProjectExtensionIDs []uuid.UUID
	cleanupFeatureIDs          []uuid.UUID
	cleanupServiceIDs          []uuid.UUID
	cleanupServiceSpecIDs      []uuid.UUID
	cleanupReleaseIDs          []uuid.UUID
	cleanupReleaseExtensionIDs []uuid.UUID
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
	suite.ReleaseExtensionResolver = &graphql_resolver.ReleaseExtensionResolver{DBReleaseExtensionResolver: &db_resolver.ReleaseExtensionResolver{DB: db}}
}

func (ts *ReleaseExtensionTestSuite) TestReleaseExtensionInterface() {
	// Environment
	envInput := model.EnvironmentInput{
		Name:      "TestReleaseExtensionInterface",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := ts.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

	// Project
	envId := fmt.Sprintf("%v", envResolver.DBEnvironmentResolver.Environment.Model.ID)
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: &envId,
	}

	createProjectResolver, err := ts.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// TODO: ADB This should be happening in the CreateProject function!
	// If an ID for an Environment is supplied, Project should try to look that up and return resolver
	// that includes project AND environment
	createProjectResolver.DBProjectResolver.Environment = envResolver.DBEnvironmentResolver.Environment
	ts.cleanupProjectIDs = append(ts.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.Model.ID)

	projectID := string(createProjectResolver.ID())

	// Extension
	extensionInput := model.ExtensionInput{
		Name:          "TestReleaseExtensionInterface",
		Key:           "test-project-interface",
		Component:     "",
		EnvironmentID: envId,
		Config:        model.JSON{[]byte("[]")},
		Type:          "workflow",
	}
	extensionResolver, err := ts.Resolver.CreateExtension(&struct {
		Extension *model.ExtensionInput
	}{Extension: &extensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupExtensionIDs = append(ts.cleanupExtensionIDs, extensionResolver.DBExtensionResolver.Extension.Model.ID)

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
	assert.Nil(ts.T(), err)

	extCustomConfigMap := make(map[string]ExtConfig)
	extCustomConfigJSON, err := json.Marshal(extCustomConfigMap)
	assert.Nil(ts.T(), err)

	extensionID := string(extensionResolver.ID())
	projExtensionInput := model.ProjectExtensionInput{
		ProjectID:     projectID,
		ExtensionID:   extensionID,
		Config:        model.JSON{extConfigJSON},
		CustomConfig:  model.JSON{extCustomConfigJSON},
		EnvironmentID: envId,
	}
	projectExtensionResolver, err := ts.Resolver.CreateProjectExtension(test.ResolverAuthContext(), &struct {
		ProjectExtension *model.ProjectExtensionInput
	}{ProjectExtension: &projExtensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupProjectExtensionIDs = append(ts.cleanupProjectExtensionIDs, projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.Model.ID)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Features
	projectIDUUID, err := uuid.FromString(strings.ToUpper(projectID))
	assert.Nil(ts.T(), err)

	feature := model.Feature{
		ProjectID:  projectIDUUID,
		Message:    "A test feature message",
		User:       "TestReleaseExtensionInterface",
		Hash:       "42941a0900e952f7f78994d53b699aea23926804",
		ParentHash: "",
		Ref:        "refs/heads/master",
		Created:    time.Now(),
	}

	db := ts.Resolver.DB.Create(&feature)
	if db.Error != nil {
		assert.FailNow(ts.T(), db.Error.Error())
	}
	ts.cleanupFeatureIDs = append(ts.cleanupFeatureIDs, feature.Model.ID)

	// Releases
	featureID := feature.Model.ID.String()
	releaseInput := model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envId,
		ForceRebuild:  false,
	}

	releaseResolver, err := ts.Resolver.CreateRelease(test.ResolverAuthContext(), &struct{ Release *model.ReleaseInput }{Release: &releaseInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	releaseID := releaseResolver.DBReleaseResolver.Model.ID
	ts.cleanupReleaseIDs = append(ts.cleanupReleaseIDs, releaseID)

	projectExtensionID := projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.Model.ID
	releaseExtension := model.ReleaseExtension{
		State:              transistor.GetState("complete"),
		StateMessage:       "TestReleaseExtensionInterface",
		ReleaseID:          releaseID,
		FeatureHash:        "42941a0900e952f7f78994d53b699aea23926804",
		ServicesSignature:  "ServicesSignature",
		SecretsSignature:   "SecretsSignature",
		ProjectExtensionID: projectExtensionID,
		Type:               "workflow",
	}

	res := ts.Resolver.DB.Create(&releaseExtension)
	if res.Error != nil {
		assert.FailNow(ts.T(), res.Error.Error())
	}
	ts.cleanupReleaseExtensionIDs = append(ts.cleanupReleaseExtensionIDs, releaseID)

	// Test Release Extension Interface
	releaseExtensionResolvers := releaseResolver.ReleaseExtensions()
	assert.NotNil(ts.T(), releaseExtensionResolvers)
	assert.NotEmpty(ts.T(), releaseExtensionResolvers)

	releaseExtensionResolver := releaseExtensionResolvers[0]
	_ = releaseExtensionResolver.ID()
	_, err = releaseExtensionResolver.Release()

	assert.Equal(ts.T(), releaseExtension.ServicesSignature, releaseExtensionResolver.ServicesSignature())
	assert.Equal(ts.T(), releaseExtension.SecretsSignature, releaseExtensionResolver.SecretsSignature())
	assert.Equal(ts.T(), "complete", releaseExtensionResolver.State())
	assert.Equal(ts.T(), string(releaseExtension.Type), releaseExtensionResolver.Type())
	assert.Equal(ts.T(), releaseExtension.StateMessage, releaseExtensionResolver.StateMessage())

	_ = releaseExtensionResolver.Artifacts()
	_ = releaseExtensionResolver.Finished()

	created_at_diff := time.Now().Sub(envResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	data, err := releaseExtensionResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = releaseExtensionResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)

	// Test Query Interface
	var ctx context.Context
	_, err = ts.Resolver.ReleaseExtensions(ctx)
	assert.NotNil(ts.T(), err)

	releaseExtensions, err := ts.Resolver.ReleaseExtensions(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), releaseExtensions)
	assert.NotEmpty(ts.T(), releaseExtensions)
}

func (ts *ReleaseExtensionTestSuite) TearDownTest() {
	for _, id := range ts.cleanupReleaseExtensionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.ReleaseExtension{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupReleaseExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupFeatureIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Feature{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupFeatureIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupServiceIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Service{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupServiceIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupServiceSpecIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.ServiceSpec{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupServiceSpecIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupExtensionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Extension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupProjectExtensionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.ProjectExtension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupProjectExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupProjectBookmarkIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.ProjectBookmark{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupProjectBookmarkIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupProjectIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Project{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupProjectIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupEnvironmentIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupEnvironmentIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupSecretIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Secret{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupSecretIDs = make([]uuid.UUID, 0)
}

func TestSuiteReleaseExtensionResolver(t *testing.T) {
	suite.Run(t, new(ReleaseExtensionTestSuite))
}
