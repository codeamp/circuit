package codeamp_test

import (
	"fmt"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

type CodeampTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
	CodeAmp  codeamp.CodeAmp
}

func (suite *CodeampTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.Environment{},
		&model.Feature{},
		&model.Release{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.CodeAmp = codeamp.CodeAmp{
		DB: db,
	}
	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
}

func (suite *CodeampTestSuite) TestComplainerAlert() {
	e := model.Environment{}
	suite.Resolver.DB.Create(&e)
	p := model.Project{}
	suite.Resolver.DB.Create(&p)
	f := model.Feature{
		ProjectID: p.Model.ID,
	}
	suite.Resolver.DB.Create(&f)

	r := model.Release{
		EnvironmentID: e.Model.ID,
		HeadFeatureID: f.Model.ID,
		TailFeatureID: f.Model.ID,
		ProjectID:     p.Model.ID,
		State:         "complete",
	}
	suite.Resolver.DB.Create(&r)

	suite.CodeAmp.ComplainIfNotInStaging(&r, &p)
}

func (suite *CodeampTestSuite) TearDownTest() {
	fmt.Println("TearDownTest")
	suite.Resolver.DB.Close()
}

/* Test successful env. creation */
func TestComplainerTestSuite(t *testing.T) {
	suite.Run(t, new(CodeampTestSuite))
}
