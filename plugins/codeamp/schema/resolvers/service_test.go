package resolvers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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

type TestService struct {
	suite.Suite
	db          *gorm.DB
	t           *transistor.Transistor
	actions     *actions.Actions
	user        models.User
	context     context.Context
	serviceSpec models.ServiceSpec
	project     models.Project
	env         models.Environment
}

func (suite *TestService) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"4500",
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

func (suite *TestService) SetupDBAndContext() {
	stamp := strings.ToLower("SetupDBAndContext")

	suite.db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.Service{},
		&models.Project{},
		&models.ContainerPort{},
		&models.Environment{},
		&models.ServiceSpec{},
	)

	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	suite.db.Save(&user)

	serviceSpec := models.ServiceSpec{
		Name:                   fmt.Sprintf("specname"),
		CpuRequest:             "500",
		CpuLimit:               "600",
		MemoryRequest:          "2",
		MemoryLimit:            "4",
		TerminationGracePeriod: "600",
		Created:                time.Now(),
	}
	suite.db.Save(&serviceSpec)

	project := models.Project{
		Name:          fmt.Sprintf("testname %s", time.Now().String()),
		Slug:          fmt.Sprintf("testslug %s", time.Now().String()),
		Repository:    "testrepository",
		Secret:        "testsecret",
		GitUrl:        "testgiturl",
		GitProtocol:   "testgitprotocol",
		RsaPrivateKey: "testrsaprivatekey",
		RsaPublicKey:  "testrsapublickey",
	}
	suite.db.Save(&project)

	env := models.Environment{
		Name: fmt.Sprintf("env%s", stamp),
	}
	suite.db.Save(&env)

	suite.context = context.WithValue(suite.context, "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
	suite.serviceSpec = serviceSpec
	suite.project = project
	suite.env = env
}

func (suite *TestService) TearDownSuite() {
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from user_permissions;")
	suite.db.Exec("delete from services;")
	suite.db.Exec("delete from service_specs;")
	suite.db.Exec("delete from projects;")
}

func (suite *TestService) TestSuccessfulCreateService() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulCreateService")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	serviceInput := struct {
		Service *resolvers.ServiceInput
	}{
		Service: &resolvers.ServiceInput{
			Name:          fmt.Sprintf("name%s", stamp),
			Command:       fmt.Sprintf("command%s", stamp),
			ServiceSpecId: suite.serviceSpec.Model.ID.String(),
			OneShot:       true,
			Count:         "1",
			ProjectId:     suite.project.Model.ID.String(),
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	createServiceResolver, _ := resolver.CreateService(&serviceInput)

	serviceInput2 := struct {
		Service *resolvers.ServiceInput
	}{
		Service: &resolvers.ServiceInput{
			Name:          fmt.Sprintf("name2%s", stamp),
			Command:       fmt.Sprintf("command2%s", stamp),
			ServiceSpecId: suite.serviceSpec.Model.ID.String(),
			OneShot:       true,
			Count:         "1",
			ProjectId:     suite.project.Model.ID.String(),
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	createServiceResolver2, _ := resolver.CreateService(&serviceInput2)

	serviceInput3 := struct {
		Service *resolvers.ServiceInput
	}{
		Service: &resolvers.ServiceInput{
			Name:          fmt.Sprintf("name3%s", stamp),
			Command:       fmt.Sprintf("command3%s", stamp),
			ServiceSpecId: suite.serviceSpec.Model.ID.String(),
			OneShot:       true,
			Count:         "1",
			ProjectId:     suite.project.Model.ID.String(),
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	createServiceResolver3, _ := resolver.CreateService(&serviceInput3)

	createServiceResolvers := []*resolvers.ServiceResolver{
		createServiceResolver3, createServiceResolver2, createServiceResolver,
	}

	serviceResolvers, _ := resolver.Services(suite.context)
	assert.Equal(suite.T(), 3, len(serviceResolvers))

	for idx, sResolver := range serviceResolvers {
		assert.Equal(suite.T(), createServiceResolvers[idx].Name(), sResolver.Name())
	}
	suite.TearDownSuite()
}

func (suite *TestService) TestSuccessfulUpdateService() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulUpdateService")

	service := models.Service{
		Name:          fmt.Sprintf("name2%s", stamp),
		Command:       fmt.Sprintf("command2%s", stamp),
		ServiceSpecId: suite.serviceSpec.Model.ID,
		OneShot:       true,
		Count:         "1",
		ProjectId:     suite.project.Model.ID,
	}
	suite.db.Save(&service)
	serviceId := service.Model.ID.String()

	serviceInput3 := struct {
		Service *resolvers.ServiceInput
	}{
		Service: &resolvers.ServiceInput{
			ID:            &serviceId,
			Name:          fmt.Sprintf("name3%s", stamp),
			Command:       fmt.Sprintf("command3%s", stamp),
			ServiceSpecId: suite.serviceSpec.Model.ID.String(),
			OneShot:       true,
			Count:         "1",
			ProjectId:     suite.project.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	updateServiceResolver3, _ := resolver.UpdateService(&serviceInput3)
	assert.Equal(suite.T(), fmt.Sprintf("name3%s", stamp), updateServiceResolver3.Name())

	// get services
	serviceResolvers, _ := resolver.Services(suite.context)
	assert.Equal(suite.T(), 1, len(serviceResolvers))
	assert.Equal(suite.T(), fmt.Sprintf("name3%s", stamp), serviceResolvers[0].Name())

	suite.TearDownSuite()
}

func (suite *TestService) TestSuccessfulDeleteService() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulDeleteService")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)

	// assert 0 services
	serviceResolvers, _ := resolver.Services(suite.context)
	assert.Equal(suite.T(), 0, len(serviceResolvers))

	service := models.Service{
		Name:          fmt.Sprintf("name2%s", stamp),
		Command:       fmt.Sprintf("command2%s", stamp),
		ServiceSpecId: suite.serviceSpec.Model.ID,
		OneShot:       true,
		Count:         "1",
		ProjectId:     suite.project.Model.ID,
		EnvironmentId: suite.env.Model.ID,
	}
	suite.db.Save(&service)
	serviceId := service.Model.ID.String()

	// assert 1
	serviceResolvers, _ = resolver.Services(suite.context)
	assert.Equal(suite.T(), 1, len(serviceResolvers))

	serviceInput3 := struct {
		Service *resolvers.ServiceInput
	}{
		Service: &resolvers.ServiceInput{
			ID:            &serviceId,
			Name:          fmt.Sprintf("name2%s", stamp),
			Command:       fmt.Sprintf("command2%s", stamp),
			ServiceSpecId: suite.serviceSpec.Model.ID.String(),
			OneShot:       true,
			Count:         "1",
			ProjectId:     suite.project.Model.ID.String(),
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	deleteServiceResolver, _ := resolver.DeleteService(&serviceInput3)
	assert.Equal(suite.T(), service.Model.ID.String(), string(deleteServiceResolver.ID()))

	serviceResolvers, _ = resolver.Services(suite.context)
	assert.Equal(suite.T(), 0, len(serviceResolvers))

	suite.TearDownSuite()
}

func TestServiceResolvers(t *testing.T) {
	suite.Run(t, new(TestService))
}
