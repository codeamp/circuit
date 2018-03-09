package codeamp_resolvers_test

import (
	"fmt"
	"log"
	"testing"

	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	Resolver *resolvers.Resolver
}

func (suite *ProjectTestSuite) SetupTest() {
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
		&resolvers.Project{},
		&resolvers.ProjectPermission{},
		&resolvers.Environment{},
	)
	suite.Resolver = &resolvers.Resolver{DB: db}
}

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestUpdateProjectPermisions() {
	// setup
	project := resolvers.Project{
		Name:          "foo",
		Slug:          "foo",
		Repository:    "foo/foo",
		Secret:        "foo",
		GitUrl:        "foo",
		GitProtocol:   "foo",
		RsaPrivateKey: "foo",
		RsaPublicKey:  "foo",
	}
	suite.Resolver.DB.Create(&project)

	env := resolvers.Environment{
		Name:  "dev",
		Color: "purple",
		Key:   "dev",
	}
	suite.Resolver.DB.Create(&env)

	projectPermissionsInput := resolvers.ProjectPermissionsInput{
		ProjectID: project.Model.ID.String(),
		Permissions: []resolvers.ProjectPermissionInput{
			resolvers.ProjectPermissionInput{
				EnvironmentID: env.Model.ID.String(),
				Grant:         true,
			},
		},
	}

	updateProjectPermissionsResp, err := suite.Resolver.UpdateProjectPermissions(nil, &struct {
		ProjectPermissions *resolvers.ProjectPermissionsInput
	}{ProjectPermissions: &projectPermissionsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// check if env is found in response
	assert.Equal(suite.T(), 1, len(updateProjectPermissionsResp))
	assert.Equal(suite.T(), env.Model.ID.String(), updateProjectPermissionsResp[0])

	projectPermissions := []resolvers.ProjectPermission{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectPermissions)

	assert.Equal(suite.T(), 1, len(projectPermissions))
	assert.Equal(suite.T(), env.Model.ID.String(), projectPermissions[0].EnvironmentID.String())

	// take away access
	projectPermissionsInput.Permissions[0].Grant = false
	updateProjectPermissionsResp, err = suite.Resolver.UpdateProjectPermissions(nil, &struct {
		ProjectPermissions *resolvers.ProjectPermissionsInput
	}{ProjectPermissions: &projectPermissionsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 0, len(updateProjectPermissionsResp))

	projectPermissions = []resolvers.ProjectPermission{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectPermissions)

	assert.Equal(suite.T(), 0, len(projectPermissions))

	suite.TearDownTest()
}

func (suite *ProjectTestSuite) TearDownTest() {
	suite.Resolver.DB.Delete(&resolvers.Project{})
	suite.Resolver.DB.Delete(&resolvers.ProjectPermission{})
	suite.Resolver.DB.Delete(&resolvers.Environment{})
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
