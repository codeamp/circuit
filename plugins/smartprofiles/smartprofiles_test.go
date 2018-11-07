package smartprofiles_test

import (
	"github.com/davecgh/go-spew/spew"
	"testing"

	"github.com/codeamp/circuit/plugins"
	_ "github.com/codeamp/circuit/plugins/smartprofiles"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	// httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  smartprofiles:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestSmartProfilesNotifySuccessfulRecommendations() {
	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), plugins.Project{})

	spew.Dump("TESTING SMART PROFILES")
	suite.transistor.Events <- ev	
	spew.Dump("SENT EVENT")
	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	spew.Dump("GOT EVENT")

	spew.Dump(e)

	return
}

func TestSmartProfiles(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
