package graphql_resolver_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

type ReleaseTestSuite struct {
	suite.Suite
	Resolver        *graphql_resolver.Resolver
	ReleaseResolver *graphql_resolver.ReleaseResolver
}

func (suite *ReleaseTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}
	_ = &graphql_resolver.Resolver{DB: db, Events: nil, Redis: nil}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.ReleaseResolver = &graphql_resolver.ReleaseResolver{DBReleaseResolver: &db_resolver.ReleaseResolver{DB: db}}
}

func (ts *ReleaseTestSuite) TearDownTest() {
}

func TestSuiteReleaseResolver(t *testing.T) {
	ts := new(ReleaseTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
