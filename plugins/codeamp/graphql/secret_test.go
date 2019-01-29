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
	yaml "gopkg.in/yaml.v2"
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
	_, err = secretResolver.User()
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

func (ts *SecretTestSuite) TestSecretsImport_Success() {
	ts.T().Log("TestSecretsImport_Success")
	// provide inputs

	// YAML string of secrets
	secretsYAMLString := `
- key: SECRET_KEY
  value: "secret_value"
  type: "env"
  isSecret: false
- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	secrets := []model.YAMLSecret{}
	err := yaml.Unmarshal([]byte(secretsYAMLString), &secrets)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// user
	userResolver := ts.helper.CreateUser(ts.T())
	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())
	// project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call importer function
	ctx := context.Background()
	secretsResolver, err := ts.Resolver.ImportSecrets(ctx, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            userResolver.DBUserResolver.User.Model.ID.String(),
			ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			SecretsYAMLString: secretsYAMLString,
		},
	})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.NotNil(ts.T(), secretsResolver)

	// check that all new keys in yaml file were created
	count := 0
	for _, inputKey := range secretsResolver {
		for _, resolverKey := range secrets {
			if inputKey.Key() == resolverKey.Key {
				count += 1
			}
		}
	}

	assert.Equal(ts.T(), 2, count)

	// assert side effects
	var numDBSecretsCreated int
	ts.Resolver.DB.Where("project_id = ? and environment_id = ?", projectResolver.DBProjectResolver.Project.Model.ID, envResolver.DBEnvironmentResolver.Environment.Model.ID).Count(&numDBSecretsCreated)
	assert.Equal(ts.T(), 2, count)
}

func (ts *SecretTestSuite) TestSecretsImport_Fail_InvalidYAMLFileFormat() {
	ts.T().Log("TestSecretsImport_Fail_InvalidYAMLFileFormat")
	// provide inputs

	// invalid YAML string of secrets
	secretsYAMLString := `
- key: SECRET_KEY
  value: "secret_value"
  type: "env"
  isSecret: false
	- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	// user
	userResolver := ts.helper.CreateUser(ts.T())
	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())
	// project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call importer function
	ctx := context.Background()
	secretsResolver, err := ts.Resolver.ImportSecrets(ctx, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            userResolver.DBUserResolver.User.Model.ID.String(),
			ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			SecretsYAMLString: secretsYAMLString,
		},
	})
	assert.Nil(ts.T(), secretsResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretsImport_Fail_InvalidProjectID() {
	ts.T().Log("TestSecretsImport_Fail_InvalidProjectID")
	// provide inputs

	// YAML string of secrets
	secretsYAMLString := `
- key: SECRET_KEY
  value: "secret_value"
  type: "env"
  isSecret: false
- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	secrets := []model.YAMLSecret{}
	err := yaml.Unmarshal([]byte(secretsYAMLString), &secrets)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// user
	userResolver := ts.helper.CreateUser(ts.T())
	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// call importer function
	ctx := context.Background()
	secretsResolver, err := ts.Resolver.ImportSecrets(ctx, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            userResolver.DBUserResolver.User.Model.ID.String(),
			ProjectID:         "invalidprojectid",
			EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			SecretsYAMLString: secretsYAMLString,
		},
	})
	assert.Nil(ts.T(), secretsResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretsImport_Fail_InvalidUserID() {
	ts.T().Log("TestSecretsImport_Fail_InvalidUserID")
	// provide inputs

	// YAML string of secrets
	secretsYAMLString := `
- key: SECRET_KEY
  value: "secret_value"
  type: "env"
  isSecret: false
- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	secrets := []model.YAMLSecret{}
	err := yaml.Unmarshal([]byte(secretsYAMLString), &secrets)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())
	// project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call importer function
	ctx := context.Background()
	secretsResolver, err := ts.Resolver.ImportSecrets(ctx, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            "invalid-user-id",
			ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			SecretsYAMLString: secretsYAMLString,
		},
	})
	assert.Nil(ts.T(), secretsResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretsImport_Fail_InvalidEnvironmentID() {
	ts.T().Log("TestSecretsImport_Fail_InvalidEnvironmentID")
	// provide inputs

	// YAML string of secrets
	secretsYAMLString := `
- key: SECRET_KEY
  value: "secret_value"
  type: "env"
  isSecret: false
- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	secrets := []model.YAMLSecret{}
	err := yaml.Unmarshal([]byte(secretsYAMLString), &secrets)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// user
	userResolver := ts.helper.CreateUser(ts.T())
	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())
	// project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call importer function
	ctx := context.Background()
	secretsResolver, err := ts.Resolver.ImportSecrets(ctx, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            userResolver.DBUserResolver.User.Model.ID.String(),
			ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:     "invalid-env-id",
			SecretsYAMLString: secretsYAMLString,
		},
	})
	assert.Nil(ts.T(), secretsResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretsImport_Fail_InvalidSecretsType() {
	ts.T().Log("TestSecretsImport_Fail_InvalidSecretsType")
	// provide inputs

	// YAML string of secrets with invalid secrets type
	secretsYAMLString := `
- key: SECRET_KEY
  value: "secret_value"
  type: "invalid-type"
  isSecret: false
- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	secrets := []model.YAMLSecret{}
	err := yaml.Unmarshal([]byte(secretsYAMLString), &secrets)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// user
	userResolver := ts.helper.CreateUser(ts.T())
	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())
	// project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call importer function
	ctx := context.Background()
	secretsResolver, err := ts.Resolver.ImportSecrets(ctx, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            userResolver.DBUserResolver.User.Model.ID.String(),
			ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			SecretsYAMLString: secretsYAMLString,
		},
	})
	assert.Nil(ts.T(), secretsResolver)
	assert.NotNil(ts.T(), err)
}

func (ts *SecretTestSuite) TestSecretsImport_Success_ProtectedSecretCreated() {
	ts.T().Log("TestSecretsImport_Fail_InvalidSecretsType")
	// provide inputs

	// YAML string of secrets with invalid secrets type
	secretsYAMLString := `
- key: SECRET_KEY
  value: "protected_secret_value"
  type: "env"
  isSecret: true
- key: SECRET_KEY_2
  value: "secret_value_2"
  type: "build"
  isSecret: false
`
	secrets := []model.YAMLSecret{}
	err := yaml.Unmarshal([]byte(secretsYAMLString), &secrets)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// user
	userResolver := ts.helper.CreateUser(ts.T())
	// env
	envResolver := ts.helper.CreateEnvironment(ts.T())
	// project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// call importer function
	userContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      userResolver.DBUserResolver.Model.ID.String(),
		Email:       userResolver.DBUserResolver.User.Email,
		Permissions: []string{""},
	})
	secretsResolver, err := ts.Resolver.ImportSecrets(userContext, &struct{ Secrets *model.ImportSecretsInput }{
		Secrets: &model.ImportSecretsInput{
			UserID:            userResolver.DBUserResolver.User.Model.ID.String(),
			ProjectID:         projectResolver.DBProjectResolver.Project.Model.ID.String(),
			EnvironmentID:     envResolver.DBEnvironmentResolver.Environment.Model.ID.String(),
			SecretsYAMLString: secretsYAMLString,
		},
	})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), 2, len(secretsResolver))

	// check that protected was created
	page := int32(0)
	limit := int32(1)

	// just in case, we want to set the environment context before querying project secrets
	projectResolver.DBProjectResolver.Environment = envResolver.DBEnvironmentResolver.Environment
	projectSecretsResolver, err := projectResolver.Secrets(userContext, &struct {
		Params    *model.PaginatorInput
		SearchKey *string
	}{
		Params: &model.PaginatorInput{
			Page:  &page,
			Limit: &limit,
		},
		SearchKey: nil,
	})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	count, err := projectSecretsResolver.Count()
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.Equal(ts.T(), 2, int(count))

	secretsCreated, err := projectSecretsResolver.Entries()
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	protectedCount := 0
	for _, secret := range secretsCreated {
		if secret.IsSecret() {
			protectedCount += 1
			assert.Equal(ts.T(), "******", secret.Value())
		}
	}

	assert.Equal(ts.T(), 1, protectedCount)
}

func (ts *SecretTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteSecretResolver(t *testing.T) {
	suite.Run(t, new(SecretTestSuite))
}
