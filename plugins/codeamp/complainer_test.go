package codeamp_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	"github.com/codeamp/circuit/plugins/codeamp/constants"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
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

func (ts *CodeampTestSuite) TestComplainerAlertSuccess() {
	constants.StagingEnvironment = "staging"
	constants.ProductionEnvironment = "production"

	prodEnv := model.Environment{
		Key:  "production",
		Name: "Production",
	}
	ts.Resolver.DB.Create(&prodEnv)

	stagingEnv := model.Environment{
		Key:  "staging",
		Name: "Staging",
	}
	ts.Resolver.DB.Create(&stagingEnv)

	p := model.Project{}
	ts.Resolver.DB.Create(&p)
	f := model.Feature{
		ProjectID: p.Model.ID,
	}
	ts.Resolver.DB.Create(&f)
	u := model.User{
		Email: "foo@foo.com",
	}
	ts.Resolver.DB.Create(&u)

	r := model.Release{
		EnvironmentID: prodEnv.Model.ID,
		HeadFeatureID: f.Model.ID,
		TailFeatureID: f.Model.ID,
		ProjectID:     p.Model.ID,
		State:         "complete",
		UserID:        u.Model.ID,
	}
	ts.Resolver.DB.Create(&r)

	complained, err := ts.CodeAmp.ComplainIfNotInStaging(&r, &p)

	assert.Nil(ts.T(), err)
	assert.True(ts.T(), complained)
}

func (ts *CodeampTestSuite) TestComplainerAlertFail_CompletedReleaseInStagingAlready() {
	constants.StagingEnvironment = "staging"
	constants.ProductionEnvironment = "production"

	prodEnv := model.Environment{
		Key:  "production",
		Name: "Production",
	}
	ts.Resolver.DB.Create(&prodEnv)

	stagingEnv := model.Environment{
		Key:  "staging",
		Name: "Staging",
	}

	ts.Resolver.DB.Create(&stagingEnv)

	p := model.Project{}
	ts.Resolver.DB.Create(&p)
	f := model.Feature{
		ProjectID: p.Model.ID,
	}
	ts.Resolver.DB.Create(&f)
	u := model.User{
		Email: "foo@foo.com",
	}
	ts.Resolver.DB.Create(&u)

	stagingRelease := model.Release{
		EnvironmentID: stagingEnv.Model.ID,
		HeadFeatureID: f.Model.ID,
		TailFeatureID: f.Model.ID,
		ProjectID:     p.Model.ID,
		State:         "complete",
		UserID:        u.Model.ID,
	}
	ts.Resolver.DB.Create(&stagingRelease)

	prodRelease := model.Release{
		EnvironmentID: prodEnv.Model.ID,
		HeadFeatureID: f.Model.ID,
		TailFeatureID: f.Model.ID,
		ProjectID:     p.Model.ID,
		State:         "complete",
		UserID:        u.Model.ID,
	}
	ts.Resolver.DB.Create(&prodRelease)
	complained, err := ts.CodeAmp.ComplainIfNotInStaging(&prodRelease, &p)

	assert.Nil(ts.T(), err)
	assert.False(ts.T(), complained)
}

func (ts *CodeampTestSuite) TearDownTest() {
	ts.Resolver.DB.Delete(model.Release{})
	ts.Resolver.DB.Delete(model.Environment{})
	ts.Resolver.DB.Delete(model.Feature{})
	ts.Resolver.DB.Delete(model.Project{})

	ts.Resolver.DB.Close()
}

/* Test successful env. creation */
func TestComplainerTestSuite(t *testing.T) {
	suite.Run(t, new(CodeampTestSuite))
}
