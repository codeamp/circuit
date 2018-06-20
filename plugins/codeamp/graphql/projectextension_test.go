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

type ProjectExtensionTestSuite struct {
	suite.Suite
	Resolver                 *graphql_resolver.Resolver
	ProjectExtensionResolver *graphql_resolver.ProjectExtensionResolver
}

func (suite *ProjectExtensionTestSuite) SetupTest() {
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
	suite.ProjectExtensionResolver = &graphql_resolver.ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: db}}
}

func (ts *ProjectExtensionTestSuite) TearDownTest() {
}

func TestSuiteProjectExtensionResolver(t *testing.T) {
	auth.SetAuthEnabled(false)

	ts := new(ProjectExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
