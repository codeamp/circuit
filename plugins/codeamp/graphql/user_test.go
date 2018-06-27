package graphql_resolver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	uuid "github.com/satori/go.uuid"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

type UserTestSuite struct {
	suite.Suite
	Resolver     *graphql_resolver.Resolver
	UserResolver *graphql_resolver.UserResolver

	cleanupUserIDs []uuid.UUID
	createdUserID  uuid.UUID
}

func (suite *UserTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.User{},
		&model.UserPermission{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}

	_ = codeamp.CodeAmp{}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.UserResolver = &graphql_resolver.UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: db}}
}

/* Test successful project permissions update */
func (ts *UserTestSuite) Test1GormCreateUser() {
	user := model.User{
		Email:    fmt.Sprintf("test%d@example.com", time.Now().Unix()),
		Password: "example",
	}

	res := ts.UserResolver.DBUserResolver.DB.Create(&user)
	if res.Error != nil {
		assert.FailNow(ts.T(), res.Error.Error())
	}

	ts.createdUserID = user.ID
	ts.T().Log("Created user with ID: ", user.ID)
}

func (ts *UserTestSuite) Test2GQLBGetUser() {
	var usr struct {
		ID *graphql.ID
	}
	graphqlID := graphql.ID(ts.createdUserID.String())
	usr.ID = &graphqlID

	// Test with auth and valid ID
	_, err := ts.Resolver.User(test.ResolverAuthContext(), &usr)
	if err != nil {
		ts.T().Log(ts.createdUserID.String())
		assert.FailNow(ts.T(), err.Error())
	}

	// Test with just auth
	_, err = ts.Resolver.User(test.ResolverAuthContext(), &struct{ ID *graphql.ID }{ID: nil})
	assert.NotNil(ts.T(), err)

	// Test with no auth
	// _, err = ts.Resolver.User(nil, &struct{ ID *graphql.ID }{ID: nil})
	// assert.NotNil(ts.T(), err)

	// Test with bad user id
	bad_gql_id := graphql.ID("11075553-5309-494B-9085-2D79A6ED1EB3")
	_, err = ts.Resolver.User(nil, &struct{ ID *graphql.ID }{ID: &bad_gql_id})
	assert.NotNil(ts.T(), err)
}

func (ts *UserTestSuite) Test3GormDeleteUser() {
	user := model.User{
		Model: model.Model{ID: ts.createdUserID},
	}

	res := ts.UserResolver.DBUserResolver.DB.Delete(&user)
	if res.Error != nil {
		assert.FailNow(ts.T(), res.Error.Error())
	}
}

func (ts *UserTestSuite) Test4GormCreate5Users() {
	for x := 0; x < 5; x++ {
		user := model.User{
			Email:    fmt.Sprintf("test%d@example.com", time.Now().Unix()),
			Password: "example",
		}

		res := ts.UserResolver.DBUserResolver.DB.Create(&user)
		if res.Error != nil {
			assert.FailNow(ts.T(), res.Error.Error())
		}

		ts.cleanupUserIDs = append(ts.cleanupUserIDs, user.ID)
	}
}

func (ts *UserTestSuite) Test5GQLBGet5Users() {
	var usr struct {
		ID *graphql.ID
	}
	graphqlID := graphql.ID(ts.createdUserID.String())
	usr.ID = &graphqlID

	res, err := ts.Resolver.Users(test.ResolverAuthContext())
	if err != nil {
		ts.T().Log(ts.createdUserID.String())
		assert.FailNow(ts.T(), err.Error())
	}

	assert.True(ts.T(), len(res) >= 5)
}

func (ts *UserTestSuite) TestGQLResolver() {
	user := model.User{
		Email:    fmt.Sprintf("test%d@example.com", time.Now().Unix()),
		Password: "example",
	}

	db := ts.UserResolver.DBUserResolver.DB.Create(&user)
	if db.Error != nil {
		assert.FailNow(ts.T(), db.Error.Error())
	}
	ts.cleanupUserIDs = append(ts.cleanupUserIDs, user.ID)

	gqr := &graphql_resolver.UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: db, User: user}}

	_ = gqr.ID()
	assert.Equal(ts.T(), user.Email, gqr.Email())

	// Test Permissions with context
	// TODO: This function needs auth!
	var context context.Context
	_ = gqr.Permissions(context)

	// Test permissions without context
	_ = gqr.Permissions(test.ResolverAuthContext())

	_ = gqr.Created()
	data, err := gqr.MarshalJSON()
	assert.Nil(ts.T(), err)

	err = gqr.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)
}

func TearDownTest(ts *UserTestSuite) {
	log.Warn("user tear down test")
	ts.UserResolver.DBUserResolver.DB.Unscoped().Delete(&model.User{Model: model.Model{ID: ts.createdUserID}})

	for _, i := range ts.cleanupUserIDs {
		ts.UserResolver.DBUserResolver.DB.Unscoped().Delete(&model.User{Model: model.Model{ID: i}})
	}
}

func TestSuiteUserResolver(t *testing.T) {
	ts := new(UserTestSuite)
	suite.Run(t, ts)
}
