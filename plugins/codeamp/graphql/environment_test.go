package graphql_resolver_test

import (
	"testing"
	"time"

	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	// See http://golang.org/pkg/time/#Parse
	timeFormat = "2006-01-02 15:04:05.999999999 -0700 MST m"
	// 2018-06-25 13:04:51.973461415 -0700 PDT m=+0.485802342
	// 			  2018-06-25T12:57:40-07:00
)

type EnvironmentTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	cleanupEnvironmentIDs []uuid.UUID
	cleanupProjectIDs     []uuid.UUID
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
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

	assert.Equal(suite.T(), envResolver.Name(), "test")
	assert.Equal(suite.T(), envResolver.Key(), "foo")
	assert.Equal(suite.T(), envResolver.IsDefault(), true)
	assert.Equal(suite.T(), envResolver.Color(), "color")
	assert.NotEqual(suite.T(), envResolver.Color(), "wrongcolor")

	js, err := envResolver.MarshalJSON()
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	unmarshalledEnvResolver := &graphql_resolver.EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{Environment: model.Environment{}}}
	err = unmarshalledEnvResolver.UnmarshalJSON(js)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
}

func (suite *EnvironmentTestSuite) TestCreateEnvironmentAndProject() {
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
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

	assert.Equal(suite.T(), envResolver.Name(), "test")
	assert.Equal(suite.T(), envResolver.Key(), "foo")
	assert.Equal(suite.T(), envResolver.IsDefault(), true)
	assert.Equal(suite.T(), envResolver.Color(), "color")
	assert.NotEqual(suite.T(), envResolver.Color(), "wrongcolor")

	environmentID := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	projectInput := model.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/bar.git",
		EnvironmentID: &environmentID,
	}

	createProjectResolver, err := suite.Resolver.CreateProject(test.ResolverAuthContext(), &struct {
		Project *model.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	suite.cleanupProjectIDs = append(suite.cleanupProjectIDs, createProjectResolver.DBProjectResolver.Project.ID)

	created_at_diff := time.Now().Sub(envResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(suite.T(), "Created at time is invalid")
	}

	projects := envResolver.Projects()
	assert.NotEqual(suite.T(), len(projects), "Environment is missing associated projects")
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
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

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
		assert.FailNow(suite.T(), err.Error())
	}

	assert.Equal(suite.T(), updateEnvResolver.ID(), graphql.ID(envResolver.DBEnvironmentResolver.Environment.Model.ID.String()))
	assert.Equal(suite.T(), updateEnvResolver.Name(), "test2")
	assert.Equal(suite.T(), updateEnvResolver.Color(), "red")
	assert.Equal(suite.T(), updateEnvResolver.Key(), "foo")

	assert.Equal(suite.T(), false, updateEnvResolver.IsDefault())
	assert.NotEqual(suite.T(), updateEnvResolver.Name(), "diffkey")
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
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

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
		assert.FailNow(suite.T(), err.Error())
	}
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver2.DBEnvironmentResolver.Environment.Model.ID)

	assert.Equal(suite.T(), envResolver2.IsDefault(), true)
	assert.Equal(suite.T(), envResolver2.Key(), "foo2")

	envInput.IsDefault = false
	envId := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	envInput.ID = &envId

	updateEnvResolver, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// Expecting this to be false since we just updated it above
	assert.Equal(suite.T(), false, updateEnvResolver.IsDefault())

	// IsDefault SHOULD be ignored since it's the only default env left
	envInput2.IsDefault = false
	envId = envResolver2.DBEnvironmentResolver.Environment.Model.ID.String()
	envInput2.ID = &envId

	updateEnvResolver2, err := suite.Resolver.UpdateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput2})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// There are issues with this test on circle
	// Temporarily Disabling
	// ADB
	// Expecting this to be false since we just updated it above
	assert.Equal(suite.T(), false, updateEnvResolver2.IsDefault())

}

func (suite *EnvironmentTestSuite) TestEnvironmentsQuery() {
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
		assert.FailNow(suite.T(), err.Error())
	}
	envId := envResolver.DBEnvironmentResolver.Environment.Model.ID.String()
	suite.cleanupEnvironmentIDs = append(suite.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)

	environmentResolvers, err := suite.Resolver.Environments(test.ResolverAuthContext(), &struct{ ProjectSlug *string }{ProjectSlug: nil})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	foundNeedle := false
	for _, env := range environmentResolvers {
		if env.DBEnvironmentResolver.Environment.Model.ID.String() == envId {
			foundNeedle = true
			break
		}
	}

	if foundNeedle == false {
		assert.FailNow(suite.T(), "Was not able to find Environment in Environments table!")
	}
}

func (suite *EnvironmentTestSuite) TearDownTest() {
	for _, id := range suite.cleanupProjectIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Project{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
	suite.cleanupProjectIDs = make([]uuid.UUID, 0)

	for _, id := range suite.cleanupEnvironmentIDs {
		err := suite.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
	suite.cleanupEnvironmentIDs = make([]uuid.UUID, 0)
}

func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}
