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

type ReleaseExtensionTestSuite struct {
	suite.Suite
	Resolver                 *graphql_resolver.Resolver
	ReleaseExtensionResolver *graphql_resolver.ReleaseExtensionResolver
}

func (suite *ReleaseExtensionTestSuite) SetupTest() {
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
	suite.ReleaseExtensionResolver = &graphql_resolver.ReleaseExtensionResolver{DBReleaseExtensionResolver: &db_resolver.ReleaseExtensionResolver{DB: db}}
}

func (ts *ReleaseExtensionTestSuite) TearDownTest() {
}

func TestSuiteReleaseExtensionResolver(t *testing.T) {
	ts := new(ReleaseExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
