package graphql_resolver_test

import (
	"log"
	"testing"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	graphql "github.com/graph-gophers/graphql-go"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EnvironmentTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
}

func (suite *EnvironmentTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Environment{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	// clean, just in case
	db.Delete(&model.Environment{})
	suite.Resolver = &graphql_resolver.Resolver{DB: db}
}

/* Test successful env. creation */
func (suite *EnvironmentTestSuite) TestCreateEnvironment() {
	envInput := model.EnvironmentInput{
		Name:      "test",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}
	assert.Equal(suite.T(), envResolver.Name(), "test")
	assert.Equal(suite.T(), envResolver.Key(), "foo")
	assert.Equal(suite.T(), envResolver.IsDefault(), true)
	assert.Equal(suite.T(), envResolver.Color(), "color")
	assert.NotEqual(suite.T(), envResolver.Color(), "wrongcolor")

	suite.TearDownTest([]uuid.UUID{envResolver.DBEnvironmentResolver.Environment.Model.ID})
}

/* Test successful env. update */
func (suite *EnvironmentTestSuite) TestUpdateEnvironment() {
	envInput := model.EnvironmentInput{
		Name:      "test",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// update environment's name with same id
	envId := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	envInput.ID = &envId
	envInput.Color = "red"
	envInput.Name = "test2"
	// key SHOULD be ignored
	envInput.Key = "diffkey"
	// IsDefault SHOULD be ignored since it's the only default env
	envInput.IsDefault = false

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver.ID(), graphql.ID(envResolver.DBEnvironmentResolver.Environment.Model.ID.String()))
	assert.Equal(suite.T(), updateEnvResolver.Name(), "test2")
	assert.Equal(suite.T(), updateEnvResolver.Color(), "red")
	assert.Equal(suite.T(), updateEnvResolver.Key(), "foo")
	assert.Equal(suite.T(), updateEnvResolver.IsDefault(), true)
	assert.NotEqual(suite.T(), updateEnvResolver.Name(), "diffkey")

	suite.TearDownTest([]uuid.UUID{updateEnvResolver.DBEnvironmentResolver.Environment.Model.ID})
}

func (suite *EnvironmentTestSuite) TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse() {
	envInput := model.EnvironmentInput{
		Name:      "test",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), envResolver.Key(), "foo")

	envInput2 := model.EnvironmentInput{
		Name:      "test",
		Key:       "foo2",
		IsDefault: true,
		Color:     "color",
	}

	envResolver2, err := suite.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput2})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), envResolver2.IsDefault(), true)
	assert.Equal(suite.T(), envResolver2.Key(), "foo2")

	envInput.IsDefault = false
	envId := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	envInput.ID = &envId

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver.IsDefault(), false)

	// IsDefault SHOULD be ignored since it's the only default env left
	envInput2.IsDefault = false
	envId = envResolver2.DBEnvironmentResolver.Environment.Model.ID.String()
	envInput2.ID = &envId

	updateEnvResolver2, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput2})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver2.IsDefault(), true)

	suite.TearDownTest([]uuid.UUID{envResolver.DBEnvironmentResolver.Environment.Model.ID, envResolver2.DBEnvironmentResolver.Environment.Model.ID})
}

func (suite *EnvironmentTestSuite) TearDownTest(ids []uuid.UUID) {
	for _, id := range ids {
		suite.Resolver.DB.Where("id = ?", id).Delete(&model.Environment{})
	}
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}
