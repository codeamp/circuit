package mongo_test

import (
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/suite"
)

type TestSuiteMongoExtension struct {
	suite.Suite
	transisdtor    *transistor.Transistor
	mongoInterface mongo.Mongoer
}
