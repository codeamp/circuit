package graphql_resolver_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

type ProjectTestSuite struct {
	suite.Suite
	UserResolver *graphql_resolver.UserResolver

	userIDs []gorm.ID
}

func (suite *ProjectTestSuite) SetupTest() {
	db, err := test.SetupResolverTest()
	if err != nil {
		log.Fatal(err.Error())
	}

	// TODO : ADB : Find a better way of passing this to the function from the
	// individual resolver test
	db.AutoMigrate(
		&graphql_resolver.Project{},
		&graphql_resolver.ProjectEnvironment{},
		&graphql_resolver.ProjectBookmark{},
		&model.UserPermission{},
		&graphql_resolver.ProjectSettings{},
		&graphql_resolver.Environment{},
	)

	suite.UserResolver = &graphql_resolver.UserResolver{UserResolver: &db_resolver.UserResolver{DB: db}}
}

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestCreateUser() {
	user := model.User{
		Email:    "test@example.com",
		Password: "example",
	}

	res := suite.UserResolver.UserResolver.DB.Create(&user)
	if res.Error != nil {
		log.Error(res.Error)
		assert.Nil(suite.T(), res.Error)
	}

	//suite.TearDownTest(deleteIds)
}

func (suite *ProjectTestSuite) TearDownTest() {
	// suite.Resolver.DB.Delete(&graphql_resolver.Project{})
	// suite.Resolver.DB.Delete(&graphql_resolver.ProjectEnvironment{})
	// suite.Resolver.DB.Delete(&graphql_resolver.Environment{})
}

func TestSuiteUserResolver(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
