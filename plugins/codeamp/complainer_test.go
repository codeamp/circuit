package codeamp_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

type CodeampTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
	CodeAmp  codeamp.CodeAmp
}

func (suite *CodeampTestSuite) SetupTest() {
	suite.CodeAmp = codeamp.CodeAmp{}
}

func (suite *CodeampTestSuite) TestComplainerAlert() {
	suite.CodeAmp.ComplainIfNotInStaging(nil, nil)
}

/* Test successful env. creation */
func TestComplainerTestSuite(t *testing.T) {
	suite.Run(t, new(CodeampTestSuite))
}
