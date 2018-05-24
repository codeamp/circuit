package codeamp_resolvers_test

import (
	"fmt"
	"log"
	"testing"

	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EnvironmentTestSuite struct {
	suite.Suite
	Resolver *resolvers.Resolver
}

func (suite *EnvironmentTestSuite) SetupTest() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("../../../configs/circuit.test.yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		log.Fatal(err.Error())
	}
	db.AutoMigrate(
		&resolvers.Environment{},
	)
	// clean, just in case
	db.Delete(&resolvers.Environment{})
	suite.Resolver = &resolvers.Resolver{DB: db}
}

/* Test successful env. creation */
func (suite *EnvironmentTestSuite) TestCreateEnvironment() {
	envInput := resolvers.EnvironmentInput{
		Name:      "test",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}
	assert.Equal(suite.T(), envResolver.Name(), "test")
	assert.Equal(suite.T(), envResolver.Key(), "foo")
	assert.Equal(suite.T(), envResolver.IsDefault(), true)
	assert.Equal(suite.T(), envResolver.Color(), "color")
	assert.NotEqual(suite.T(), envResolver.Color(), "wrongcolor")

	suite.TearDownTest([]string{envResolver.Environment.Model.ID.String()})
}

/* Test successful env. update */
func (suite *EnvironmentTestSuite) TestUpdateEnvironment() {
	envInput := resolvers.EnvironmentInput{
		Name:      "test",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// update environment's name with same id
	envId := envResolver.Environment.Model.ID.String()
	envInput.ID = &envId
	envInput.Color = "red"
	envInput.Name = "test2"
	// key SHOULD be ignored
	envInput.Key = "diffkey"
	// IsDefault SHOULD be ignored since it's the only default env
	envInput.IsDefault = false

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver.ID(), graphql.ID(envResolver.Environment.Model.ID.String()))
	assert.Equal(suite.T(), updateEnvResolver.Name(), "test2")
	assert.Equal(suite.T(), updateEnvResolver.Color(), "red")
	assert.Equal(suite.T(), updateEnvResolver.Key(), "foo")
	assert.Equal(suite.T(), updateEnvResolver.IsDefault(), true)
	assert.NotEqual(suite.T(), updateEnvResolver.Name(), "diffkey")

	suite.TearDownTest([]string{updateEnvResolver.Environment.Model.ID.String()})
}

func (suite *EnvironmentTestSuite) TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse() {
	envInput := resolvers.EnvironmentInput{
		Name:      "test",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := suite.Resolver.CreateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), envResolver.Key(), "foo")

	envInput2 := resolvers.EnvironmentInput{
		Name:      "test",
		Key:       "foo2",
		IsDefault: true,
		Color:     "color",
	}

	envResolver2, err := suite.Resolver.CreateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput2})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), envResolver2.IsDefault(), true)
	assert.Equal(suite.T(), envResolver2.Key(), "foo2")

	envInput.IsDefault = false
	envId := envResolver.Environment.Model.ID.String()
	envInput.ID = &envId

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver.IsDefault(), false)

	// IsDefault SHOULD be ignored since it's the only default env left
	envInput2.IsDefault = false
	envId = envResolver2.Environment.Model.ID.String()
	envInput2.ID = &envId

	updateEnvResolver2, err := suite.Resolver.UpdateEnvironment(nil, &struct{ Environment *resolvers.EnvironmentInput }{Environment: &envInput2})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver2.IsDefault(), true)

	suite.TearDownTest([]string{envResolver.Environment.Model.ID.String(), envResolver2.Environment.Model.ID.String()})
}

func (suite *EnvironmentTestSuite) TearDownTest(ids []string) {
	for _, id := range ids {
		suite.Resolver.DB.Where("id = ?", id).Delete(&resolvers.Environment{})
	}
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}
