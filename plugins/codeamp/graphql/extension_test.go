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

type ExtensionTestSuite struct {
	suite.Suite
	Resolver          *graphql_resolver.Resolver
	ExtensionResolver *graphql_resolver.ExtensionResolver
}

func (suite *ExtensionTestSuite) SetupTest() {
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
	suite.ExtensionResolver = &graphql_resolver.ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: db}}
}

func (ts *ExtensionTestSuite) TearDownTest() {
}

func TestSuiteExtensionResolver(t *testing.T) {
	auth.SetAuthEnabled(false)

	ts := new(ExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
