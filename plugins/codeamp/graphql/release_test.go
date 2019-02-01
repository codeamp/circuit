package graphql_resolver_test

import (
	"context"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm/dialects/postgres"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	graphql "github.com/graph-gophers/graphql-go"
)

type ReleaseTestSuite struct {
	suite.Suite
	Resolver   *graphql_resolver.Resolver
	helper     Helper
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
redis:
  username:
  password:
  server: "redis:6379"
  database: "0"
  pool: "30"
  process: "1"
plugins:
  codeamp:
    workers: 1
    postgres:
      host: "postgres"
      port: "5432"
      user: "postgres"
      dbname: "codeamp"
      sslmode: "disable"
      password: ""
    service_address: ":3011"
`)

func (suite *ReleaseTestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin {
		return &codeamp.CodeAmp{
			Events: suite.transistor.TestEvents,
		}
	}, plugins.ReleaseExtension{}, plugins.Release{})
	go suite.transistor.Run()
}

func (suite *ReleaseTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
		&model.SecretValue{},
		&model.Release{},
		&model.User{},
		&model.ServiceDeploymentStrategy{},
		&model.ServiceHealthProbe{},
		&model.ServicePort{},
		&model.ServiceHealthProbeHttpHeader{},
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

	// Service
	ts.helper.CreateServiceSpec(ts.T(), true)

	portOne := int32(9090)
	scheme := "http"
	path := "/healthz"

	headers := []model.HealthProbeHttpHeaderInput{
		{
			Name:  "X-Forwarded-Proto",
			Value: "https",
		},
		{
			Name:  "X-Forwarded-For",
			Value: "www.example.com",
		},
	}

	healthProbe := model.ServiceHealthProbeInput{
		Method:      "http",
		Port:        &portOne,
		Scheme:      &scheme,
		Path:        &path,
		HttpHeaders: &headers,
	}

	readinessProbe := healthProbe
	livenessProbe := healthProbe
	preStopHook := "sleep 15"

	ts.helper.CreateService(ts.T(), projectResolver, nil, &readinessProbe, &livenessProbe, &preStopHook)

	// Make Project Secret
	envID := string(environmentResolver.ID())
	projectID := string(projectResolver.ID())
	secretInput := model.SecretInput{
		Key:           "TestCreateReleaseSuccess-project",
		Type:          "env",
		Scope:         "project",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Make Global Secret
	secretInput = model.SecretInput{
		Key:           "TestCreateReleaseSuccess-global",
		Type:          "env",
		Scope:         "global",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
}

func (ts *ReleaseTestSuite) TestCreateReleaseSuccessNoTailFeature() {
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

	// Service
	ts.helper.CreateServiceSpec(ts.T(), true)
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

	// Make Project Secret
	envID := string(environmentResolver.ID())
	projectID := string(projectResolver.ID())
	secretInput := model.SecretInput{
		Key:           "TestCreateReleaseSuccess-project",
		Type:          "env",
		Scope:         "project",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Make Global Secret
	secretInput = model.SecretInput{
		Key:           "TestCreateReleaseSuccess-global",
		Type:          "env",
		Scope:         "global",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Remove all features so as to force CreateRelease to not use a tail feature id
	err = ts.Resolver.DB.Where("project_id = ?", string(projectResolver.ID())).Delete(&model.Release{}).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
}

func (ts *ReleaseTestSuite) TestCreateRelease_CacheableUsed() {
	log.Info("TestCreateReleaseSuccess_SecondReleaseUsesCachedReleaseExtension")

	// pre-requisites for a release with one release extension
	envResolver := ts.helper.CreateEnvironment(ts.T())
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)
	extension := extensionResolver.DBExtensionResolver.Extension
	extension.Cacheable = true
	extension.Key = "dockerbuilder"

	ts.Resolver.DB.Save(&extension)

	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	userResolver := ts.helper.CreateUser(ts.T())

	// manually create release object and corresponding release extension
	// since we don't want to trigger events from the CreateRelease mutation
	release := model.Release{
		State:             transistor.GetState("complete"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{},
		Secrets:           postgres.Jsonb{},
		ProjectExtensions: postgres.Jsonb{},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		Finished:          time.Now(),
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&release)

	releaseExtension := model.ReleaseExtension{
		ReleaseID:          release.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		State:              transistor.GetState("complete"),
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts: postgres.Jsonb{[]byte(`[{
			"key": "key1",
			"value": "val1",
			"secret": false,
			"allowOverride": false
		}]`)},
		Started: time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtension)

	releaseThatUsesCacheable := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{[]byte(`[]`)},
		Secrets:           postgres.Jsonb{[]byte(`[]`)},
		ProjectExtensions: postgres.Jsonb{[]byte(`[]`)},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&releaseThatUsesCacheable)

	releaseExtensionThatUsesCacheable := model.ReleaseExtension{
		ReleaseID:          releaseThatUsesCacheable.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		State:              transistor.GetState("waiting"),
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts:          postgres.Jsonb{[]byte(`[]`)},
		Started:            time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtensionThatUsesCacheable)

	eventPayload := plugins.Release{
		ID:      releaseThatUsesCacheable.Model.ID.String(),
		Project: plugins.Project{},
		Git:     plugins.Git{},
		HeadFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		User: userResolver.DBUserResolver.Email,
		TailFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		Services:    []plugins.Service{},
		Secrets:     []plugins.Secret{},
		Environment: envResolver.Key(),
		IsRollback:  releaseThatUsesCacheable.IsRollback,
	}

	// send release:create event with payload
	ev := transistor.NewEvent("release", "create", eventPayload)

	ts.transistor.Events <- ev

	// get the dockerbuilder:status event in TestEvents and confirm the State is complete
	e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("status"), 60)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	cachedArtifact, err := e.GetArtifact("key1")
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(ts.T(), transistor.GetState("complete"), e.State)
	assert.Equal(ts.T(), "val1", cachedArtifact.String())
}

func (ts *ReleaseTestSuite) TestCreateRelease_CacheableNotUsed_DifferentServicesSignature() {
	// pre-requisites for a release with one release extension
	envResolver := ts.helper.CreateEnvironment(ts.T())
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)
	extension := extensionResolver.DBExtensionResolver.Extension
	extension.Cacheable = true
	extension.Key = "dockerbuilder"

	ts.Resolver.DB.Save(&extension)

	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	userResolver := ts.helper.CreateUser(ts.T())

	// manually create release object and corresponding release extension
	// since we don't want to trigger events from the CreateRelease mutation
	release := model.Release{
		State:             transistor.GetState("complete"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{},
		Secrets:           postgres.Jsonb{},
		ProjectExtensions: postgres.Jsonb{},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		Finished:          time.Now(),
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&release)

	releaseExtension := model.ReleaseExtension{
		ReleaseID:          release.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		State:              transistor.GetState("complete"),
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts: postgres.Jsonb{[]byte(`[{
			"key": "key1",
			"value": "val1",
			"secret": false,
			"allowOverride": false
		}]`)},
		Started: time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtension)

	releaseThatUsesCacheable := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{[]byte(`[]`)},
		Secrets:           postgres.Jsonb{[]byte(`[]`)},
		ProjectExtensions: postgres.Jsonb{[]byte(`[]`)},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&releaseThatUsesCacheable)

	releaseExtensionThatUsesCacheable := model.ReleaseExtension{
		ReleaseID:   releaseThatUsesCacheable.Model.ID,
		FeatureHash: "featurehash",
		// different services signature so cached release extension should not be used
		ServicesSignature:  "differentsignature",
		SecretsSignature:   "signature",
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		State:              transistor.GetState("waiting"),
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts:          postgres.Jsonb{[]byte(`[]`)},
		Started:            time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtensionThatUsesCacheable)

	eventPayload := plugins.Release{
		ID:      releaseThatUsesCacheable.Model.ID.String(),
		Project: plugins.Project{},
		Git:     plugins.Git{},
		HeadFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		User: userResolver.DBUserResolver.Email,
		TailFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		Services:    []plugins.Service{},
		Secrets:     []plugins.Secret{},
		Environment: envResolver.Key(),
		IsRollback:  releaseThatUsesCacheable.IsRollback,
	}

	// send release:create event with payload
	ev := transistor.NewEvent("release", "create", eventPayload)

	ts.transistor.Events <- ev

	// get the dockerbuilder:status event in TestEvents and confirm the State is complete
	e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("create"), 60)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), transistor.GetState("waiting"), e.State)
}

func (ts *ReleaseTestSuite) TestCreateRelease_CacheableNotUsed_DifferentSecretsSignature() {
	// pre-requisites for a release with one release extension
	envResolver := ts.helper.CreateEnvironment(ts.T())
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)
	extension := extensionResolver.DBExtensionResolver.Extension
	extension.Cacheable = true
	extension.Key = "dockerbuilder"

	ts.Resolver.DB.Save(&extension)

	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	userResolver := ts.helper.CreateUser(ts.T())

	// manually create release object and corresponding release extension
	// since we don't want to trigger events from the CreateRelease mutation
	release := model.Release{
		State:             transistor.GetState("complete"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{},
		Secrets:           postgres.Jsonb{},
		ProjectExtensions: postgres.Jsonb{},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		Finished:          time.Now(),
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&release)

	releaseExtension := model.ReleaseExtension{
		ReleaseID:          release.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		State:              transistor.GetState("complete"),
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts: postgres.Jsonb{[]byte(`[{
			"key": "key1",
			"value": "val1",
			"secret": false,
			"allowOverride": false
		}]`)},
		Started: time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtension)

	releaseThatUsesCacheable := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{[]byte(`[]`)},
		Secrets:           postgres.Jsonb{[]byte(`[]`)},
		ProjectExtensions: postgres.Jsonb{[]byte(`[]`)},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&releaseThatUsesCacheable)

	releaseExtensionThatUsesCacheable := model.ReleaseExtension{
		ReleaseID:   releaseThatUsesCacheable.Model.ID,
		FeatureHash: "featurehash",
		// different services signature so cached release extension should not be used
		ServicesSignature:  "signature",
		SecretsSignature:   "differentsignature",
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		State:              transistor.GetState("waiting"),
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts:          postgres.Jsonb{[]byte(`[]`)},
		Started:            time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtensionThatUsesCacheable)

	eventPayload := plugins.Release{
		ID:      releaseThatUsesCacheable.Model.ID.String(),
		Project: plugins.Project{},
		Git:     plugins.Git{},
		HeadFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		User: userResolver.DBUserResolver.Email,
		TailFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		Services:    []plugins.Service{},
		Secrets:     []plugins.Secret{},
		Environment: envResolver.Key(),
		IsRollback:  releaseThatUsesCacheable.IsRollback,
	}

	// send release:create event with payload
	ev := transistor.NewEvent("release", "create", eventPayload)

	ts.transistor.Events <- ev

	// get the dockerbuilder:status event in TestEvents and confirm the State is complete
	e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("create"), 60)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), transistor.GetState("waiting"), e.State)
}

func (ts *ReleaseTestSuite) TestCreateRelease_CacheableNotUsed_DifferentFeatureHash() {
	// pre-requisites for a release with one release extension
	envResolver := ts.helper.CreateEnvironment(ts.T())
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)
	extension := extensionResolver.DBExtensionResolver.Extension
	extension.Cacheable = true
	extension.Key = "dockerbuilder"

	ts.Resolver.DB.Save(&extension)

	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	userResolver := ts.helper.CreateUser(ts.T())

	// manually create release object and corresponding release extension
	// since we don't want to trigger events from the CreateRelease mutation
	release := model.Release{
		State:             transistor.GetState("complete"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{},
		Secrets:           postgres.Jsonb{},
		ProjectExtensions: postgres.Jsonb{},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		Finished:          time.Now(),
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&release)

	releaseExtension := model.ReleaseExtension{
		ReleaseID:          release.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		State:              transistor.GetState("complete"),
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts: postgres.Jsonb{[]byte(`[{
			"key": "key1",
			"value": "val1",
			"secret": false,
			"allowOverride": false
		}]`)},
		Started: time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtension)

	releaseThatUsesCacheable := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{[]byte(`[]`)},
		Secrets:           postgres.Jsonb{[]byte(`[]`)},
		ProjectExtensions: postgres.Jsonb{[]byte(`[]`)},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&releaseThatUsesCacheable)

	releaseExtensionThatUsesCacheable := model.ReleaseExtension{
		ReleaseID:   releaseThatUsesCacheable.Model.ID,
		FeatureHash: "differentfeaturehash",
		// different services signature so cached release extension should not be used
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		State:              transistor.GetState("waiting"),
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts:          postgres.Jsonb{[]byte(`[]`)},
		Started:            time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtensionThatUsesCacheable)

	eventPayload := plugins.Release{
		ID:      releaseThatUsesCacheable.Model.ID.String(),
		Project: plugins.Project{},
		Git:     plugins.Git{},
		HeadFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		User: userResolver.DBUserResolver.Email,
		TailFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		Services:    []plugins.Service{},
		Secrets:     []plugins.Secret{},
		Environment: envResolver.Key(),
		IsRollback:  releaseThatUsesCacheable.IsRollback,
	}

	// send release:create event with payload
	ev := transistor.NewEvent("release", "create", eventPayload)

	ts.transistor.Events <- ev

	// get the dockerbuilder:status event in TestEvents and confirm the State is complete
	e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("create"), 60)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), transistor.GetState("waiting"), e.State)
}

func (ts *ReleaseTestSuite) TestCreateRelease_CacheableNotUsed_ForceRebuildIsTrue() {
	// pre-requisites for a release with one release extension
	envResolver := ts.helper.CreateEnvironment(ts.T())
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)
	extension := extensionResolver.DBExtensionResolver.Extension
	extension.Cacheable = true
	extension.Key = "dockerbuilder"

	ts.Resolver.DB.Save(&extension)

	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	userResolver := ts.helper.CreateUser(ts.T())

	// manually create release object and corresponding release extension
	// since we don't want to trigger events from the CreateRelease mutation
	release := model.Release{
		State:             transistor.GetState("complete"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{},
		Secrets:           postgres.Jsonb{},
		ProjectExtensions: postgres.Jsonb{},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		Finished:          time.Now(),
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&release)

	releaseExtension := model.ReleaseExtension{
		ReleaseID:          release.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		State:              transistor.GetState("complete"),
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts: postgres.Jsonb{[]byte(`[{
			"key": "key1",
			"value": "val1",
			"secret": false,
			"allowOverride": false
		}]`)},
		Started: time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtension)

	releaseThatUsesCacheable := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{[]byte(`[]`)},
		Secrets:           postgres.Jsonb{[]byte(`[]`)},
		ProjectExtensions: postgres.Jsonb{[]byte(`[]`)},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		// ForceRebuild is true, so cacheable should not be used
		ForceRebuild: true,
		IsRollback:   false,
	}

	ts.Resolver.DB.Save(&releaseThatUsesCacheable)

	releaseExtensionThatUsesCacheable := model.ReleaseExtension{
		ReleaseID:   releaseThatUsesCacheable.Model.ID,
		FeatureHash: "featurehash",
		// different services signature so cached release extension should not be used
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		State:              transistor.GetState("waiting"),
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts:          postgres.Jsonb{[]byte(`[]`)},
		Started:            time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtensionThatUsesCacheable)

	eventPayload := plugins.Release{
		ID:      releaseThatUsesCacheable.Model.ID.String(),
		Project: plugins.Project{},
		Git:     plugins.Git{},
		HeadFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		User: userResolver.DBUserResolver.Email,
		TailFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		Services:    []plugins.Service{},
		Secrets:     []plugins.Secret{},
		Environment: envResolver.Key(),
		IsRollback:  releaseThatUsesCacheable.IsRollback,
	}

	// send release:create event with payload
	ev := transistor.NewEvent("release", "create", eventPayload)

	ts.transistor.Events <- ev

	// get the dockerbuilder:status event in TestEvents and confirm the State is complete
	e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("create"), 60)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), transistor.GetState("waiting"), e.State)
}

func (ts *ReleaseTestSuite) TestCreateRelease_CacheableNotUsed_IsTurnedOff() {
	// pre-requisites for a release with one release extension
	envResolver := ts.helper.CreateEnvironment(ts.T())
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)
	extension := extensionResolver.DBExtensionResolver.Extension
	// cacheable is false so workflow release extensions will always be created in the build process
	extension.Cacheable = false
	extension.Key = "dockerbuilder"

	ts.Resolver.DB.Save(&extension)

	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	userResolver := ts.helper.CreateUser(ts.T())

	// manually create release object and corresponding release extension
	// since we don't want to trigger events from the CreateRelease mutation
	release := model.Release{
		State:             transistor.GetState("complete"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{},
		Secrets:           postgres.Jsonb{},
		ProjectExtensions: postgres.Jsonb{},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		Finished:          time.Now(),
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&release)

	releaseExtension := model.ReleaseExtension{
		ReleaseID:          release.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		State:              transistor.GetState("complete"),
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts: postgres.Jsonb{[]byte(`[{
			"key": "key1",
			"value": "val1",
			"secret": false,
			"allowOverride": false
		}]`)},
		Started: time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtension)

	releaseThatUsesCacheable := model.Release{
		State:             transistor.GetState("waiting"),
		StateMessage:      "",
		ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID,
		UserID:            userResolver.DBUserResolver.User.Model.ID,
		HeadFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		TailFeatureID:     featureResolver.DBFeatureResolver.Feature.Model.ID,
		Services:          postgres.Jsonb{[]byte(`[]`)},
		Secrets:           postgres.Jsonb{[]byte(`[]`)},
		ProjectExtensions: postgres.Jsonb{[]byte(`[]`)},
		EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID,
		ForceRebuild:      false,
		IsRollback:        false,
	}

	ts.Resolver.DB.Save(&releaseThatUsesCacheable)

	releaseExtensionThatUsesCacheable := model.ReleaseExtension{
		ReleaseID:          releaseThatUsesCacheable.Model.ID,
		FeatureHash:        "featurehash",
		ServicesSignature:  "signature",
		SecretsSignature:   "signature",
		ProjectExtensionID: projectExtensionResolver.DBProjectExtensionResolver.Model.ID,
		State:              transistor.GetState("waiting"),
		StateMessage:       "",
		Type:               plugins.GetType("workflow"),
		Artifacts:          postgres.Jsonb{[]byte(`[]`)},
		Started:            time.Now(),
	}

	ts.Resolver.DB.Save(&releaseExtensionThatUsesCacheable)

	eventPayload := plugins.Release{
		ID:      releaseThatUsesCacheable.Model.ID.String(),
		Project: plugins.Project{},
		Git:     plugins.Git{},
		HeadFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		User: userResolver.DBUserResolver.Email,
		TailFeature: plugins.Feature{
			ID:         featureResolver.DBFeatureResolver.Feature.Model.ID.String(),
			Hash:       featureResolver.DBFeatureResolver.Feature.Hash,
			ParentHash: featureResolver.DBFeatureResolver.Feature.ParentHash,
			User:       userResolver.DBUserResolver.Email,
			Message:    featureResolver.DBFeatureResolver.Feature.Message,
			Created:    releaseThatUsesCacheable.CreatedAt,
		},
		Services:    []plugins.Service{},
		Secrets:     []plugins.Secret{},
		Environment: envResolver.Key(),
		IsRollback:  releaseThatUsesCacheable.IsRollback,
	}

	// send release:create event with payload
	ev := transistor.NewEvent("release", "create", eventPayload)

	ts.transistor.Events <- ev

	// get the dockerbuilder:status event in TestEvents and confirm the State is complete
	e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("create"), 60)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), transistor.GetState("waiting"), e.State)
}

// 	// Project
// 	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Secret
// 	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

// 	// Extension
// 	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

// 	// Project Extension
// 	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

// 	// Force to set to 'complete' state for testing purposes
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
// 	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

// 	// Feature
// 	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

// 	// Service
// 	ts.helper.CreateServiceSpec(ts.T(), true)
// 	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

// 	// Make Project Secret
// 	envID := string(environmentResolver.ID())
// 	projectID := string(projectResolver.ID())
// 	secretInput := model.SecretInput{
// 		Key:           "TestCreateReleaseSuccess-project",
// 		Type:          "env",
// 		Scope:         "project",
// 		EnvironmentID: envID,
// 		ProjectID:     &projectID,
// 		IsSecret:      false,
// 	}
// 	_, err = ts.helper.CreateSecretWithInput(&secretInput)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Make Global Secret
// 	secretInput = model.SecretInput{
// 		Key:           "TestCreateReleaseSuccess-global",
// 		Type:          "env",
// 		Scope:         "global",
// 		EnvironmentID: envID,
// 		ProjectID:     &projectID,
// 		IsSecret:      false,
// 	}
// 	_, err = ts.helper.CreateSecretWithInput(&secretInput)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Release
// 	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

// 	featureID := string(featureResolver.ID())
// 	releaseInput := &model.ReleaseInput{
// 		HeadFeatureID: featureID,
// 		ProjectID:     projectID,
// 		EnvironmentID: envID,
// 		ForceRebuild:  false,
// 	}
// 	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
// 	assert.NotNil(ts.T(), err)
// }

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoAuth() {
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
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Release
	var ctx context.Context
	ts.helper.SetContext(ctx)
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)

	ts.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoProjectExtensions() {
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
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Delete Project Extension
	err = ts.Resolver.DB.Delete(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

// func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoReleaseExtensions() {
// 	// Environment
// 	environmentResolver := ts.helper.CreateEnvironment(ts.T())

// 	// Project
// 	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Secret
// 	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

// 	// Extension
// 	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)
// 	extensionResolver.DBExtensionResolver.Extension.Type = "once"
// 	ts.Resolver.DB.Save(&extensionResolver.DBExtensionResolver.Extension)

// 	// Project Extension
// 	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

// 	// Force to set to 'complete' state for testing purposes
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
// 	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

// 	// Feature
// 	projectID := string(projectResolver.ID())
// 	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
// 	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

// 	featureID := string(featureResolver.ID())
// 	releaseInput := &model.ReleaseInput{
// 		HeadFeatureID: featureID,
// 		ProjectID:     projectID,
// 		EnvironmentID: envID,
// 		ForceRebuild:  false,
// 	}

// 	// Release
// 	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
// 	assert.NotNil(ts.T(), err)
// }

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoPermission() {
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
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Delete ProjectEnvironment
	err = ts.Resolver.DB.Where("project_id = ? and environment_id = ?", projectResolver.ID(), environmentResolver.ID()).Delete(&model.ProjectEnvironment{}).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

// ADB - 7/18/2018
// This error condition needs more work on the CreateRelease side.
// It's not clear how it is supposed to work when comparing secrets and services signatures
//
// func (ts *ReleaseTestSuite) TestCreateReleaseFailureSameSignatures() {
// 	// Environment
// 	environmentResolver := ts.helper.CreateEnvironment(ts.T())

// 	// Project
// 	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Secret
// 	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

// 	// Extension
// 	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

// 	// Project Extension
// 	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

// 	// Force to set to 'complete' state for testing purposes
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
// 	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

// 	// Feature
// 	projectID := string(projectResolver.ID())
// 	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
// 	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

// 	featureID := string(featureResolver.ID())
// 	releaseInput := &model.ReleaseInput{
// 		HeadFeatureID: featureID,
// 		ProjectID:     projectID,
// 		EnvironmentID: envID,
// 		ForceRebuild:  false,
// 	}

// 	// Release
// 	ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, releaseInput)

// 	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
// 	log.Error(err)
// 	assert.NotNil(ts.T(), err)
// }

func (ts *ReleaseTestSuite) TestCreateReleaseFailureBadHeadFeature() {
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
	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx"
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureBadEnvID() {
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
	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	releaseInput := &model.ReleaseInput{
		HeadFeatureID: test.InvalidUUID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureBadProjectID() {
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

	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()
	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     test.InvalidUUID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
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

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureBadEnvironment() {
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
		EnvironmentID: test.InvalidUUID,
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureBadReleaseID() {
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
	releaseID := test.ValidUUID
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

// func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureReleaseInProgress() {
// 	// Environment
// 	environmentResolver := ts.helper.CreateEnvironment(ts.T())

// 	// Project
// 	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Secret
// 	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

// 	// Extension
// 	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

// 	// Project Extension
// 	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

// 	// Force to set to 'complete' state for testing purposes
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
// 	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

// 	// Feature
// 	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

// 	// Release
// 	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

// 	// Rollback
// 	releaseID := string(releaseResolver.ID())
// 	releaseInput := model.ReleaseInput{
// 		ID:            &releaseID,
// 		ProjectID:     string(projectResolver.ID()),
// 		HeadFeatureID: string(featureResolver.ID()),
// 		EnvironmentID: string(environmentResolver.ID()),
// 	}
// 	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
// 	assert.NotNil(ts.T(), err)
// }

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

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

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

func (ts *ReleaseTestSuite) TestStopReleaseFailureBadProjectExtension() {
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

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// Delete the project resolver to trigger the error condition
	ts.Resolver.DB.Delete(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)

}

func (ts *ReleaseTestSuite) TestStopReleaseFailureReleaseNotFound() {
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

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// Delete the release
	err = ts.Resolver.DB.Where("id = ?", string(releaseResolver.ID())).Delete(&model.Release{}).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureBadReleaseExtension() {
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

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// Delete the project resolver to trigger the error condition
	ts.Resolver.DB.Delete(&extensionResolver.DBExtensionResolver.Extension)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureNoReleaseExtensions() {
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
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateRollbackReleaseSuccess() {
	var e transistor.Event
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

	// ServiceSpec
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Service
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
	for len(ts.Resolver.Events) > 0 {
		e = <-ts.Resolver.Events
	}

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	log.Warn("Stopping rleease 1")
	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	for len(ts.Resolver.Events) > 0 {
		log.Warn("dumping event 2")
		<-ts.Resolver.Events
	}

	releaseID := string(releaseResolver.ID())

	// Rollback the deploy
	log.Warn("Creating rollback release")
	releaseResolver = ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, &model.ReleaseInput{
		ID:            &releaseID,
		HeadFeatureID: string(featureResolver.ID()),
		ProjectID:     string(projectResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	})
	log.Warn("Waiting for event")
	e = <-ts.Resolver.Events

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// releaseExtensionEvent := e.Payload.(plugins.ReleaseExtension)
	// release := releaseExtensionEvent.Release
	// assert.Equal(ts.T(), true, release.IsRollback)
	// assert.Equal(ts.T(), true, releaseResolver.IsRollback())

	log.Warn("Stopping release 2")
	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ReleaseTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func (ts *ReleaseTestSuite) TearDownSuite() {
	ts.transistor.Stop()
}

func TestSuiteReleaseResolver(t *testing.T) {
	suite.Run(t, new(ReleaseTestSuite))
}
