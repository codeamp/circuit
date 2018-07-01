package graphql_resolver_test

import (
	"testing"
	"time"

	log "github.com/codeamp/logger"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EnvironmentTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *EnvironmentTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Environment{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.helper.SetResolver(suite.Resolver, "TestEnvironment")
}

/* Test successful env. creation */
func (suite *EnvironmentTestSuite) TestCreateEnvironment() {
	// Environment
	suite.helper.CreateEnvironment(suite.T())
}

func (suite *EnvironmentTestSuite) TestEnvironmentInterface() {
	// Environment
	envResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver, _ := suite.helper.CreateProject(suite.T(), envResolver)

	assert.Equal(suite.T(), "TestEnvironment", envResolver.Name())
	assert.Equal(suite.T(), "TestEnvironment", envResolver.Key())
	assert.Equal(suite.T(), true, envResolver.IsDefault())
	assert.Equal(suite.T(), "color", envResolver.Color())

	js, err := envResolver.MarshalJSON()
	assert.Nil(suite.T(), err)

	err = envResolver.UnmarshalJSON(js)
	assert.Nil(suite.T(), err)

	created_at_diff := time.Now().Sub(envResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(suite.T(), "Created at time is too old")
	}

	projects := envResolver.Projects()
	assert.NotEmpty(suite.T(), projects, "Environment is missing associated projects")

	// Test Environments Query endpoint with a ProjectSlug
	projectSlug := string(projectResolver.Slug())
	_, err = suite.Resolver.Environments(test.ResolverAuthContext(), &struct{ ProjectSlug *string }{&projectSlug})
	assert.Nil(suite.T(), err)

	// Test without authorization
	_, err = suite.Resolver.Environments(nil, &struct{ ProjectSlug *string }{&projectSlug})
	assert.NotNil(suite.T(), err)

	// Test with an incorrect slug
	invalid_slug := "this-is-an-invalid-slug"
	_, err = suite.Resolver.Environments(test.ResolverAuthContext(), &struct{ ProjectSlug *string }{&invalid_slug})
	assert.NotNil(suite.T(), err)
}

func (suite *EnvironmentTestSuite) TestEnvironmentsQuery() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	envId := string(environmentResolver.ID())
	environmentResolvers, err := suite.Resolver.Environments(test.ResolverAuthContext(), &struct{ ProjectSlug *string }{nil})
	assert.Nil(suite.T(), err)

	foundNeedle := false
	for _, env := range environmentResolvers {
		if env.DBEnvironmentResolver.Environment.Model.ID.String() == envId {
			foundNeedle = true
			break
		}
	}

	assert.True(suite.T(), foundNeedle, "Was not able to find Environment in Environments table!")
}

/* Test successful env. update */
func (suite *EnvironmentTestSuite) TestUpdateEnvironment() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// update environment's name with same id
	envInput := model.EnvironmentInput{}
	envId := string(environmentResolver.ID())
	envInput.ID = &envId
	envInput.Color = "red"
	envInput.Name = "TestUpdateEnvironment"
	envInput.Key = "bar"

	// key SHOULD be ignored
	envInput.Key = "diffkey"
	// IsDefault SHOULD be ignored since it's the only default env
	envInput.IsDefault = false

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{&envInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.Equal(suite.T(), environmentResolver.ID(), updateEnvResolver.ID())
	assert.Equal(suite.T(), "TestUpdateEnvironment", updateEnvResolver.Name())
	assert.Equal(suite.T(), "red", updateEnvResolver.Color())
	assert.Equal(suite.T(), "TestEnvironment", updateEnvResolver.Key())

	assert.False(suite.T(), updateEnvResolver.IsDefault())
}

func (suite *EnvironmentTestSuite) TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse() {
	envResolvers := []*graphql_resolver.EnvironmentResolver{
		suite.helper.CreateEnvironmentWithName(suite.T(), "TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse"),
		suite.helper.CreateEnvironmentWithName(suite.T(), "TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse2"),
	}

	// 1
	assert.Equal(suite.T(), "TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse", envResolvers[0].Key())

	// 2
	assert.Equal(suite.T(), true, envResolvers[1].IsDefault())
	assert.Equal(suite.T(), "TestCreate2EnvsUpdateFirstEnvironmentIsDefaultToFalse2", envResolvers[1].Key())

	envInput := model.EnvironmentInput{}
	envInput.IsDefault = false
	envId := string(envResolvers[0].ID())
	envInput.ID = &envId

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{&envInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// Expecting this to be false since we just updated it above
	assert.Equal(suite.T(), false, updateEnvResolver.IsDefault())

	// IsDefault SHOULD be ignored since it's the only default env left
	envInput2 := model.EnvironmentInput{}
	envInput2.IsDefault = false
	envId = string(envResolvers[1].ID())
	envInput2.ID = &envId

	updateEnvResolver2, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{&envInput2})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.Equal(suite.T(), false, updateEnvResolver2.IsDefault())
}

func (suite *EnvironmentTestSuite) TearDownTest() {
	suite.helper.TearDownTest(suite.T())
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}
