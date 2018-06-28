package db_resolver_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	model "github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
}

func (suite *FeatureTestSuite) SetupTest() {
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
		&model.Project{},
		&model.Feature{},
	)
	suite.Resolver = &graphql_resolver.Resolver{DB: db}
}

/* Test successful project permissions update */
func (suite *FeatureTestSuite) TestCreateFeature() {
	// setup
	project := model.Project{
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

	feature := model.Feature{
		ProjectID:  project.Model.ID,
		Message:    "messagefoo",
		Hash:       "hashfoo",
		ParentHash: "parenthashfoo",
		Ref:        "reffoo",
		Created:    time.Now(),
	}

	suite.Resolver.DB.Create(&feature)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      "foo",
		Email:       "foo@gmail.com",
		Permissions: []string{"admin"},
	})

	featureList, err := suite.Resolver.Features(authContext, &struct {
		Params *model.PaginatorInput
	}{})
	if err != nil {
		log.Fatal(err.Error())
	}

	entries, err := featureList.Entries()
	if err != nil {
		log.Panic(err.Error())
	}

	assert.Equal(suite.T(), len(entries), 1)

	suite.TearDownTest()
}

func (suite *FeatureTestSuite) TearDownTest() {
	suite.Resolver.DB.Delete(&model.Project{})
	suite.Resolver.DB.Delete(&model.Feature{})
}

func TestFeatureTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}
