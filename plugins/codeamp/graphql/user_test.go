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

type UserTestSuite struct {
	suite.Suite
	UserResolver *graphql_resolver.UserResolver

	userIDs []string
}

func (suite *UserTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectEnvironment{},
		&model.ProjectBookmark{},
		&model.User{},
		&model.UserPermission{},
		&model.ProjectSettings{},
		&model.Environment{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.UserResolver = &graphql_resolver.UserResolver{UserResolver: &db_resolver.UserResolver{DB: db}}
}

/* Test successful project permissions update */
func (suite *UserTestSuite) TestCreateUser() {
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

func (suite *UserTestSuite) TearDownTest() {
	// suite.Resolver.DB.Delete(&model.Project{})
	// suite.Resolver.DB.Delete(&model.ProjectEnvironment{})
	// suite.Resolver.DB.Delete(&model.Environment{})
}

func TestSuiteUserResolver(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
