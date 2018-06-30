package graphql_resolver_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type Helper struct {
	Resolver *graphql_resolver.Resolver
	name     string

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

func (helper *Helper) SetResolver(resolver *graphql_resolver.Resolver, name string) {
	helper.Resolver = resolver
	helper.name = name
}

func (helper *Helper) CreateEnvironment(t *testing.T) *graphql_resolver.EnvironmentResolver {
	// Environment
	return helper.CreateEnvironmentWithName(t, helper.name)
}

func (helper *Helper) CreateEnvironmentWithName(t *testing.T, name string) *graphql_resolver.EnvironmentResolver {
	// Environment
	envInput := model.EnvironmentInput{
		Name:      name,
		Key:       name,
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := helper.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{&envInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupEnvironmentIDs = append(helper.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)
	return envResolver
}

func (helper *Helper) CreateProject(t *testing.T, envResolver *graphql_resolver.EnvironmentResolver) *graphql_resolver.ProjectResolver {
	return helper.CreateProjectWithRepo(t, envResolver, "https://github.com/foo/goo.git")
}

func (helper *Helper) CreateProjectWithInput(t *testing.T,
	envResolver *graphql_resolver.EnvironmentResolver,
	projectInput *model.ProjectInput) *graphql_resolver.ProjectResolver {

	projectResolver, err := helper.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: projectInput})
	if err != nil {
		assert.FailNow(t, err.Error(), projectInput.GitUrl)
	}

	// TODO: ADB This should be happening in the CreateProject function!
	// If an ID for an Environment is supplied, Project should try to look that up and return resolver
	// that includes project AND environment
	projectResolver.DBProjectResolver.Environment = envResolver.DBEnvironmentResolver.Environment

	helper.cleanupProjectIDs = append(helper.cleanupProjectIDs, projectResolver.DBProjectResolver.Project.Model.ID)
	return projectResolver
}

func (helper *Helper) CreateProjectWithRepo(t *testing.T, envResolver *graphql_resolver.EnvironmentResolver, gitUrl string) *graphql_resolver.ProjectResolver {
	envId := string(envResolver.ID())
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        gitUrl,
		EnvironmentID: &envId,
	}

	return helper.CreateProjectWithInput(t, envResolver, &projectInput)
}

func (helper *Helper) CreateSecret(t *testing.T,
	projectResolver *graphql_resolver.ProjectResolver) *graphql_resolver.SecretResolver {

	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

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

	secretResolver, err := helper.Resolver.CreateSecret(test.ResolverAuthContext(), &struct {
		Secret *model.SecretInput
	}{Secret: &secretInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupSecretIDs = append(helper.cleanupSecretIDs, secretResolver.DBSecretResolver.Secret.Model.ID)
	return secretResolver
}

func (helper *Helper) CreateExtension(t *testing.T, envResolver *graphql_resolver.EnvironmentResolver) *graphql_resolver.ExtensionResolver {
	envId := fmt.Sprintf("%v", envResolver.DBEnvironmentResolver.Environment.Model.ID)
	// Extension
	extensionInput := model.ExtensionInput{
		Name:          helper.name,
		Key:           "test-project-interface",
		Component:     "",
		EnvironmentID: envId,
		Config:        model.JSON{[]byte("[]")},
		Type:          "once",
	}
	extensionResolver, err := helper.Resolver.CreateExtension(&struct {
		Extension *model.ExtensionInput
	}{Extension: &extensionInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupExtensionIDs = append(helper.cleanupExtensionIDs, extensionResolver.DBExtensionResolver.Extension.Model.ID)
	return extensionResolver
}

func (helper *Helper) CreateProjectExtension(t *testing.T,
	extensionResolver *graphql_resolver.ExtensionResolver,
	projectResolver *graphql_resolver.ProjectResolver) *graphql_resolver.ProjectExtensionResolver {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	// Project Extension
	extConfigMap := make([]model.ExtConfig, 0)
	extConfigJSON, err := json.Marshal(extConfigMap)
	assert.Nil(t, err)

	extCustomConfigMap := make(map[string]model.ExtConfig)
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
	projectExtensionResolver, err := helper.Resolver.CreateProjectExtension(test.ResolverAuthContext(), &struct {
		ProjectExtension *model.ProjectExtensionInput
	}{ProjectExtension: &projectExtensionInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupProjectExtensionIDs = append(helper.cleanupProjectExtensionIDs, projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.Model.ID)
	return projectExtensionResolver
}

func (helper *Helper) CreateFeature(t *testing.T, projectResolver *graphql_resolver.ProjectResolver) *graphql_resolver.FeatureResolver {
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
	return &graphql_resolver.FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: helper.Resolver.DB, Feature: feature}}
}

func (helper *Helper) CreateFeatureWithParent(t *testing.T, projectResolver *graphql_resolver.ProjectResolver) *graphql_resolver.FeatureResolver {
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
	return &graphql_resolver.FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: helper.Resolver.DB, Feature: feature}}
}

func (helper *Helper) CreateRelease(t *testing.T,
	featureResolver *graphql_resolver.FeatureResolver,
	projectResolver *graphql_resolver.ProjectResolver) *graphql_resolver.ReleaseResolver {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	// Releases
	featureID := featureResolver.DBFeatureResolver.Feature.Model.ID.String()
	releaseInput := model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}
	releaseResolver, err := helper.Resolver.CreateRelease(test.ResolverAuthContext(), &struct{ Release *model.ReleaseInput }{Release: &releaseInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupReleaseIDs = append(helper.cleanupServiceSpecIDs, releaseResolver.DBReleaseResolver.Release.Model.ID)
	return releaseResolver
}

func (helper *Helper) CreateReleaseExtension(t *testing.T,
	releaseResolver *graphql_resolver.ReleaseResolver,
	projectExtensionResolver *graphql_resolver.ProjectExtensionResolver) *graphql_resolver.ReleaseExtensionResolver {

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

	return &graphql_resolver.ReleaseExtensionResolver{DBReleaseExtensionResolver: &db_resolver.ReleaseExtensionResolver{ReleaseExtension: releaseExtension, DB: helper.Resolver.DB}}
}

func (helper *Helper) CreateServiceSpec(t *testing.T) *graphql_resolver.ServiceSpecResolver {
	// Service Spec ID
	serviceSpecInput := model.ServiceSpecInput{
		Name:                   helper.name,
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	serviceSpecResolver, err := helper.Resolver.CreateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupServiceSpecIDs = append(helper.cleanupServiceSpecIDs, serviceSpecResolver.DBServiceSpecResolver.ServiceSpec.Model.ID)
	return serviceSpecResolver
}

func (helper *Helper) CreateService(t *testing.T,
	serviceSpecResolver *graphql_resolver.ServiceSpecResolver,
	projectResolver *graphql_resolver.ProjectResolver) *graphql_resolver.ServiceResolver {

	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	// Services
	servicePortInputs := []model.ServicePortInput{}
	serviceInput := model.ServiceInput{
		ProjectID:     projectID,
		Command:       "echo \"hello\" && exit 0",
		Name:          helper.name,
		ServiceSpecID: string(serviceSpecResolver.ID()),
		Count:         "0",
		Ports:         &servicePortInputs,
		Type:          "general",
		EnvironmentID: envID,
	}

	serviceResolver, err := helper.Resolver.CreateService(&struct{ Service *model.ServiceInput }{Service: &serviceInput})
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	helper.cleanupServiceIDs = append(helper.cleanupServiceIDs, serviceResolver.DBServiceResolver.Service.Model.ID)
	return serviceResolver
}

func (helper *Helper) TearDownTest(t *testing.T) {
	for _, id := range helper.cleanupFeatureIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Feature{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupFeatureIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupServiceIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Service{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupServiceIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupServiceSpecIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ServiceSpec{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupServiceSpecIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupReleaseExtensionIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ReleaseExtension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupReleaseExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupExtensionIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Extension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupReleaseIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Release{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupReleaseIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupProjectExtensionIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ProjectExtension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupProjectExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupProjectBookmarkIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.ProjectBookmark{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupProjectBookmarkIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupProjectIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Project{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
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
		err := helper.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupEnvironmentIDs = make([]uuid.UUID, 0)

	for _, id := range helper.cleanupSecretIDs {
		err := helper.Resolver.DB.Unscoped().Delete(&model.Secret{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	helper.cleanupSecretIDs = make([]uuid.UUID, 0)
}
