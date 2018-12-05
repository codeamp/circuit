package smartprofiles

import (
	client "github.com/influxdata/influxdb/client/v2"
	"fmt"
	"time"
)

type SmartProfilesClienter interface {
	InitInfluxClient(string, string) error
	GetService(string, string, string, string, chan *Service)
	QueryInfluxDB(string) ([]client.Result, error)
}

type SmartProfilesClient struct {
	InfluxClient       client.Client
	InfluxDBName string
}


func (ic *SmartProfilesClient) InitInfluxClient(influxHost string, influxDBName string) (error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:    influxHost,
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return err
	}

	defer c.Close()

	ic.InfluxClient = c
	ic.InfluxDBName = influxDBName
	return nil
}


// Computes Service details for a given request
func (ic *SmartProfilesClient) GetService(id string, name string, namespace string, timeRange string, svcChan chan *Service) {
	fmt.Println(fmt.Sprintf("[...] appending %s - %s", name, namespace))
	memoryCost, err := GetServiceMemoryCost(ic, name, namespace, timeRange)
	if err != nil {
		fmt.Println(err.Error())
		svcChan <- &Service{}
		return
	}

	cpuCost, err := GetServiceCPUCost(ic, name, namespace, timeRange)
	if err != nil {
		fmt.Println(err.Error())
		svcChan <- &Service{}
		return
	}

	memRecommendation := GetResourceRecommendation(memoryCost)
	cpuRecommendation := GetResourceRecommendation(cpuCost)

	memDifference := GetResourceDiff(*memoryCost, *memRecommendation)
	cpuDifference := GetResourceDiff(*cpuCost, *cpuRecommendation)

	fmt.Println(memRecommendation, cpuRecommendation)

	service := &Service{
		ID:        id,
		Name:      name,
		Namespace: namespace,
		CurrentState: State{
			Memory: *memoryCost,
			CPU:    *cpuCost,
		},
		RecommendedState: State{
			Memory: *memRecommendation,
			CPU:    *cpuRecommendation,
		},
		StateDifference: State{
			Memory: *memDifference,
			CPU:    *cpuDifference,
		},
	}

	svcChan <- service
}

// QueryDB convenience function to query the database
func (ic *SmartProfilesClient) QueryInfluxDB(cmd string) (res []client.Result, err error) {
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
}