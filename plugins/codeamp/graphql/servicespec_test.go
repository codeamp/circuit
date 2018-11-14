package graphql_resolver_test

import (
	"context"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceSpecTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *ServiceSpecTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
		&model.ServiceSpec{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.helper.SetResolver(suite.Resolver, "TestServiceSpec")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ServiceSpecTestSuite) TestCreateServiceSpecSuccess() {
	// Service Spec
	ts.helper.CreateServiceSpec(ts.T())
}

func (ts *ServiceSpecTestSuite) TestDeleteServiceSpecOnDefaultFailure() {
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	assert.Equal(ts.T(), false, serviceSpecResolver.IsDefault())
	
	serviceSpecResolverID := string(serviceSpecResolver.ID())

	// update 1st service spec with default true
	serviceSpecInput := model.ServiceSpecInput{
		ID: 					&serviceSpecResolverID,
		Name:                   serviceSpecResolver.Name(),
		CpuRequest:             serviceSpecResolver.CpuRequest(),
		CpuLimit:               serviceSpecResolver.CpuLimit(),
		MemoryRequest:          serviceSpecResolver.MemoryRequest(),
		MemoryLimit:            serviceSpecResolver.MemoryLimit(),
		TerminationGracePeriod: serviceSpecResolver.TerminationGracePeriod(),
		IsDefault: true,
	}	

	serviceSpecResolver, err := ts.helper.Resolver.UpdateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	assert.Equal(ts.T(), true, serviceSpecResolver.IsDefault())
	assert.Nil(ts.T(), err)

	_, err = ts.helper.Resolver.DeleteServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	assert.NotNil(ts.T(), err)	
}

func (ts *ServiceSpecTestSuite) TestCreateServiceSpecWithNewDefaultSuccess() {
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())
	serviceSpecResolver2 := ts.helper.CreateServiceSpec(ts.T())

	assert.Equal(ts.T(), false, serviceSpecResolver.IsDefault())
	
	serviceSpecResolverID := string(serviceSpecResolver.ID())
	serviceSpecResolver2ID := string(serviceSpecResolver2.ID())	

	// update 1st service spec with default true
	serviceSpecInput := model.ServiceSpecInput{
		ID: 					&serviceSpecResolverID,
		Name:                   serviceSpecResolver.Name(),
		CpuRequest:             serviceSpecResolver.CpuRequest(),
		CpuLimit:               serviceSpecResolver.CpuLimit(),
		MemoryRequest:          serviceSpecResolver.MemoryRequest(),
		MemoryLimit:            serviceSpecResolver.MemoryLimit(),
		TerminationGracePeriod: serviceSpecResolver.TerminationGracePeriod(),
		IsDefault: true,
	}	

	serviceSpecResolver, err := ts.helper.Resolver.UpdateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	assert.Equal(ts.T(), true, serviceSpecResolver.IsDefault())
	assert.Nil(ts.T(), err)
	
	// update 2nd service spec with default true
	serviceSpecInput = model.ServiceSpecInput{
		ID: 					&serviceSpecResolver2ID,
		Name:                   serviceSpecResolver2.Name(),
		CpuRequest:             serviceSpecResolver2.CpuRequest(),
		CpuLimit:               serviceSpecResolver2.CpuLimit(),
		MemoryRequest:          serviceSpecResolver2.MemoryRequest(),
		MemoryLimit:            serviceSpecResolver2.MemoryLimit(),
		TerminationGracePeriod: serviceSpecResolver2.TerminationGracePeriod(),
		IsDefault: true,
	}		

	serviceSpecResolver2, err = ts.helper.Resolver.UpdateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{ServiceSpec: &serviceSpecInput})
	
	// 1st service spec is now default = false
	firstSS := model.ServiceSpec{}
	ts.helper.Resolver.DB.Where("id = ?", serviceSpecResolverID).First(&firstSS)

	assert.Equal(ts.T(), false, firstSS.IsDefault)
	// 2nd service spec is now default = true
	assert.Equal(ts.T(), true, serviceSpecResolver2.IsDefault())
	assert.Nil(ts.T(), err)
}

func (ts *ServiceSpecTestSuite) TestUpdateServiceSpecSuccess() {
	// Service Spec
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Update Service Spec
	serviceSpecID := string(serviceSpecResolver.ID())
	serviceSpecInput := model.ServiceSpecInput{
		ID: &serviceSpecID,
	}
	_, err := ts.Resolver.UpdateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{&serviceSpecInput})
	assert.Nil(ts.T(), err)
}

func (ts *ServiceSpecTestSuite) TestDeleteServiceSpecSuccess() {
	// Service Spec
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Delete Service Spec
	serviceSpecID := string(serviceSpecResolver.ID())
	serviceSpecInput := model.ServiceSpecInput{
		ID: &serviceSpecID,
	}
	_, err := ts.Resolver.DeleteServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{&serviceSpecInput})
	assert.Nil(ts.T(), err)
}

func (ts *ServiceSpecTestSuite) TestDeleteServiceSpecFailureBadRecordID() {
	// Delete Service Spec
	serviceSpecID := test.ValidUUID
	serviceSpecInput := model.ServiceSpecInput{
		ID: &serviceSpecID,
	}
	_, err := ts.Resolver.DeleteServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{&serviceSpecInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ServiceSpecTestSuite) TestDeleteServiceSpecFailureHasDependencies() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	projectResolver.DBProjectResolver.Environment = environmentResolver.DBEnvironmentResolver.Environment

	// Service Spec
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, nil, nil, nil, nil)

	// Delete Service Spec
	serviceSpecID := test.ValidUUID
	serviceSpecInput := model.ServiceSpecInput{
		ID: &serviceSpecID,
	}
	_, err = ts.Resolver.DeleteServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{&serviceSpecInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ServiceSpecTestSuite) TestServiceSpecInterface() {
	// Service Spec
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	_ = serviceSpecResolver.ID()
	assert.Equal(ts.T(), "TestServiceSpec", serviceSpecResolver.Name())
	assert.Equal(ts.T(), "100", serviceSpecResolver.CpuRequest())
	assert.Equal(ts.T(), "200", serviceSpecResolver.CpuLimit())
	assert.Equal(ts.T(), "300", serviceSpecResolver.MemoryRequest())
	assert.Equal(ts.T(), "400", serviceSpecResolver.MemoryLimit())
	assert.Equal(ts.T(), "500", serviceSpecResolver.TerminationGracePeriod())
	assert.Equal(ts.T(), false, serviceSpecResolver.IsDefault())	

	data, err := serviceSpecResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = serviceSpecResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)

	created_at_diff := time.Now().Sub(serviceSpecResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	var ctx context.Context
	_, err = ts.Resolver.ServiceSpecs(ctx)
	assert.NotNil(ts.T(), err)

	serviceSpecResolvers, err := ts.Resolver.ServiceSpecs(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), serviceSpecResolvers)
}

func (ts *ServiceSpecTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteServiceSpecResolver(t *testing.T) {
	suite.Run(t, new(ServiceSpecTestSuite))
}
