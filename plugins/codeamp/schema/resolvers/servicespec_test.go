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

type TestServiceSpec struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
	user    models.User
	context context.Context
}

func (suite *TestServiceSpec) SetupSuite() {

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

func (suite *TestServiceSpec) SetupDBAndContext() {
	suite.db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.ServiceSpec{},
	)

	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	suite.db.Save(&user)

	suite.context = context.WithValue(suite.context, "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
}

func (suite *TestServiceSpec) TearDownSuite() {
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from user_permissions;")
	suite.db.Exec("delete from service_specs;")
}

/*

type ServiceSpec struct {
	Model                  `json:",inline"`
	Name                   string    `json:"name"`
	CpuRequest             string    `json:"cpuRequest"`
	CpuLimit               string    `json:"cpuLimit"`
	MemoryRequest          string    `json:"memoryRequest"`
	MemoryLimit            string    `json:"memoryLimit"`
	TerminationGracePeriod string    `json:"terminationGracePeriod"`
	Created                time.Time `json:"created"`
}
*/

func (suite *TestServiceSpec) TestSuccessfulCreateServiceSpec() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulCreateServiceSpec")

	// validate the table is empty
	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	serviceSpecResolvers, _ := resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 0, len(serviceSpecResolvers))

	serviceSpecInput := struct {
		ServiceSpec *resolvers.ServiceSpecInput
	}{
		ServiceSpec: &resolvers.ServiceSpecInput{
			Name:                   fmt.Sprintf("name%s", stamp),
			CpuRequest:             "500",
			CpuLimit:               "100",
			MemoryRequest:          "1",
			MemoryLimit:            "2",
			TerminationGracePeriod: "3600",
		},
	}

	createServiceSpecResolver, _ := resolver.CreateServiceSpec(&serviceSpecInput)
	assert.Equal(suite.T(), fmt.Sprintf("name%s", stamp), createServiceSpecResolver.Name(suite.context))

	serviceSpecInput2 := struct {
		ServiceSpec *resolvers.ServiceSpecInput
	}{
		ServiceSpec: &resolvers.ServiceSpecInput{
			Name:                   fmt.Sprintf("name%s", stamp),
			CpuRequest:             "500",
			CpuLimit:               "100",
			MemoryRequest:          "1",
			MemoryLimit:            "2",
			TerminationGracePeriod: "3600",
		},
	}

	createServiceSpecResolver2, _ := resolver.CreateServiceSpec(&serviceSpecInput2)
	assert.Equal(suite.T(), fmt.Sprintf("name%s", stamp), createServiceSpecResolver.Name(suite.context))

	serviceSpecInput3 := struct {
		ServiceSpec *resolvers.ServiceSpecInput
	}{
		ServiceSpec: &resolvers.ServiceSpecInput{
			Name:                   fmt.Sprintf("name%s", stamp),
			CpuRequest:             "500",
			CpuLimit:               "100",
			MemoryRequest:          "1",
			MemoryLimit:            "2",
			TerminationGracePeriod: "3600",
		},
	}

	createServiceSpecResolver3, _ := resolver.CreateServiceSpec(&serviceSpecInput3)
	assert.Equal(suite.T(), fmt.Sprintf("name%s", stamp), createServiceSpecResolver.Name(suite.context))

	createServiceSpecResolvers := []*resolvers.ServiceSpecResolver{
		createServiceSpecResolver3, createServiceSpecResolver2, createServiceSpecResolver,
	}

	// get services
	serviceSpecResolvers, _ = resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 3, len(serviceSpecResolvers))
	for idx, ssResolver := range serviceSpecResolvers {
		assert.Equal(suite.T(), createServiceSpecResolvers[idx].Name(suite.context), ssResolver.Name(suite.context))
	}

	suite.TearDownSuite()
}

func (suite *TestServiceSpec) TestSuccessfulUpdateServiceSpec() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulUpdateServiceSpec")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	serviceSpecResolvers, _ := resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 0, len(serviceSpecResolvers))

	serviceSpec := models.ServiceSpec{
		Name:                   fmt.Sprintf("name%s", stamp),
		CpuRequest:             "500",
		CpuLimit:               "100",
		MemoryRequest:          "1",
		MemoryLimit:            "2",
		TerminationGracePeriod: "3600",
	}
	suite.db.Save(&serviceSpec)

	serviceSpecResolvers, _ = resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 1, len(serviceSpecResolvers))
	assert.Equal(suite.T(), fmt.Sprintf("name%s", stamp), serviceSpecResolvers[0].Name(suite.context))

	serviceSpecId := serviceSpec.Model.ID.String()

	serviceSpecInput := struct {
		ServiceSpec *resolvers.ServiceSpecInput
	}{
		ServiceSpec: &resolvers.ServiceSpecInput{
			ID:                     &serviceSpecId,
			Name:                   fmt.Sprintf("name2%s", stamp),
			CpuRequest:             "500",
			CpuLimit:               "100",
			MemoryRequest:          "1",
			MemoryLimit:            "2",
			TerminationGracePeriod: "3600",
		},
	}

	updateServiceSpecResolver, _ := resolver.UpdateServiceSpec(&serviceSpecInput)
	assert.Equal(suite.T(), fmt.Sprintf("name2%s", stamp), updateServiceSpecResolver.Name(suite.context))

	serviceSpecResolvers, _ = resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 1, len(serviceSpecResolvers))
	assert.Equal(suite.T(), fmt.Sprintf("name2%s", stamp), serviceSpecResolvers[0].Name(suite.context))

	suite.TearDownSuite()
}

func (suite *TestServiceSpec) TestSuccessfulDeleteServiceSpec() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulDeleteServiceSpec")

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	serviceSpecResolvers, _ := resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 0, len(serviceSpecResolvers))

	serviceSpec := models.ServiceSpec{
		Name:                   fmt.Sprintf("name%s", stamp),
		CpuRequest:             "500",
		CpuLimit:               "100",
		MemoryRequest:          "1",
		MemoryLimit:            "2",
		TerminationGracePeriod: "3600",
	}
	suite.db.Save(&serviceSpec)

	serviceSpecResolvers, _ = resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 1, len(serviceSpecResolvers))
	assert.Equal(suite.T(), fmt.Sprintf("name%s", stamp), serviceSpecResolvers[0].Name(suite.context))

	serviceSpecId := serviceSpec.Model.ID.String()

	serviceSpecInput := struct {
		ServiceSpec *resolvers.ServiceSpecInput
	}{
		ServiceSpec: &resolvers.ServiceSpecInput{
			ID:                     &serviceSpecId,
			Name:                   fmt.Sprintf("name%s", stamp),
			CpuRequest:             "500",
			CpuLimit:               "100",
			MemoryRequest:          "1",
			MemoryLimit:            "2",
			TerminationGracePeriod: "3600",
		},
	}

	deleteServiceSpecResolver, _ := resolver.DeleteServiceSpec(&serviceSpecInput)
	assert.Equal(suite.T(), fmt.Sprintf("name%s", stamp), deleteServiceSpecResolver.Name(suite.context))

	serviceSpecResolvers, _ = resolver.ServiceSpecs(suite.context)
	assert.Equal(suite.T(), 0, len(serviceSpecResolvers))

	suite.TearDownSuite()
}

func TestServiceSpecResolvers(t *testing.T) {
	suite.Run(t, new(TestServiceSpec))
}
