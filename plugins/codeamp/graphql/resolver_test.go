package graphql_resolver_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

type ResolverTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
}

func (suite *ResolverTestSuite) SetupTest() {
	// migrators := []interface{}{
	// 	&model.Extension{},
	// }

	// db, err := test.SetupResolverTest(migrators)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// _ = codeamp.CodeAmp{}
	// _ = &graphql_resolver.Resolver{DB: db, Events: nil, Redis: nil}

	// suite.Resolver = &graphql_resolver.Resolver{DB: db}
	// suite.ReleaseExtensionResolver = &graphql_resolver.ReleaseExtensionResolver{DBReleaseExtensionResolver: &db_resolver.ReleaseExtensionResolver{DB: db}}
}

func (ts *ResolverTestSuite) TearDownTest() {
}

func TestSuiteBaseResolver(t *testing.T) {
	auth.SetAuthEnabled(false)

	ts := new(ResolverTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
