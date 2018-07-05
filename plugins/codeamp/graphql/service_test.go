package graphql_resolver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	Resolver        *graphql_resolver.Resolver
	ServiceResolver *graphql_resolver.ServiceResolver

	cleanupEnvironmentIDs []uuid.UUID
	cleanupProjectIDs     []uuid.UUID
	cleanupServiceIDs     []uuid.UUID
	cleanupServiceSpecIDs []uuid.UUID
}

func (suite *ServiceTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectEnvironment{},
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}
	_ = &graphql_resolver.Resolver{DB: db, Events: nil, Redis: nil}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.ServiceResolver = &graphql_resolver.ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: db}}
}

func (ts *ServiceTestSuite) TestCreateService() {
	emptyPaginatorInput := &struct {
		Params *model.PaginatorInput
	}{nil}

	// Environment
	envInput := model.EnvironmentInput{
		Name:      "TestProjectInterface",
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
	projectId := string(createProjectResolver.ID())

	// TODO: ADB This should be happening in the CreateProject function!
	// If an ID for an Environment is supplied, Project should try to look that up and return resolver
	// that includes project AND environment
	createProjectResolver.DBProjectResolver.Environment = envResolver.DBEnvironmentResolver.Environment
	ts.cleanupProjectIDs = append(ts.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.Model.ID)

	// Service Spec ID
	serviceSpecInput := model.ServiceSpecInput{
		Name:                   "test",
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	serviceSpecResolver, err := ts.Resolver.CreateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupServiceSpecIDs = append(ts.cleanupServiceSpecIDs, serviceSpecResolver.DBServiceSpecResolver.ServiceSpec.Model.ID)
	serviceSpecID := serviceSpecResolver.ID()

	// Services
	servicePortInputs := []model.ServicePortInput{
		model.ServicePortInput{
			Port:     "80",
			Protocol: "HTTP",
		},
	}
	serviceInput := model.ServiceInput{
		ProjectID:     projectId,
		Command:       "echo \"hello\" && exit 0",
		Name:          "test-service",
		ServiceSpecID: string(serviceSpecResolver.ID()),
		Count:         "0",
		Ports:         &servicePortInputs,
		Type:          "general",
		EnvironmentID: envId,
	}

	serviceResolver, err := ts.Resolver.CreateService(&struct{ Service *model.ServiceInput }{Service: &serviceInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupServiceIDs = append(ts.cleanupServiceIDs, serviceResolver.DBServiceResolver.Service.Model.ID)

	// Test Service Interface
	_ = serviceResolver.ID()
	projectResolver := serviceResolver.Project()
	assert.Equal(ts.T(), createProjectResolver.ID(), projectResolver.ID())

	assert.Equal(ts.T(), serviceInput.Command, serviceResolver.Command())
	assert.Equal(ts.T(), serviceInput.Name, serviceResolver.Name())

	serviceSpecResolver = serviceResolver.ServiceSpec()
	assert.Equal(ts.T(), serviceSpecID, serviceSpecResolver.ID())

	assert.Equal(ts.T(), serviceInput.Count, serviceResolver.Count())

	servicePorts, err := serviceResolver.Ports()
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), servicePorts, "Service Ports was empty")

	assert.Equal(ts.T(), serviceInput.Type, serviceResolver.Type())
	created_at_diff := time.Now().Sub(serviceResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	var ctx context.Context
	_, err = serviceResolver.Environment(ctx)
	log.Error(err)
	assert.NotNil(ts.T(), err)

	serviceEnvironment, err := serviceResolver.Environment(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), envResolver.ID(), serviceEnvironment.ID())

	data, err := serviceResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = serviceResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)

	// Test Service Query
	_, err = ts.Resolver.Services(ctx, emptyPaginatorInput)
	assert.NotNil(ts.T(), err)

	serviceResolvers, err := ts.Resolver.Services(test.ResolverAuthContext(), emptyPaginatorInput)
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), serviceResolvers)
	assert.NotEmpty(ts.T(), serviceResolvers, "Service Resolvers was empty")
}

func (ts *ServiceTestSuite) TearDownTest() {
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
}

func TestSuiteServiceResolver(t *testing.T) {
	ts := new(ServiceTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
