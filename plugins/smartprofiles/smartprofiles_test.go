package smartprofiles_test

import (
	"github.com/spf13/viper"
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
    influxdb:
      host: ""
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

func (suite *TestSuite) TestSmartProfilesSuccess() {
	project := plugins.Project{
		Slug: "codeamp-circuit",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	return
}

func (suite *TestSuite) TestSmartProfilesFailInvalidProjectInputs() {
	project := plugins.Project{
		Slug: "",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	assert.NotNil(suite.T(), err.Error())

	return
}


func (suite *TestSuite) TestSmartProfilesFailInvalidArtifactInputs() {
	project := plugins.Project{
		Slug: "checkr-checkr",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	assert.NotNil(suite.T(), err.Error())

	return
}

func (suite *TestSuite) TestSmartProfilesFailInfluxConnectionTimeout() {
	project := plugins.Project{
		Slug: "",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	assert.NotNil(suite.T(), err.Error())

	return
}


func (suite *TestSuite) TestSmartProfilesFailInfluxQueryFails() {
	project := plugins.Project{
		Slug: "",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	assert.NotNil(suite.T(), err.Error())

	return
}


func (suite *TestSuite) TestSmartProfilesFailInvalidOutputsRequestEmptyString() {
	project := plugins.Project{
		Slug: "",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	assert.NotNil(suite.T(), err.Error())

	return
}


func (suite *TestSuite) TestSmartProfilesFailInvalidOutputs() {
	project := plugins.Project{
		Slug: "",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
			},
		},
	}	

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), project)
	ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
	ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

	suite.transistor.Events <- ev	
	_, err := suite.transistor.GetTestEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), 100)
	assert.NotNil(suite.T(), err.Error())

	return
}

func TestSmartProfiles(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
