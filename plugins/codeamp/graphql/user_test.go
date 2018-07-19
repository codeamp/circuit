package graphql_resolver_test

import (
	"context"
	"testing"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	graphql "github.com/graph-gophers/graphql-go"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
)

type UserTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *UserTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.User{},
		&model.UserPermission{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.helper.SetResolver(suite.Resolver, "TestUser")
}

/* Test successful project permissions update */
func (ts *UserTestSuite) TestCreateUserSuccess() {
	// User
	ts.helper.CreateUser(ts.T())
}

func (ts *UserTestSuite) TestQueryUser() {
	user := ts.helper.CreateUser(ts.T())
	userID := user.ID()

	userInput := struct {
		ID *graphql.ID
	}{&userID}

	// Test with auth and valid ID
	_, err := ts.Resolver.User(test.ResolverAuthContext(), &userInput)
	if err != nil {
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

func (ts *UserTestSuite) TestDeleteUserSuccess() {
	// User
	userResolver := ts.helper.CreateUser(ts.T())
	userID, err := uuid.FromString(string(userResolver.ID()))
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	user := model.User{
		Model: model.Model{ID: userID},
	}

	res := ts.Resolver.DB.Delete(&user)
	if res.Error != nil {
		assert.FailNow(ts.T(), res.Error.Error())
	}
}

func (ts *UserTestSuite) TestUsersQuery() {
	for x := 0; x < 5; x++ {
		ts.helper.CreateUser(ts.T())
	}

	res, err := ts.Resolver.Users(test.ResolverAuthContext())
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	assert.True(ts.T(), len(res) >= 5)
}

func (ts *UserTestSuite) TestUserInterface() {
	// Create User
	userResolver := ts.helper.CreateUser(ts.T())

	_ = userResolver.ID()
	assert.Equal(ts.T(), "test@example.com", userResolver.Email())

	// Test Permissions with context
	// TODO: This function needs auth!
	var context context.Context
	_ = userResolver.Permissions(context)

	// Test permissions without context
	_ = userResolver.Permissions(test.ResolverAuthContext())

	_ = userResolver.Created()
	data, err := userResolver.MarshalJSON()
	assert.Nil(ts.T(), err)

	err = userResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)
}

func (ts *UserTestSuite) TestUpdatePermissionsFailureNoAuth() {
	// Create User
	user := ts.helper.CreateUser(ts.T())

	// Create User Permission
	userPermissionsInput := model.UserPermissionsInput{
		UserID: string(user.ID()),
		Permissions: []model.PermissionInput{
			model.PermissionInput{
				"admin",
				true,
			},
		},
	}

	var ctx context.Context
	_, err := ts.Resolver.UpdateUserPermissions(ctx, &struct{ UserPermissions *model.UserPermissionsInput }{&userPermissionsInput})
	assert.NotNil(ts.T(), err)
}

func (ts *UserTestSuite) TestUpdatePermissionsFailureBadUser() {
	// Create User Permission
	userPermissionsInput := model.UserPermissionsInput{
		UserID: test.ValidUUID,
		Permissions: []model.PermissionInput{
			model.PermissionInput{
				"admin",
				true,
			},
		},
	}
	_, err := ts.Resolver.UpdateUserPermissions(test.ResolverAuthContext(), &struct{ UserPermissions *model.UserPermissionsInput }{&userPermissionsInput})
	assert.NotNil(ts.T(), err)
}

func (ts *UserTestSuite) TestUpdatePermissionsRemovePermission() {
	// Create User
	user := ts.helper.CreateUser(ts.T())

	// Create User Permission
	userPermissionsInput := model.UserPermissionsInput{
		UserID: string(user.ID()),
		Permissions: []model.PermissionInput{
			model.PermissionInput{
				"admin",
				true,
			},
		},
	}
	_, err := ts.Resolver.UpdateUserPermissions(test.ResolverAuthContext(), &struct{ UserPermissions *model.UserPermissionsInput }{&userPermissionsInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Create User Permission
	userPermissionsInput = model.UserPermissionsInput{
		UserID: string(user.ID()),
		Permissions: []model.PermissionInput{
			model.PermissionInput{
				"admin",
				false,
			},
		},
	}
	_, err = ts.Resolver.UpdateUserPermissions(test.ResolverAuthContext(), &struct{ UserPermissions *model.UserPermissionsInput }{&userPermissionsInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *UserTestSuite) TestPermissionInterface() {
	// Create User
	user := ts.helper.CreateUser(ts.T())

	// Create User Permission
	userPermissionsInput := model.UserPermissionsInput{
		UserID: string(user.ID()),
		Permissions: []model.PermissionInput{
			model.PermissionInput{
				"admin",
				true,
			},
		},
	}
	ts.Resolver.UpdateUserPermissions(test.ResolverAuthContext(), &struct{ UserPermissions *model.UserPermissionsInput }{&userPermissionsInput})

	var ctx context.Context
	_, err := ts.Resolver.Permissions(ctx)
	assert.NotNil(ts.T(), err)

	permissions, err := ts.Resolver.Permissions(test.ResolverAuthContext())
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), permissions)
	assert.NotEmpty(ts.T(), permissions)
}

func TearDownTest(ts *UserTestSuite) {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteUserResolver(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
