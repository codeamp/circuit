package graphql_resolver_test

import (
	"context"
	"fmt"

	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	_ "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	Resolver        *graphql_resolver.Resolver
	ServiceResolver *graphql_resolver.ServiceResolver

	helper Helper
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

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestService")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ServiceTestSuite) TestCreateServiceSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)
}

func (ts *ServiceTestSuite) TestCreateServiceDeploymentStrategyDefault() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("default"),
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)
}

func (ts *ServiceTestSuite) TestCreateServiceDeploymentStrategyRecreate() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)
}

func (ts *ServiceTestSuite) TestCreateServiceDeploymentStrategyRollingUpdate() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: "30",
		MaxSurge:       "60",
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)
}

func (ts *ServiceTestSuite) TestCreateServiceDeploymentStrategyRollingUpdateFailure() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("rollingUpdate"),
	}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("DeploymentStrategy of type rollingUpdate created with invalid inputs"))
	}
}

func (ts *ServiceTestSuite) TestCreateServiceDeploymentStrategyInvalid() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.Type("invalidStrategy"),
	}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("DeploymentStrategy created with invalid parameters"))
	}
}

func (ts *ServiceTestSuite) TestUpdateServiceSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)

	// Update Service
	serviceID := string(serviceResolver.ID())

	servicePorts := []model.ServicePortInput{
		model.ServicePortInput{
			Port:     "80",
			Protocol: "HTTP",
		},
	}
	serviceInput := &model.ServiceInput{
		ID:            &serviceID,
		ProjectID:     string(projectResolver.ID()),
		ServiceSpecID: string(serviceSpecResolver.ID()),
		DeploymentStrategy: &model.DeploymentStrategyInput{
			Type:           plugins.GetType("rollingUpdate"),
			MaxUnavailable: "30",
			MaxSurge:       "60",
		},
		Ports: &servicePorts,
	}
	_, err = ts.Resolver.UpdateService(&struct{ Service *model.ServiceInput }{serviceInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ServiceTestSuite) TestUpdateServiceFailureNullID() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver)

	// Update Service
	serviceID := "null"
	serviceInput := &model.ServiceInput{ID: &serviceID}
	_, err = ts.Resolver.UpdateService(&struct{ Service *model.ServiceInput }{serviceInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ServiceTestSuite) TestUpdateServiceFailureBadRecordID() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver)

	// Update Service
	serviceID := test.ValidUUID
	serviceInput := &model.ServiceInput{ID: &serviceID}
	_, err = ts.Resolver.UpdateService(&struct{ Service *model.ServiceInput }{serviceInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ServiceTestSuite) TestDeleteServiceSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)

	// Update Service
	serviceID := string(serviceResolver.ID())
	projectID := string(projectResolver.ID())
	serviceSpecID := string(serviceSpecResolver.ID())

	serviceInput := &model.ServiceInput{
		ID:            &serviceID,
		ProjectID:     projectID,
		ServiceSpecID: serviceSpecID,
	}
	_, err = ts.Resolver.DeleteService(&struct{ Service *model.ServiceInput }{serviceInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ServiceTestSuite) TestDeleteServiceFailure() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver)

	// Update Service
	serviceID := "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx"
	serviceInput := &model.ServiceInput{
		ID: &serviceID,
	}
	_, err = ts.Resolver.DeleteService(&struct{ Service *model.ServiceInput }{serviceInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ServiceTestSuite) TestServiceInterface() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())
	serviceSpecID := serviceSpecResolver.ID()

	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)

	// Test Service Interface
	_ = serviceResolver.ID()
	serviceProjectResolver := serviceResolver.Project()
	assert.Equal(ts.T(), projectResolver.ID(), serviceProjectResolver.ID())

	assert.Equal(ts.T(), "echo \"hello\" && exit 0", serviceResolver.Command())
	assert.Equal(ts.T(), "TestService", serviceResolver.Name())

	serviceSpecResolver = serviceResolver.ServiceSpec()
	assert.Equal(ts.T(), serviceSpecID, serviceSpecResolver.ID())

	assert.Equal(ts.T(), "1", serviceResolver.Count())

	servicePorts, err := serviceResolver.Ports()
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), servicePorts, "Service Ports was empty")

	assert.Equal(ts.T(), "general", serviceResolver.Type())
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
}

func (ts *ServiceTestSuite) TestServiceQuery() {
	emptyPaginatorInput := &struct{ Params *model.PaginatorInput }{}

	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)

	// Test Service Query
	var ctx context.Context
	_, err = ts.Resolver.Services(ctx, emptyPaginatorInput)
	assert.NotNil(ts.T(), err)

	serviceResolvers, err := ts.Resolver.Services(test.ResolverAuthContext(), emptyPaginatorInput)
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), serviceResolvers)
	assert.NotEmpty(ts.T(), serviceResolvers, "Service Resolvers was empty")
}

func (ts *ServiceTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteServiceResolver(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
