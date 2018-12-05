package smartprofiles_test

import (
	"fmt"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/davecgh/go-spew/spew"
	smartprofiles "github.com/codeamp/circuit/plugins/smartprofiles"
)

type MockInfluxClient struct {}

func (ic MockInfluxClient) InitInfluxClient(influxHost string, influxDBName string) (error) {
	spew.Dump("MockInflixClient InitInfluxClient")
	return nil
}

func (ic MockInfluxClient) GetService(id string, name string, namespace string, timeRange string, svcChan chan *smartprofiles.Service) {
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

func (ic MockInfluxClient) QueryDB(cmd string) (res []client.Result, err error) {
	spew.Dump("MockInflixClient QueryDB")
	return []client.Result{}, nil
}