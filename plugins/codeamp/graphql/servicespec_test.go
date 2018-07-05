package graphql_resolver_test

import (
	"context"
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

type ServiceSpecTestSuite struct {
	suite.Suite
	Resolver            *graphql_resolver.Resolver
	ServiceSpecResolver *graphql_resolver.ServiceSpecResolver

	cleanupServiceSpecIDs []uuid.UUID
}

func (suite *ServiceSpecTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.ServiceSpecResolver = &graphql_resolver.ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: db}}
}

func (ts *ServiceSpecTestSuite) TestServiceSpecInterface() {
	// Service Spec ID
	serviceSpecInput := model.ServiceSpecInput{
		Name:                   "TestServiceSpecInterface",
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	serviceSpecResolver, err := ts.Resolver.CreateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{&serviceSpecInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupServiceSpecIDs = append(ts.cleanupServiceSpecIDs, serviceSpecResolver.DBServiceSpecResolver.ServiceSpec.Model.ID)

	_ = serviceSpecResolver.ID()
	assert.Equal(ts.T(), serviceSpecInput.Name, serviceSpecResolver.Name())
	assert.Equal(ts.T(), serviceSpecInput.CpuRequest, serviceSpecResolver.CpuRequest())
	assert.Equal(ts.T(), serviceSpecInput.CpuLimit, serviceSpecResolver.CpuLimit())
	assert.Equal(ts.T(), serviceSpecInput.MemoryRequest, serviceSpecResolver.MemoryRequest())
	assert.Equal(ts.T(), serviceSpecInput.MemoryLimit, serviceSpecResolver.MemoryLimit())
	assert.Equal(ts.T(), serviceSpecInput.TerminationGracePeriod, serviceSpecResolver.TerminationGracePeriod())

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
	for _, id := range ts.cleanupServiceSpecIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.ServiceSpec{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupServiceSpecIDs = make([]uuid.UUID, 0)
}

func TestSuiteServiceSpecResolver(t *testing.T) {
	ts := new(ServiceSpecTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
