package smartprofiles_test

import (
	"fmt"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/davecgh/go-spew/spew"
	smartprofiles "github.com/codeamp/circuit/plugins/smartprofiles"
)

// MockInfluxClient
type MockInfluxClient struct{}

// Query method
func (m *MockInfluxClient) Query(q client.Query) (*client.Response, error) {
	return &client.Response{}, nil
}

// MockSmartProfilesClient
type MockSmartProfilesClient struct {
	InfluxClient MockInfluxClient
	InfluxDBName string
}

// InitInfluxClient
func (ic *MockSmartProfilesClient) InitInfluxClient(influxHost string, influxDBName string) (error) {
	ic.InfluxClient = MockInfluxClient{}
	ic.InfluxDBName = "telegraf"

	return nil
}

// GetService
func (ic *MockSmartProfilesClient) GetService(id string, name string, namespace string, timeRange string, svcChan chan *smartprofiles.Service) {
	spew.Dump("MockInfluxClient GetService")
	
	fmt.Println(fmt.Sprintf("[...] appending %s - %s", name, namespace))
	memoryCost, err := smartprofiles.GetServiceMemoryCost(ic, name, namespace, timeRange)
	if err != nil {
		fmt.Println(err.Error())
		svcChan <- &smartprofiles.Service{}
		return
	}

	cpuCost, err := smartprofiles.GetServiceCPUCost(ic, name, namespace, timeRange)
	if err != nil {
		fmt.Println(err.Error())
		svcChan <- &smartprofiles.Service{}
		return
	}

	memRecommendation := smartprofiles.GetResourceRecommendation(memoryCost)
	cpuRecommendation := smartprofiles.GetResourceRecommendation(cpuCost)

	memDifference := smartprofiles.GetResourceDiff(*memoryCost, *memRecommendation)
	cpuDifference := smartprofiles.GetResourceDiff(*cpuCost, *cpuRecommendation)

	service := &smartprofiles.Service{
		ID:        id,
		Name:      name,
		Namespace: namespace,
		CurrentState: smartprofiles.State{
			Memory: *memoryCost,
			CPU:    *cpuCost,
		},
		RecommendedState: smartprofiles.State{
			Memory: *memRecommendation,
			CPU:    *cpuRecommendation,
		},
		StateDifference: smartprofiles.State{
			Memory: *memDifference,
			CPU:    *cpuDifference,
		},
	}

	svcChan <- service
}

// QueryInfluxDB
func (ic *MockSmartProfilesClient) QueryInfluxDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: ic.InfluxDBName,
	}

	if response, err := ic.InfluxClient.Query(q); err != nil {
		return nil, err
	} else {
		if response.Error() != nil {
			return nil, response.Error()
		} else {
			return response.Results, nil
		}
	}	

	return []client.Result{}, nil
}