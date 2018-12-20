package smartprofiles_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins/smartprofiles"
	"github.com/davecgh/go-spew/spew"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/codeamp/circuit/plugins"

	// httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

func (suite *TestSuite) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  smartprofiles:
    workers: 1
`)	

	transistor.RegisterPlugin("smartprofiles", func() transistor.Plugin {
		return &smartprofiles.SmartProfiles{
			SmartProfilesClienter: &MockSmartProfilesClient{},
		}
	}, plugins.Project{})	

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

/* Tests:
  - length of inputted services = length of outputted services
  - recommended state is correct
*/
func (suite *TestSuite) TestSmartProfilesSuccess() {
	spew.Dump("TestSmartProfilesSuccess")
	project := plugins.Project{
		Slug:        "foobar",
		Environment: "development",
		Services: []plugins.Service{
			plugins.Service{
				Name: "www",
			},
		},
	}

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", "", false)
	ev.AddArtifact("INFLUX_DB", "", false)

	suite.transistor.Events <- ev
	evt, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	projectPayload := evt.Payload.(plugins.Project)
	for _, svc := range projectPayload.Services {
		assert.NotNil(suite.T(), svc.Spec)
	}

	assert.Equal(suite.T(), len(projectPayload.Services), 1)

	return
}

func (suite *TestSuite) TestComputeSampledQuerySuccess() {
	spew.Dump("TestComputeSampledQuerySuccess")
	spClienter := &smartprofiles.SmartProfiles{
		SmartProfilesClienter: &MockSmartProfilesClient{},
	}	

	_, err := smartprofiles.ComputeMeanSampledQuery(spClienter, "mean(memory_usage_bytes)/1000000 from kubernetes_pod_container where container = 'container' and namespace ='namespace'", 14)
	if err != nil {
		return
	}

	return
}


func TestSmartProfiles(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
