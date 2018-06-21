package db_resolver_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	Resolver *resolvers.Resolver
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
		&resolvers.Project{},
		&resolvers.Feature{},
	)
	suite.Resolver = &resolvers.Resolver{DB: db}
}

/* Test successful project permissions update */
func (suite *FeatureTestSuite) TestCreateFeature() {
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

	feature := resolvers.Feature{
		ProjectID:  project.Model.ID,
		Message:    "messagefoo",
		Hash:       "hashfoo",
		ParentHash: "parenthashfoo",
		Ref:        "reffoo",
		Created:    time.Now(),
	}

	suite.Resolver.DB.Create(&feature)

	authContext := context.WithValue(context.Background(), "jwt", resolvers.Claims{
		UserID:      "foo",
		Email:       "foo@gmail.com",
		Permissions: []string{"admin"},
	})

	featureList, err := suite.Resolver.Features(authContext, &struct {
		Params *resolvers.PaginatorInput
	}{})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), len(featureList.FeatureList), 1)

	suite.TearDownTest()
}

func (suite *FeatureTestSuite) TearDownTest() {
	suite.Resolver.DB.Delete(&resolvers.Project{})
	suite.Resolver.DB.Delete(&resolvers.Feature{})
}

func TestFeatureTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}
