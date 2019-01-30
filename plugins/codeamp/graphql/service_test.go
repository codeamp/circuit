package graphql_resolver_test

import (
	"context"
	"fmt"

	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	_ "github.com/satori/go.uuid"
	yaml "gopkg.in/yaml.v2"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var validYAMLServicesString = `
- name: service_name_1
  command: npm start
  type: general
  count: 2
  ports:
    - protocol: TCP
      port: 8000
  deploymentStrategy:
    type: recreate
    maxUnavailable: 70
    maxSurge: 30
  readinessProbe:
    type: readinessProbe # should this be omitted since we already specify it one layer up?
    method: exec
    command: exec
    port: 8080
    scheme: http
    path: /
    httpHeaders:
      - name: foo  
        value: val
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    successThreshold: 10
    failureThreshold: 10
  livenessProbe:
    type: livenessProbe # should this be omitted since we already specify it one layer up?
    method: exec
    command: exec
    port: 8080
    scheme: http
    path: /
    httpHeaders:
      - name: foo  
        value: val
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    successThreshold: 10
    failureThreshold: 10    
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    successThreshold: 10
    failureThreshold: 10
  preStopHook: service_name_1
- name: service_name_2
  command: python app.py
  type: general
  count: 2
  ports:
    - protocol: TCP
      port: 8000
  deploymentStrategy:
    type: recreate
    maxUnavailable: 70
    maxSurge: 30
  readinessProbe:
    type: readinessProbe # should this be omitted since we already specify it one layer up?
    method: exec
    command: exec
    port: 8080
    scheme: http
    path: /
    httpHeaders:
      - name: foo  
        value: val
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    successThreshold: 10
    failureThreshold: 10
  livenessProbe:
    type: livenessProbe # should this be omitted since we already specify it one layer up?
    method: exec
    command: exec
    port: 8080
    scheme: http
    path: /
    httpHeaders:
      - name: foo  
        value: val
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 10
    successThreshold: 10
    failureThreshold: 10
  preStopHook: service_name_2_prestophook
`

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

func (ts *ServiceTestSuite) TestCreateServiceNoDefaultServiceSpecFailure() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, nil, nil, nil, nil)
	assert.NotNil(ts.T(), err)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}
	preStopHookCommand := "sleep 15"

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, &preStopHookCommand)
}

func (ts *ServiceTestSuite) TestCreateServiceNameTooLong() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Services
	ts.helper.name = "this-service-name-is-too-long-to-be-accepeted-fooooooooooooooooo"
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, nil, nil, nil, nil)

	assert.NotNil(ts.T(), err)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("default"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("rollingUpdate"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.Type("invalidStrategy"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("DeploymentStrategy succesfully created with invalid parameters"))
	}
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesDefault() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	readinessProbe := model.ServiceHealthProbeInput{Method: "default"}
	livenessProbe := model.ServiceHealthProbeInput{Method: "default"}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)
	portTwo := int32(8080)

	readinessProbe := model.ServiceHealthProbeInput{Method: "tcp", Port: &portOne}
	livenessProbe := model.ServiceHealthProbeInput{Method: "tcp", Port: &portTwo}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)

	readinessProbe := model.ServiceHealthProbeInput{Method: "tcp", Port: &portOne}
	livenessProbe := model.ServiceHealthProbeInput{Method: "tcp"}

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)
	scheme := "http"
	path := "/healthz"

	healthProbe := model.ServiceHealthProbeInput{
		Method: "http",
		Port:   &portOne,
		Scheme: &scheme,
		Path:   &path,
	}

	readinessProbe := healthProbe
	livenessProbe := healthProbe

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
}

func (ts *ServiceTestSuite) TestCreateServiceHealthProbesHTTPWIthHeaders() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

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

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	portOne := int32(9090)
	scheme := "invalid"
	path := "/healthz"

	healthProbe := model.ServiceHealthProbeInput{
		Method: "http",
		Port:   &portOne,
		Scheme: &scheme,
		Path:   &path,
	}

	readinessProbe := healthProbe
	livenessProbe := healthProbe

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	command := "./runcheck.sh"
	healthProbe := model.ServiceHealthProbeInput{
		Method:  "exec",
		Command: &command,
	}

	readinessProbe := healthProbe
	livenessProbe := healthProbe

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type:           plugins.GetType("rollingUpdate"),
		MaxUnavailable: 30,
		MaxSurge:       60,
	}

	healthProbe := model.ServiceHealthProbeInput{Method: "exec"}

	readinessProbe := healthProbe
	livenessProbe := healthProbe

	// Services
	_, err = ts.helper.CreateServiceWithError(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)
	if err == nil {
		assert.FailNow(ts.T(), fmt.Sprint("Health Probes successfully created with invalid parameters"))
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

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
	preStopHookCommand := "/bin/true"

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, &preStopHookCommand)

	preStopHookCommand = "/bin/change"

	// Update Service
	serviceID := string(serviceResolver.ID())

	servicePorts := []model.ServicePortInput{
		{
			Port:     80,
			Protocol: "HTTP",
		},
	}
	serviceInput := &model.ServiceInput{
		ID:        &serviceID,
		ProjectID: string(projectResolver.ID()),
		DeploymentStrategy: &model.DeploymentStrategyInput{
			Type:           plugins.GetType("rollingUpdate"),
			MaxUnavailable: 30,
			MaxSurge:       60,
		},
		Ports:       &servicePorts,
		PreStopHook: &preStopHookCommand,
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

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

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)

	// Update Service
	serviceID := string(serviceResolver.ID())
	projectID := string(projectResolver.ID())

	serviceInput := &model.ServiceInput{
		ID:        &serviceID,
		ProjectID: projectID,
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
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

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
	ts.helper.CreateServiceSpec(ts.T(), true)

	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	preHookCommand := "sleep 15"

	// Services
	serviceResolver := ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, &preHookCommand)

	// Test Service Interface
	_ = serviceResolver.ID()
	serviceProjectResolver := serviceResolver.Project()
	assert.Equal(ts.T(), projectResolver.ID(), serviceProjectResolver.ID())

	assert.Equal(ts.T(), "echo \"hello\" && exit 0", serviceResolver.Command())
	assert.Equal(ts.T(), "TestService", serviceResolver.Name())

	assert.Equal(ts.T(), int32(1), serviceResolver.Count())

	servicePorts, err := serviceResolver.Ports()
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), servicePorts, "Service Ports was empty")

	assert.Equal(ts.T(), "general", serviceResolver.Type())
	created_at_diff := time.Now().Sub(serviceResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	assert.Equal(ts.T(), preHookCommand, *serviceResolver.PreStopHook())

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
	ts.helper.CreateServiceSpec(ts.T(), true)

	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	livenessProbe := model.ServiceHealthProbeInput{}
	readinessProbe := model.ServiceHealthProbeInput{}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy,
		&readinessProbe, &livenessProbe, nil)

	// Test Service Query
	var ctx context.Context
	_, err = ts.Resolver.Services(ctx, emptyPaginatorInput)
	assert.NotNil(ts.T(), err)

	serviceResolvers, err := ts.Resolver.Services(test.ResolverAuthContext(), emptyPaginatorInput)
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), serviceResolvers)
	assert.NotEmpty(ts.T(), serviceResolvers, "Service Resolvers was empty")
}

func (ts *ServiceTestSuite) TestServiceImport_Success() {
	// pre-reqs
	envResolver := ts.helper.CreateEnvironment(ts.T())
	ts.helper.CreateServiceSpec(ts.T(), true)
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// validate input yaml string
	yamlServices := []model.ServiceInput{}
	err = yaml.Unmarshal([]byte(validYAMLServicesString), &yamlServices)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call service import
	serviceResolverMutation := graphql_resolver.ServiceResolverMutation{
		DB: ts.Resolver.DB,
	}

	serviceResolvers, err := serviceResolverMutation.ImportServices(&struct{ Services *model.ImportServicesInput }{
		Services: &model.ImportServicesInput{
			ProjectID:          projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:      envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			ServicesYAMLString: validYAMLServicesString,
		},
	})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// check outputs
	count := 0
	for _, inputService := range yamlServices {
		for _, createdServiceResolver := range serviceResolvers {
			if inputService.Name == createdServiceResolver.DBServiceResolver.Service.Name {
				count += 1
			}
		}
	}

	assert.Equal(ts.T(), count, len(yamlServices))

	// check side-effects
	createdServices := []model.Service{}
	err = ts.Resolver.DB.Where("environment_id = ? and project_id = ?",
		envResolver.DBEnvironmentResolver.Environment.Model.ID,
		projectResolver.DBProjectResolver.Project.Model.ID).Find(&createdServices).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), len(yamlServices), len(createdServices))
}

func (ts *ServiceTestSuite) TestServiceImport_Fail_InvalidProjectID() {
	return
}

func (ts *ServiceTestSuite) TestServiceImport_Fail_InvalidUserID() {
	return
}

func (ts *ServiceTestSuite) TestServiceImport_Fail_InvalidEnvironmentID() {
	return
}

func (ts *ServiceTestSuite) TestServiceImport_Fail_InvlidYAMLFileFormat() {
	return
}

func (ts *ServiceTestSuite) TestServiceImport_Fail_ServiceDoesNotExist() {
	return
}

func (ts *ServiceTestSuite) TestServiceImport_Fail_InvalidServiceType() {
	return
}

func (ts *ServiceTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteServiceResolver(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
