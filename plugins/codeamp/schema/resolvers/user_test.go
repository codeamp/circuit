package resolvers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestUser struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
}

func (suite *TestUser) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"15432",
		"postgres",
		"codeamp_test",
		"disable",
		"",
	))

	db.Exec(fmt.Sprintf("CREATE DATABASE %s", "codeamp_test"))
	db.Exec("CREATE EXTENSION \"uuid-ossp\"")
	db.Exec("CREATE EXTENSION IF NOT EXISTS hstore")

	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return codeamp.NewCodeAmp() })
	t, _ := transistor.NewTestTransistor(transistor.Config{
		Server:    "http://127.0.0.1:16379",
		Password:  "",
		Database:  "0",
		Namespace: "",
		Pool:      "30",
		Process:   "1",
		Plugins: map[string]interface{}{
			"codeamp": map[string]interface{}{
				"workers": 1,
				"postgres": map[string]interface{}{
					"host":     "0.0.0.0",
					"port":     "15432",
					"user":     "postgres",
					"dbname":   "codeamp_test",
					"sslmode":  "disable",
					"password": "",
				},
			},
		},
		EnabledPlugins: []string{},
		Queueing:       false,
	})

	actions := actions.NewActions(t.TestEvents, db)

	suite.db = db
	suite.t = t
	suite.actions = actions
}

func (suite *TestUser) SetupDBAndContext() {
	suite.db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
	)
}

func (suite *TestUser) TearDownSuite() {
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from user_permissions;")
}

/*
Test user permissions related resolvers:

- First, we must test retrieval of all distinct user permissions filtered by your privilege
- Second, we must test your ability to modify another user's permissions filtered by your privilege

By "filtered by your privilege", if you have sub-admin as your highest privilege, then you can only
modify privileges of that level or below to other users.

*/
func (suite *TestUser) TestGetSelfUserPermissions() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestGetSelfUserPermissions")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	suite.TearDownSuite()
}

func (suite *TestUser) TestUserPermissionsRetrieval() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestUserPermissionsRetrieval")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	suite.TearDownSuite()
}

func (suite *TestUser) TestUserPermissionsUpdateUser() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestUserPermissionsUpdateUser")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	suite.TearDownSuite()
}

func (suite *TestUser) TestSuccessfulCreateUser() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulCreateUser")

	// context = context.WithValue(suite.context, "jwt", utils.Claims{UserId: user.Model.ID.String()})

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	userInput := struct {
		User *resolvers.UserInput
	}{
		User: &resolvers.UserInput{
			Email:    fmt.Sprintf("foo%s@boo.com", stamp),
			Password: "secret",
		},
	}
	userResolver := resolver.CreateUser(&userInput)
	context1 := context.WithValue(context.TODO(), "jwt", utils.Claims{
		UserId: string(userResolver.ID()),
		Groups: []string{"admin"},
	})

	email, _ := userResolver.Email(context1)
	assert.Equal(suite.T(), fmt.Sprintf("foo%s@boo.com", stamp), email)

	userInput2 := struct {
		User *resolvers.UserInput
	}{
		User: &resolvers.UserInput{
			Email:    fmt.Sprintf("foo2%s@boo.com", stamp),
			Password: "secret",
		},
	}
	userResolver2 := resolver.CreateUser(&userInput2)
	context2 := context.WithValue(context.TODO(), "jwt", utils.Claims{
		UserId: string(userResolver2.ID()),
		Groups: []string{"admin"},
	})

	email2, _ := userResolver2.Email(context2)
	assert.Equal(suite.T(), fmt.Sprintf("foo2%s@boo.com", stamp), email2)

	userInput3 := struct {
		User *resolvers.UserInput
	}{
		User: &resolvers.UserInput{
			Email:    fmt.Sprintf("foo3%s@boo.com", stamp),
			Password: "secret",
		},
	}
	userResolver3 := resolver.CreateUser(&userInput3)
	context3 := context.WithValue(context.TODO(), "jwt", utils.Claims{
		UserId: string(userResolver3.ID()),
		Groups: []string{"admin"},
	})

	email3, _ := userResolver3.Email(context3)
	assert.Equal(suite.T(), fmt.Sprintf("foo3%s@boo.com", stamp), email3)

	createUserResolvers := []*resolvers.UserResolver{
		userResolver3, userResolver2, userResolver,
	}
	contexts := []context.Context{
		context3, context2, context1,
	}

	userResolvers, _ := resolver.Users(context1)

	assert.Equal(suite.T(), 3, len(userResolvers))
	for idx, userResolver := range userResolvers {
		expectedEmail, _ := createUserResolvers[idx].Email(contexts[idx])
		actualEmail, _ := userResolver.Email(contexts[idx])

		assert.Equal(suite.T(), expectedEmail, actualEmail)
	}

	suite.TearDownSuite()
}

func TestUserResolvers(t *testing.T) {
	suite.Run(t, new(TestUser))
}
