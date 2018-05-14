package k8s_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/k8s"
	"github.com/codeamp/circuit/plugins/k8s/testdata"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  k8s:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("YAML")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CODEAMP")
	viper.AutomaticEnv()
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	transistor.RegisterPlugin("k8s", func() transistor.Plugin {
		return &k8s.K8s{}
	})

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"k8s"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

// func (suite *TestSuite) TestCleanupLBOffice() {
// 	suite.transistor.Events <- testdata.LBTCPEvent(plugins.GetAction("destroy"), plugins.GetType("office"))

// 	e := suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 60)
// 	assert.Equal(suite.T(), plugins.GetState("deleted"), e.Payload.(plugins.ProjectExtension).State, e.Payload.(plugins.ProjectExtension).StateMessage)
// }

// Load Balancers Tests
func (suite *TestSuite) TestLBTCPOffice() {
	timer := time.NewTimer(time.Second * 100)
	defer timer.Stop()

	go func() {
		<-timer.C
		log.Fatal("TestLBTCPOffice: Test timeout")
	}()

	suite.transistor.Events <- testdata.LBTCPEvent(plugins.GetAction("update"), plugins.GetType("office"))

	e := suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 120)
	assert.Equal(suite.T(), plugins.GetState("complete"), e.Payload.(plugins.ProjectExtension).State, e.Payload.(plugins.ProjectExtension).StateMessage)
	if e.Payload.(plugins.ProjectExtension).State != plugins.GetState("complete") {
		return
	}

	testdata.LBTCPEvent(plugins.GetAction("update"), plugins.GetType("office"))

	for {
		e = suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 120)
		if e.Payload.(plugins.ReleaseExtension).State != "running" {
			break
		}
	}

	suite.transistor.Events <- testdata.LBTCPEvent(plugins.GetAction("destroy"), plugins.GetType("office"))

	e = suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 10)
	assert.Equal(suite.T(), string(plugins.GetState("deleted")), string(e.Payload.(plugins.ProjectExtension).State))
}

func strMapKeys(strMap map[string]string) string {
	keys := make([]string, len(strMap))

	i := 0
	for k := range strMap {
		keys[i] = k
		i++
	}

	return strings.Join(keys, "\n")
}

// Deploys Tests
func (suite *TestSuite) TestBasicSuccessDeploy() {
	timer := time.NewTimer(time.Second * 60)
	defer timer.Stop()

	var e transistor.Event
	suite.transistor.Events <- testdata.BasicReleaseEvent()
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 5)
	assert.Equal(suite.T(), string(plugins.GetState("running")), string(e.Payload.(plugins.ReleaseExtension).State))

	go func() {
		<-timer.C
		log.Fatal("TestBasicSuccessDeploy: Test timeout")
	}()

	for {
		e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 30)
		if e.Payload.(plugins.ReleaseExtension).State != "running" {
			break
		}
	}

	suite.T().Log(e.Payload.(plugins.ReleaseExtension).StateMessage)
	assert.Equal(suite.T(), string(plugins.GetState("complete")), string(e.Payload.(plugins.ReleaseExtension).State))
}

func (suite *TestSuite) TestBasicFailedDeploy() {
	timer := time.NewTimer(time.Second * 60)
	defer timer.Stop()

	go func() {
		<-timer.C
		log.Fatal("TestBasicFailedDeploy: Test timeout")
	}()

	var e transistor.Event
	suite.transistor.Events <- testdata.BasicFailedReleaseEvent()

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 5)
	assert.Equal(suite.T(), string(plugins.GetState("running")), string(e.Payload.(plugins.ReleaseExtension).State))

	for {
		e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 30)
		if e.Payload.(plugins.ReleaseExtension).State != "running" {
			break
		}
	}

	suite.T().Log(e.Payload.(plugins.ReleaseExtension).StateMessage)
	assert.Equal(suite.T(), string(plugins.GetState("failed")), string(e.Payload.(plugins.ReleaseExtension).State))
}

func TestDeployments(t *testing.T) {
	proceed := true

	if err := verifyDeploymentArtifacts(); err != nil {
		proceed = false
		assert.Nil(t, err, err.Error())
	}

	if err := verifyLoadBalancerArtifacts(); err != nil {
		proceed = false
		assert.Nil(t, err, err.Error())
	}

	if proceed {
		suite.Run(t, new(TestSuite))
	}
}

func (suite *TestSuite) TearDownSuite() {
	// TODO:
	// teardown docker-io secret?
	// teardown the deployment / namespaces
	suite.transistor.Stop()
}

func verifyDeploymentArtifacts() error {
	e := testdata.BasicReleaseEvent()

	basicReleaseEventArtifacts := map[string]string{
		"user":                  "",
		"password":              "",
		"host":                  "",
		"email":                 "",
		"image":                 "",
		"kubeconfig":            "",
		"client_certificate":    "",
		"client_key":            "",
		"certificate_authority": "",
	}

	for _, artifact := range e.Artifacts {
		delete(basicReleaseEventArtifacts, artifact.Key)
	}

	if len(basicReleaseEventArtifacts) != 0 {
		return errors.New("BasicReleaseEvent\nMissing Artifacts:\n" + strMapKeys(basicReleaseEventArtifacts))
	}

	return nil
}

func verifyLoadBalancerArtifacts() error {
	e := testdata.LBTCPEvent(plugins.GetAction("update"), plugins.GetType("office"))

	lbTCPArtifacts := map[string]string{
		"service":               "",
		"name":                  "",
		"ssl_cert_arn":          "",
		"access_log_s3_bucket":  "",
		"type":                  "",
		"kubeconfig":            "",
		"client_certificate":    "",
		"client_key":            "",
		"certificate_authority": "",
		"listener_pairs":        "",
	}

	for _, artifact := range e.Artifacts {
		delete(lbTCPArtifacts, artifact.Key)
	}

	if len(lbTCPArtifacts) != 0 {
		return errors.New("LoadBalancer\nMissing Artifacts:\n" + strMapKeys(lbTCPArtifacts))
	}

	return nil
}
