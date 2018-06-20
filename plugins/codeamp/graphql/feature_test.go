package graphql_resolver_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	Resolver        *graphql_resolver.Resolver
	FeatureResolver *graphql_resolver.FeatureResolver
}

func (suite *FeatureTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Feature{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}
	_ = &graphql_resolver.Resolver{DB: db, Events: nil, Redis: nil}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.FeatureResolver = &graphql_resolver.FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: db}}
}

func (ts *FeatureTestSuite) TearDownTest() {
}

func TestSuiteFeatureResolver(t *testing.T) {
	auth.SetAuthEnabled(false)

	ts := new(FeatureTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
