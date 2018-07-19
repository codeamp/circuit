package graphql_resolver_test

import (
	"context"
	"testing"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SecretTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
	helper   Helper
}

func (suite *SecretTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.helper.SetResolver(suite.Resolver, "TestSecret")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *SecretTestSuite) TestCreateSecret() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	ts.helper.CreateSecret(ts.T(), projectResolver)
}

func (ts *SecretTestSuite) TestUpdateSecretSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	secretResolver := ts.helper.CreateSecret(ts.T(), projectResolver)

	secretID := string(secretResolver.ID())
	secretInput := model.SecretInput{
		ID: &secretID,
	}

	_, err = ts.Resolver.UpdateSecret(test.ResolverAuthContext(), &struct{ Secret *model.SecretInput }{&secretInput})
	assert.Nil(ts.T(), err)
}

func (ts *SecretTestSuite) TestUpdateFailureNoAuth() {
	var ctx context.Context
	_, err := ts.Resolver.UpdateSecret(ctx, nil)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestUpdateFailureEnvMissingRecord() {
	secretID := test.ValidUUID
	secretInput := model.SecretInput{
		ID: &secretID,
	}

	_, err := ts.Resolver.UpdateSecret(test.ResolverAuthContext(), &struct{ Secret *model.SecretInput }{&secretInput})
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestDeleteSecretSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	secretResolver := ts.helper.CreateSecret(ts.T(), projectResolver)

	secretID := string(secretResolver.ID())
	secretInput := model.SecretInput{
		ID: &secretID,
	}

	_, err = ts.Resolver.DeleteSecret(test.ResolverAuthContext(), &struct{ Secret *model.SecretInput }{&secretInput})
	assert.Nil(ts.T(), err)
}

func (ts *SecretTestSuite) TestDeleteSecretFailure() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	ts.helper.CreateSecret(ts.T(), projectResolver)

	secretID := "123e4567-e89b-12d3-a456-426655440000"
	secretInput := model.SecretInput{
		ID: &secretID,
	}

	_, err = ts.Resolver.DeleteSecret(test.ResolverAuthContext(), &struct{ Secret *model.SecretInput }{&secretInput})
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretInterface() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	secretResolver := ts.helper.CreateSecret(ts.T(), projectResolver)

	_ = secretResolver.ID()
	_ = secretResolver.Key()
	_ = secretResolver.Value()
	_ = secretResolver.Scope()
	_ = secretResolver.Project()
	_ = secretResolver.User()
	_ = secretResolver.Type()
	_, _ = secretResolver.Versions()
	_ = secretResolver.Environment()
	_ = secretResolver.Created()
	_ = secretResolver.IsSecret()

	data, err := secretResolver.MarshalJSON()
	assert.Nil(ts.T(), err)

	err = secretResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretsQuery() {
	emptyPaginatorInput := &struct {
		Params *model.PaginatorInput
	}{nil}
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	ts.helper.CreateSecret(ts.T(), projectResolver)

	// Test with auth
	secretPaginator, err := ts.Resolver.Secrets(test.ResolverAuthContext(), emptyPaginatorInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	secretResolvers, err := secretPaginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), secretResolvers)

	// Test without auth
	var ctx context.Context
	secretPaginator, err = ts.Resolver.Secrets(ctx, emptyPaginatorInput)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretScopes() {
	testCases := []struct {
		input          string
		expectedResult model.SecretScope
	}{
		{"project", model.SecretScope("project")},
		{"failure-case", model.SecretScope("unknown")},
	}

	for _, testCase := range testCases {
		result := graphql_resolver.GetSecretScope(testCase.input)
		assert.Equal(ts.T(), testCase.expectedResult, result)
	}
}

func (ts *SecretTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteSecretResolver(t *testing.T) {
	suite.Run(t, new(SecretTestSuite))
}
