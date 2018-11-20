package smartprofiles_test

import (
	"testing"

	_ "github.com/codeamp/circuit/plugins/smartprofiles"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
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
    influxdb:
      host: "http://10.0.1.143:8086"
      db: "telegraf"
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

// func (suite *TestSuite) TestSmartProfilesSuccess() {
// 	project := plugins.Project{
// 		Slug:        "checkr-deploy-test",
// 		Environment: "production",
// 		Services: []plugins.Service{
// 			plugins.Service{
// 				Name: "www",
// 			},
// 		},
// 	}
//
// 	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
// 	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
// 	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)
//
// 	suite.transistor.Events <- ev
// 	evt, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
// 	if err != nil {
// 		assert.FailNow(suite.T(), err.Error())
// 	}
//
// 	for _, svc := range evt.Payload.(plugins.Project).Services {
// 		spew.Dump(svc.ServiceSpec)
// 	}
//
// 	return
// }

func TestSmartProfiles(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
