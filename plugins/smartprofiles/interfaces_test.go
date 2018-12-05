package smartprofiles_test

import (
	"fmt"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/davecgh/go-spew/spew"
	smartprofiles "github.com/codeamp/circuit/plugins/smartprofiles"
)

type MockInfluxClient struct {}

func (ic *MockInfluxClient) InitInfluxClient(influxHost string, influxDBName string) (error) {
	spew.Dump("MockInflixClient InitInfluxClient")
	return nil
}

func (ic MockInfluxClient) GetService(id string, name string, namespace string, timeRange string, svcChan chan *smartprofiles.Service) {
	spew.Dump("MockInfluxClient GetService")
	svcChan <- &smartprofiles.Service{}
	return
}

func (ic MockInfluxClient) QueryDB(cmd string) (res []client.Result, err error) {
	spew.Dump("MockInflixClient QueryDB")
	return []client.Result{}, fmt.Errorf("error")
}