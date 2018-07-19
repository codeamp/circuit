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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
}

func (ts *ServiceTestSuite) TestCreateServiceDeploymentStrategyRollingUpdate() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Service Spec ID
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("DeploymentStrategy succesfully created with invalid parameters"))
	}
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesTCP() {
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
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)
	portTwo := int32(8080)

	healthProbeOne := model.ServiceHealthProbeInput{Method: "tcp", Port: &portOne}
	healthProbeTwo := model.ServiceHealthProbeInput{Method: "tcp", Port: &portTwo}

	readinessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}
	livenessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne, &healthProbeTwo}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesTCPInvalid() {
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
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)

	healthProbeOne := model.ServiceHealthProbeInput{Method: "tcp", Port: &portOne}
	healthProbeTwo := model.ServiceHealthProbeInput{Method: "tcp"}

	readinessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}
	livenessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne, &healthProbeTwo}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("Health Probes successfully created with invalid parameters"))
	}
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesHTTP() {
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
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)
	scheme := "http"
	path := "/healthz"

	healthProbeOne := model.ServiceHealthProbeInput{
		Method: "http",
		Port:   &portOne,
		Scheme: &scheme,
		Path:   &path,
	}

	readinessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}
	livenessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesHTTPInvalid() {
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
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)
	scheme := "invalid"
	path := "/healthz"

	healthProbeOne := model.ServiceHealthProbeInput{
		Method: "http",
		Port:   &portOne,
		Scheme: &scheme,
		Path:   &path,
	}

	readinessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}
	livenessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("Health Probes successfully created with invalid parameters"))
	}
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesExec() {
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
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	command := "./runcheck.sh"
	healthProbeOne := model.ServiceHealthProbeInput{
		Method:  "exec",
		Command: &command,
	}

	readinessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}
	livenessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesExecInvalid() {
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
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	healthProbeOne := model.ServiceHealthProbeInput{Method: "exec"}

	readinessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}
	livenessProbes := []*model.ServiceHealthProbeInput{&healthProbeOne}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("Health Probes successfully created with invalid parameters"))
	}
}

func (ts *ServiceTestSuite) TestUpdateService() {
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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)

	// Update Service
	serviceID := string(serviceResolver.ID())
	projectID := string(projectResolver.ID())
	serviceSpecID := string(serviceSpecResolver.ID())

	serviceInput := &model.ServiceInput{
		ID:            &serviceID,
		ProjectID:     projectID,
		ServiceSpecID: serviceSpecID,
		DeploymentStrategy: &model.DeploymentStrategyInput{
			Type:           plugins.GetType("rollingUpdate"),
			MaxUnavailable: 30,
			MaxSurge:       60,
		},
	}
	_, err = ts.Resolver.UpdateService(&struct{ Service *model.ServiceInput }{serviceInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ServiceTestSuite) TestDeleteService() {
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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)

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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)

	// Test Service Interface
	_ = serviceResolver.ID()
	serviceProjectResolver := serviceResolver.Project()
	assert.Equal(ts.T(), projectResolver.ID(), serviceProjectResolver.ID())

	assert.Equal(ts.T(), "echo \"hello\" && exit 0", serviceResolver.Command())
	assert.Equal(ts.T(), "TestService", serviceResolver.Name())

	serviceSpecResolver = serviceResolver.ServiceSpec()
	assert.Equal(ts.T(), serviceSpecID, serviceSpecResolver.ID())

	assert.Equal(ts.T(), int32(1), serviceResolver.Count())

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

	livenessProbes := []*model.ServiceHealthProbeInput{}
	readinessProbes := []*model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy,
		&readinessProbes, &livenessProbes)

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
}

func TestSuiteServiceResolver(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
