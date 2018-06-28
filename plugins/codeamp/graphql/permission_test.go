package graphql_resolver_test

import (
	"fmt"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PermissionTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
	// PermissionResolver *graphql_resolver.ReleaseResolver

	userPermissionIDs []uuid.UUID
	permissionIDs     []uuid.UUID
	userIDs           []uuid.UUID
}

func (ts *PermissionTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	ts.Resolver = &graphql_resolver.Resolver{DB: db}
}

func (ts *PermissionTestSuite) TestPermissionInterface() {
	// permissionInput := model.PermissionInput{
	// 	Value: "TestPermissionInterface",
	// 	Grant: true,
	// }

	// userPermissionInput := model.UserPermissionsInput{
	// 	UserID: userID,
	// 	Permissions: []model.PermissionInput{
	// 		permissionInput,
	// 	},
	// }

	// userPermission := model.UserPermission{
	// 		UserID: uuid.FromStringOrNil(userId),
	// 		Value:  fmt.Sprintf("projects/%s", project.Repository),
	// }
	// r.DB.Create(&userPermission)

	// Create User
	user := model.User{
		Email:    fmt.Sprintf("test%d@example.com", time.Now().Unix()),
		Password: "TestPermissionInterface",
	}

	res := ts.Resolver.DB.Create(&user)
	if res.Error != nil {
		assert.FailNow(ts.T(), res.Error.Error())
	}
	ts.userIDs = append(ts.userIDs, user.Model.ID)
}

func (ts *PermissionTestSuite) TearDownTest() {
	for _, id := range ts.userPermissionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.UserPermission{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.userPermissionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.userIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.User{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.userIDs = make([]uuid.UUID, 0)
}

func TestSuitePermissionResolver(t *testing.T) {
	ts := new(PermissionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
