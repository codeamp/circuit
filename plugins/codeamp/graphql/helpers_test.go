package graphql_resolver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	. "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	_ "github.com/davecgh/go-spew/spew"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type Helper struct {
	Resolver *Resolver
	name     string
	context  context.Context

	cleanupExtensionIDs                 []uuid.UUID
	cleanupEnvironmentIDs               []uuid.UUID
	cleanupProjectIDs                   []uuid.UUID
	cleanupSecretIDs                    []uuid.UUID
	cleanupProjectBookmarkIDs           []uuid.UUID
	cleanupProjectExtensionIDs          []uuid.UUID
	cleanupFeatureIDs                   []uuid.UUID
	cleanupServiceIDs                   []uuid.UUID
	cleanupServiceDeploymentStrategyIDs []uuid.UUID
	cleanupServiceSpecIDs               []uuid.UUID
	cleanupReleaseIDs                   []uuid.UUID
	cleanupReleaseExtensionIDs          []uuid.UUID
}

func (helper *Helper) SetResolver(resolver *Resolver, name string) {
	helper.Resolver = resolver
	helper.name = name
}

func (helper *Helper) SetContext(context context.Context) {
	helper.context = context
}

func (helper *Helper) CreateEnvironment(t *testing.T) *EnvironmentResolver {
	// Environment
	return helper.CreateEnvironmentWithName(t, helper.name)
}

func (helper *Helper) CreateEnvironmentWithName(t *testing.T, name string) *EnvironmentResolver {
	// Environment
	resolver, err := helper.CreateEnvironmentWithError(name)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	return resolver
}

func (helper *Helper) CreateEnvironmentWithError(name string) (*EnvironmentResolver, error) {
	envInput := model.EnvironmentInput{
		Name:      name,
		Key:       name,
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := helper.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{&envInput})
	if err == nil {
		helper.cleanupEnvironmentIDs = append(helper.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)
	}
	return envResolver, err
}

func (helper *Helper) CreateProject(t *testing.T, envResolver *EnvironmentResolver) (*ProjectResolver, error) {
	projectResolver, err := helper.CreateProjectWithRepo(t, envResolver, "https://github.com/foo/goo.git")
	if err == nil {
		projectResolver.DBProjectResolver.Environment = envResolver.DBEnvironmentResolver.Environment
	}
	return projectResolver, err
}

func (helper *Helper) CreateProjectWithInput(t *testing.T,
	projectInput *model.ProjectInput) (*ProjectResolver, error) {

	projectResolver, err := helper.Resolver.CreateProject(helper.context, &struct {
		Project *model.ProjectInput
	}{Project: projectInput})
	if err == nil {
		helper.cleanupProjectIDs = append(helper.cleanupProjectIDs, projectResolver.DBProjectResolver.Project.Model.ID)
	}
	return projectResolver, err
}

func (helper *Helper) CreateProjectWithRepo(t *testing.T, envResolver *EnvironmentResolver, gitUrl string) (*ProjectResolver, error) {
	envId := string(envResolver.ID())
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        gitUrl,
		EnvironmentID: &envId,
	}

	return helper.CreateProjectWithInput(t, &projectInput)
}

func (helper *Helper) CreateSecret(t *testing.T,
	projectResolver *ProjectResolver) *SecretResolver {

	envID := string(projectResolver.Environments()[0].ID())

	// Secret
	projectID := string(projectResolver.ID())
	secretInput := model.SecretInput{
		Key:           helper.name,
		Type:          "env",
		Scope:         "extension",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}

	secretResolver, err := helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	return secretResolver
}

func (helper *Helper) CreateSecretWithInput(secretInput *model.SecretInput) (*SecretResolver, error) {
	secretResolver, err := helper.Resolver.CreateSecret(helper.context, &struct {
		Secret *model.SecretInput
	}{Secret: secretInput})

	if err == nil {
		helper.cleanupSecretIDs = append(helper.cleanupSecretIDs, secretResolver.DBSecretResolver.Secret.Model.ID)
	}
	return secretResolver, err
}

func (helper *Helper) CreateExtensionOfType(t *testing.T, envResolver *EnvironmentResolver, extensionKey string, extensionType string, config *[]model.ExtConfig) *ExtensionResolver {
	envId := fmt.Sprintf("%v", envResolver.DBEnvironmentResolver.Environment.Model.ID)

	configData, err := json.Marshal(&config)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Extension
	extensionInput := model.ExtensionInput{
		Name:          helper.name,
		Key:           extensionKey,
		Component:     "test-component",
		EnvironmentID: envId,
		Config:        model.JSON{configData},
		Type:          extensionType,
		Cacheable:     false,
	}

	return helper.CreateExtensionWithInput(t, &extensionInput)
}

func (helper *Helper) CreateExtension(t *testing.T, envResolver *EnvironmentResolver) *ExtensionResolver {
	envId := fmt.Sprintf("%v", envResolver.DBEnvironmentResolver.Environment.Model.ID)

	config := []model.ExtConfig{
		{
			Key:           "test-key",
			Value:         "test-value",
			AllowOverride: true,
		},
		{
			Key:           "test-key",
			Value:         "test-value",
			AllowOverride: true,
			Secret:        true,
		},
	}
	configData, err := json.Marshal(&config)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Extension
	extensionInput := model.ExtensionInput{
		Name:          helper.name,
		Key:           "test-project-interface",
		Component:     "test-component",
		EnvironmentID: envId,
		Config:        model.JSON{configData},
		Type:          "workflow",
		Cacheable:     false,
	}

	return helper.CreateExtensionWithInput(t, &extensionInput)
}

func (helper *Helper) CreateExtensionWithInput(t *testing.T, extensionInput *model.ExtensionInput) *ExtensionResolver {
	extensionResolver, err := helper.Resolver.CreateExtension(&struct {
		Extension *model.ExtensionInput
	}{Extension: extensionInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupExtensionIDs = append(helper.cleanupExtensionIDs, extensionResolver.DBExtensionResolver.Extension.Model.ID)
	return extensionResolver
}

func (helper *Helper) CreateProjectExtension(t *testing.T,
	extensionResolver *ExtensionResolver,
	projectResolver *ProjectResolver) *ProjectExtensionResolver {

	resolver, err := helper.CreateProjectExtensionWithError(t, extensionResolver, projectResolver)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return resolver
}

func (helper *Helper) CreateProjectExtensionWithConfig(t *testing.T,
	extensionResolver *ExtensionResolver,
	projectResolver *ProjectResolver,
	extConfig *[]model.ExtConfig,
	extCustomConfigMap *map[string]interface{}) (*ProjectExtensionResolver, error) {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	// Project Extension
	extConfigJSON, err := json.Marshal(extConfig)
	assert.Nil(t, err)

	extCustomConfigJSON, err := json.Marshal(extCustomConfigMap)
	assert.Nil(t, err)

	extensionID := string(extensionResolver.ID())
	projectExtensionInput := model.ProjectExtensionInput{
		ProjectID:     projectID,
		ExtensionID:   extensionID,
		Config:        model.JSON{extConfigJSON},
		CustomConfig:  model.JSON{extCustomConfigJSON},
		EnvironmentID: envID,
	}

	projectExtensionResolver, err := helper.Resolver.CreateProjectExtension(helper.context, &struct {
		ProjectExtension *model.ProjectExtensionInput
	}{ProjectExtension: &projectExtensionInput})
	if err == nil {
		helper.cleanupProjectExtensionIDs = append(helper.cleanupProjectExtensionIDs, projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.Model.ID)
	}
	return projectExtensionResolver, err
}

func (helper *Helper) CreateProjectExtensionWithError(t *testing.T,
	extensionResolver *ExtensionResolver,
	projectResolver *ProjectResolver) (*ProjectExtensionResolver, error) {

	// Project Extension
	extConfig := make([]model.ExtConfig, 0)
	extConfig = append(extConfig,
		model.ExtConfig{
			Key:           "test-key",
			Value:         "test-value",
			AllowOverride: true,
			Secret:        false,
		})

	extCustomConfigMap := make(map[string]interface{})
	extCustomConfigMap["test"] = model.ExtConfig{
		Key:           "test-key",
		Value:         "test-value",
		AllowOverride: true,
		Secret:        false,
	}

	return helper.CreateProjectExtensionWithConfig(t, extensionResolver, projectResolver, &extConfig, &extCustomConfigMap)
}

func (helper *Helper) CreateFeature(t *testing.T, projectResolver *ProjectResolver) *FeatureResolver {
	projectID := string(projectResolver.ID())

	// Features
	projectIDUUID, err := uuid.FromString(strings.ToUpper(projectID))
	assert.Nil(t, err)

	feature := model.Feature{
		ProjectID:  projectIDUUID,
		Message:    "A test feature message",
		User:       helper.name,
		Hash:       "42941a0900e952f7f78994d53b699aea23926804",
		ParentHash: "",
		Ref:        "refs/heads/master",
		Created:    time.Now(),
	}

	db := helper.Resolver.DB.Create(&feature)
	if db.Error != nil {
		assert.FailNow(t, db.Error.Error())
	}

	helper.cleanupFeatureIDs = append(helper.cleanupFeatureIDs, feature.Model.ID)
	return &FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: helper.Resolver.DB, Feature: feature}}
}

func (helper *Helper) CreateFeatureWithParent(t *testing.T, projectResolver *ProjectResolver) *FeatureResolver {
	if projectResolver == nil {
		assert.FailNow(t, "Project Resolver is NULL")
	}
	projectID := string(projectResolver.ID())

	// Features
	projectIDUUID, err := uuid.FromString(strings.ToUpper(projectID))
	assert.Nil(t, err)

	feature := model.Feature{
		ProjectID:  projectIDUUID,
		Message:    "A test feature message",
		User:       helper.name,
		Hash:       "42941a0900e952f7f78994d53b699aea23926804",
		ParentHash: "7f78994d53b699aea239268950441a090952f0e9",
		Ref:        "refs/heads/master",
		Created:    time.Now(),
	}

	db := helper.Resolver.DB.Create(&feature)
	if db.Error != nil {
		assert.FailNow(t, db.Error.Error())
	}

	helper.cleanupFeatureIDs = append(helper.cleanupFeatureIDs, feature.Model.ID)
	return &FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: helper.Resolver.DB, Feature: feature}}
}

func (helper *Helper) CreateRelease(t *testing.T,
	featureResolver *FeatureResolver,
	projectResolver *ProjectResolver) *ReleaseResolver {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	// Releases
	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	return helper.CreateReleaseWithInput(t, projectResolver, releaseInput)
}

func (helper *Helper) CreateReleaseWithInput(t *testing.T,
	projectResolver *ProjectResolver,
	releaseInput *model.ReleaseInput) *ReleaseResolver {

	// Release
	releaseResolver, err := helper.CreateReleaseWithError(t, projectResolver, releaseInput)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return releaseResolver
}

func (helper *Helper) CreateReleaseWithError(t *testing.T,
	projectResolver *ProjectResolver,
	releaseInput *model.ReleaseInput) (*ReleaseResolver, error) {

	// Release
	releaseResolver, err := helper.Resolver.CreateRelease(helper.context, &struct{ Release *model.ReleaseInput }{releaseInput})
	if err == nil {
		helper.cleanupReleaseIDs = append(helper.cleanupReleaseIDs, releaseResolver.DBReleaseResolver.Release.Model.ID)
	}
	return releaseResolver, err
}

func (helper *Helper) CreateReleaseExtension(t *testing.T,
	releaseResolver *ReleaseResolver,
	projectExtensionResolver *ProjectExtensionResolver) *ReleaseExtensionResolver {

	releaseID, err := uuid.FromString(string(releaseResolver.ID()))
	assert.Nil(t, err)

	projectExtensionID, err := uuid.FromString(string(projectExtensionResolver.ID()))
	assert.Nil(t, err)

	releaseExtension := model.ReleaseExtension{
		State:              transistor.GetState("waiting"),
		StateMessage:       helper.name,
		ReleaseID:          releaseID,
		FeatureHash:        "42941a0900e952f7f78994d53b699aea23926804",
		ServicesSignature:  "ServicesSignature",
		SecretsSignature:   "SecretsSignature",
		ProjectExtensionID: projectExtensionID,
		Type:               "workflow",
	}

	res := helper.Resolver.DB.Create(&releaseExtension)
	if res.Error != nil {
		assert.FailNow(t, res.Error.Error())
	}

	helper.cleanupReleaseExtensionIDs = append(helper.cleanupReleaseExtensionIDs, releaseID)

	return &ReleaseExtensionResolver{DBReleaseExtensionResolver: &db_resolver.ReleaseExtensionResolver{ReleaseExtension: releaseExtension, DB: helper.Resolver.DB}}
}

func (helper *Helper) CreateServiceSpec(t *testing.T, isDefault bool) *ServiceSpecResolver {
	// Service Spec ID
	serviceSpecInput := model.ServiceSpecInput{
		Name:                   helper.name,
		CpuRequest:             "100",
		CpuLimit:               "200",
		MemoryRequest:          "300",
		MemoryLimit:            "400",
		TerminationGracePeriod: "500",
		IsDefault: isDefault,
	}
	serviceSpecResolver, err := helper.Resolver.CreateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupServiceSpecIDs = append(helper.cleanupServiceSpecIDs, serviceSpecResolver.DBServiceSpecResolver.ServiceSpec.Model.ID)
	return serviceSpecResolver
}


func (helper *Helper) CreateService(t *testing.T,
	projectResolver *ProjectResolver,
	deploymentStrategy *model.DeploymentStrategyInput,
	readinessProbe *model.ServiceHealthProbeInput,
	livenessProbe *model.ServiceHealthProbeInput,
	preStopHook *string) *ServiceResolver {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	servicePortInputs := []model.ServicePortInput{
		{
			Port:     80,
			Protocol: "HTTP",
		},
	}

	// Services
	serviceInput := model.ServiceInput{
		ProjectID:          projectID,
		Command:            "echo \"hello\" && exit 0",
		Name:               helper.name,
		Count:              1,
		Ports:              &servicePortInputs,
		Type:               "general",
		EnvironmentID:      envID,
		DeploymentStrategy: deploymentStrategy,
		ReadinessProbe:     readinessProbe,
		LivenessProbe:      livenessProbe,
		PreStopHook:        preStopHook,
	}

	serviceResolver, err := helper.Resolver.CreateService(&struct{ Service *model.ServiceInput }{Service: &serviceInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupServiceIDs = append(helper.cleanupServiceIDs, serviceResolver.DBServiceResolver.Service.Model.ID)
	return serviceResolver
}

func (helper *Helper) CreateUser(t *testing.T) *UserResolver {
	user := model.User{
		Email:    "test@example.com",
		Password: "example",
	}

	res := helper.Resolver.DB.Create(&user)
	if res.Error != nil {
		assert.FailNow(t, res.Error.Error())
	}

	return &UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: helper.Resolver.DB, User: user}}
}

func (helper *Helper) CreateServiceWithError(t *testing.T,
	projectResolver *ProjectResolver,
	deploymentStrategy *model.DeploymentStrategyInput,
	readinessProbe *model.ServiceHealthProbeInput,
	livenessProbe *model.ServiceHealthProbeInput,
	preStopHook *string) (*ServiceResolver, error) {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	servicePortInputs := []model.ServicePortInput{
		{
			Port:     80,
			Protocol: "HTTP",
		},
	}

	// Services
	serviceInput := model.ServiceInput{
		ProjectID:          projectID,
		Command:            "echo \"hello\" && exit 0",
		Name:               helper.name,
		Count:              1,
		Ports:              &servicePortInputs,
		Type:               "general",
		EnvironmentID:      envID,
		DeploymentStrategy: deploymentStrategy,
		ReadinessProbe:     readinessProbe,
		LivenessProbe:      livenessProbe,
		PreStopHook:        preStopHook,
	}

	serviceResolver, err := helper.Resolver.CreateService(&struct{ Service *model.ServiceInput }{Service: &serviceInput})
	if err != nil {
		return serviceResolver, err
	}

	helper.cleanupServiceIDs = append(helper.cleanupServiceIDs, serviceResolver.DBServiceResolver.Service.Model.ID)
	return serviceResolver, nil
}

func (helper *Helper) TearDownTest(t *testing.T) {
	for _, id := range helper.cleanupFeatureIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Feature{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupFeatureIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupServiceIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Service{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupServiceIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupServiceSpecIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ServiceSpec{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupServiceSpecIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupServiceDeploymentStrategyIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ServiceDeploymentStrategy{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}

	helper.cleanupServiceDeploymentStrategyIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupReleaseExtensionIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ReleaseExtension{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}

	helper.cleanupReleaseExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupExtensionIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Extension{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupReleaseIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Release{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}

		err = helper.Resolver.DB.Unscoped().Delete(&model.ReleaseExtension{}, "release_id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupReleaseIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupProjectExtensionIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ProjectExtension{}, "id = ?", id).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupProjectExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupProjectBookmarkIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ProjectBookmark{}, "id = ?", id).Error
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}
	helper.cleanupProjectBookmarkIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupProjectIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Project{}, "id = ?", id).Error
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		// Delete all associated bookmarks as well
		var bookmarks []model.ProjectBookmark
		if err = helper.Resolver.DB.Unscoped().Where("project_id = ?", id).Find(&bookmarks).Error; err != nil {
			log.Error(err)
			continue
		}

		for _, bookmark := range bookmarks {
			helper.Resolver.DB.Unscoped().Delete(&bookmark)
		}

		// Don't forget to also delete any project_environment associations
		var projectEnvironments []model.ProjectEnvironment
		if err = helper.Resolver.DB.Unscoped().Where("project_id = ?", id).Find(&projectEnvironments).Error; err != nil {
			log.Error(err)
			continue
		}

		for _, projectEnvironment := range projectEnvironments {
			helper.Resolver.DB.Unscoped().Delete(&projectEnvironment)
		}
	}
	helper.cleanupProjectIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupEnvironmentIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Environment{}, "id = ?", id).Error
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}
	helper.cleanupEnvironmentIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupSecretIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Secret{}, "id = ?", id).Error
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}
	helper.cleanupSecretIDs = make([]uuid.UUID, 0)
}
