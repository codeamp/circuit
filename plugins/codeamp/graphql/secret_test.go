package graphql_resolver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SecretTestSuite struct {
	suite.Suite
	Resolver       *graphql_resolver.Resolver
	SecretResolver *graphql_resolver.SecretResolver

	cleanupSecretIDs      []interface{}
	cleanupEnvironmentIDs []interface{}
}

func (suite *SecretTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.SecretResolver = &graphql_resolver.SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: db}}
}

func (ts *SecretTestSuite) TestCreateSecret() {
	envInput := model.EnvironmentInput{
		Name:      "TestCreateSecret",
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
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, envResolver.ID())

	secretInput := model.SecretInput{
		Key:           "TestCreateSecret",
		Type:          "env",
		Scope:         "extension",
		EnvironmentID: string(envResolver.ID()),
		IsSecret:      false,
	}

	secretResolver, err := ts.Resolver.CreateSecret(test.ResolverAuthContext(), &struct {
		Secret *model.SecretInput
	}{Secret: &secretInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupSecretIDs = append(ts.cleanupSecretIDs, secretResolver.ID())

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

	envInput := model.EnvironmentInput{
		Name:      "TestSecretsQuery",
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
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, envResolver.ID())

	secretInput := model.SecretInput{
		Key:           "TestSecretsQuery",
		Type:          "env",
		Scope:         "extension",
		EnvironmentID: string(envResolver.ID()),
		IsSecret:      false,
	}

	secretResolver, err := ts.Resolver.CreateSecret(test.ResolverAuthContext(), &struct {
		Secret *model.SecretInput
	}{Secret: &secretInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupSecretIDs = append(ts.cleanupSecretIDs, secretResolver.ID())

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
	for _, id := range ts.cleanupSecretIDs {
		uid, _ := uuid.FromString(fmt.Sprintf("%v", id))
		err := ts.Resolver.DB.Unscoped().Delete(&model.Secret{Model: model.Model{ID: uid}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupSecretIDs = make([]interface{}, 0)

	for _, id := range ts.cleanupEnvironmentIDs {
		uid, _ := uuid.FromString(fmt.Sprintf("%v", id))
		err := ts.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: uid}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupEnvironmentIDs = make([]interface{}, 0)
}

func TestSuiteSecretResolver(t *testing.T) {
	ts := new(SecretTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
