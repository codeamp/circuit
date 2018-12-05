package smartprofiles

import (
	client "github.com/influxdata/influxdb/client/v2"
	"fmt"
	"time"
)

type InfluxClienter interface {
	InitInfluxClient(string, string) (error)
	GetService(string, string, string, string, chan *Service)
	QueryDB(string) ([]client.Result, error)
}

type InfluxClient struct {
	Client       client.Client
	InfluxDBName string
}


func (ic *InfluxClient) InitInfluxClient(influxHost string, influxDBName string) (error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:    influxHost,
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return err
	}

	defer c.Close()

	ic.Client = c
	ic.InfluxDBName = influxDBName
	return nil
}


// Computes Service details for a given request
func (ic InfluxClient) GetService(id string, name string, namespace string, timeRange string, svcChan chan *Service) {
	fmt.Println(fmt.Sprintf("[...] appending %s - %s", name, namespace))
	memoryCost, err := ic.getServiceMemoryCost(name, namespace, timeRange)
	if err != nil {
		fmt.Println(err.Error())
		svcChan <- &Service{}
		return
	}

	cpuCost, err := ic.getServiceCPUCost(name, namespace, timeRange)
	if err != nil {
		fmt.Println(err.Error())
		svcChan <- &Service{}
		return
	}

	memRecommendation := ic.getResourceRecommendation(memoryCost)
	cpuRecommendation := ic.getResourceRecommendation(cpuCost)

	memDifference := ic.getResourceDiff(*memoryCost, *memRecommendation)
	cpuDifference := ic.getResourceDiff(*cpuCost, *cpuRecommendation)

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
func (ic InfluxClient) QueryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: ic.InfluxDBName,
	}

	if response, err := ic.Client.Query(q); err != nil {
		return nil, err
	} else {
		if response.Error() != nil {
			return nil, response.Error()
		} else {
			return response.Results, nil
		}
	}
}