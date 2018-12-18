package smartprofiles

import (
	"github.com/davecgh/go-spew/spew"
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/influxdata/influxdb/client/v2"
)

/*
Relevant influx measurements:
	1. memory_current_cost --> telegraf.autogen.kubernetes_pod_container.memory_usage_bytes
	2. memory_min_cost     --> telegraf.autogen.prom_kube_pod_container_resource_requests_memory_bytes.gauge
	3. memory_max_cost     --> telegraf.autogen.prom_kube_pod_container_resource_limits_memory_bytes.gauge
	4. cpu_current_cost    --> telegraf.autogen.kubernetes_pod_container.cpu_usage_nanocores
	5. cpu_min_cost        --> telegraf.autogen.prom_kube_pod_container_resource_requests_cpu_cores.gauge
	6. cpu_max_cost        --> telegraf.autogen.prom_kube_pod_container_resource_limits_cpu_cores.gauge
*/

// CPUUnitCost is the rate per hour to run a cpu on a r4.4xlarge node
const CPUUnitCost = 0.07

// MemUnitCost is the rate per hour to run a gb on a r4.4xlarge node
const MemUnitCost = 0.01

// Service is a representation of a codeamp service
type Service struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	CurrentState     State  `json:"currentState"`
	RecommendedState State  `json:"recommendedState"`
	StateDifference  State  `json:"difference"`
}

type State struct {
	Memory Resource `json:"memory"`
	CPU    Resource `json:"cpu"`
}

type Resource struct {
	Request         string `json:"request"`
	Current         string `json:"current"`
	P90             string `json:"p90"`
	Limit           string `json:"limit"`
	OverProvisioned bool   `json:"overProvisioned"`
}

func GetResourceDiff(current Resource, recommended Resource) *Resource {
	r := Resource{}

	// get request diff
	recRequestFloat, _ := strconv.ParseFloat(recommended.Request, 64)
	curRequestFloat, _ := strconv.ParseFloat(current.Request, 64)
	r.Request = fmt.Sprintf("%2f", recRequestFloat-curRequestFloat)

	// get limit diff
	recLimitFloat, _ := strconv.ParseFloat(recommended.Limit, 64)
	curLimitFloat, _ := strconv.ParseFloat(current.Limit, 64)
	r.Limit = fmt.Sprintf("%2f", recLimitFloat-curLimitFloat)

	return &r
}

func GetResourceRecommendation(cost *Resource) *Resource {
	r := Resource{}
	// request recommendation - 20% above usage
	currentFloat, _ := strconv.ParseFloat(cost.Current, 64)
	r.Request = fmt.Sprintf("%f", float64(1.2)*currentFloat)

	// limit recommendation - 20% above p90
	p90Float, _ := strconv.ParseFloat(cost.P90, 64)
	r.Limit = fmt.Sprintf("%f", float64(1.2)*p90Float)
	r.OverProvisioned = cost.OverProvisioned

	return &r
}

/*
ComputeMeanSampledQuery takes in a query and how many days to sample and returns
an average over those past x number of days. 

E.g. ComputeMeanSampledQuery("select mean(memory_usage_bytes)/100000 from kubernetes_pod_container", "container", "namespace")

- day 1, day 2, day 3
*/
func ComputeMeanSampledQuery(ic SmartProfilesClienter, measurement string, containerName string, namespace string, days int) (*float64, error) {
	// queryResults := []float64{}

	minTime := 1
	maxTime := 24
	sum := 0.0
	mean := 0.0

	for i := 1; i <= days; i++ {
		val := 0.0

		influxRes, err := ic.QueryInfluxDB(fmt.Sprintf("select %s where time > now() - %dh and time < now() - %dh", measurement, maxTime, minTime))		
		if err != nil {
			return nil, err
		}

		influxResFirstValue, err := getFirstValue(influxRes)
		if influxResFirstValue != nil {
			val, err = influxResFirstValue.(json.Number).Float64()
			if err != nil {
				return nil, err
			}	
		}

		sum += val
		
		minTime += 24
		maxTime += 24		
	}

	mean = sum / float64(days)

	return &mean, nil
}

func GetServiceMemoryCost(ic SmartProfilesClienter, serviceName string, namespace string, timeRange string) (*Resource, error) {
	// memory_current_cost --> telegraf.autogen.kubernetes_pod_container.memory_usage_bytes
	// memory_min_cost     --> telegraf.autogen.prom_kube_pod_container_resource_requests_memory_bytes.gauge
	// memory_max_cost     --> telegraf.autogen.prom_kube_pod_container_resource_limits_memory_bytes.gauge

	// min > current by atleast 20%
	overProvisioned := false

	// get current cost
	spew.Dump("computing mean")
	currentCostFloat, err := ComputeMeanSampledQuery(ic, "mean(memory_usage_bytes)/1000000 from kubernetes_pod_container", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}

	// get min cost
	minCostFloat, err := ComputeMeanSampledQuery(ic, "mean(gauge)/1000000 from prom_kube_pod_container_resource_requests_memory_bytes", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}

	// get max cost
	maxCostFloat, err := ComputeMeanSampledQuery(ic, "mean(gauge)/1000000 from prom_kube_pod_container_resource_limits_memory_bytes", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}

	// get p90
	p90Float, err := ComputeMeanSampledQuery(ic, "percentile(memory_usage_bytes, 90)/1000000 from kubernetes_pod_container", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}

	if *minCostFloat*0.8 > *currentCostFloat {
		overProvisioned = true
	}

	return &Resource{
		Limit:           fmt.Sprintf("%.2f", *maxCostFloat),
		Request:         fmt.Sprintf("%.2f", *minCostFloat),
		Current:         fmt.Sprintf("%.2f", *currentCostFloat),
		P90:             fmt.Sprintf("%.2f", *p90Float),
		OverProvisioned: overProvisioned,
	}, nil
}

func GetServiceCPUCost(ic SmartProfilesClienter, serviceName string, namespace string, timeRange string) (*Resource, error) {
	// cpu_current_cost    --> telegraf.autogen.kubernetes_pod_container.cpu_usage_nanocores
	// cpu_min_cost        --> telegraf.autogen.prom_kube_pod_container_resource_requests_cpu_cores.gauge
	// cpu_max_cost        --> telegraf.autogen.prom_kube_pod_container_resource_limits_cpu_cores.gauge

	// min > current by atleast 20%
	overProvisioned := false

	// get current cost
	currentCostFloat, err := ComputeMeanSampledQuery(ic, "mean(cpu_usage_nanocores)/10000 from kubernetes_pod_container", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}

	minCostFloat, err := ComputeMeanSampledQuery(ic, "mean(gauge) from prom_kube_pod_container_resource_requests_cpu_cores", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}	

	maxCostFloat, err := ComputeMeanSampledQuery(ic, "mean(gauge) from prom_kube_pod_container_resource_limits_cpu_cores", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}		

	// get p90
	p90Float, err := ComputeMeanSampledQuery(ic, "percentile(cpu_usage_nanocores, 90)/10000 from kubernetes_pod_container", serviceName, namespace, 14)
	if err != nil {
		spew.Dump(err)
		return nil, err
	}	

	if *minCostFloat*0.8 > *currentCostFloat {
		overProvisioned = true
	}

	return &Resource{
		Limit:           fmt.Sprintf("%.2f", *maxCostFloat),
		Request:         fmt.Sprintf("%.2f", *minCostFloat),
		Current:         fmt.Sprintf("%.2f", *currentCostFloat),
		P90:             fmt.Sprintf("%.2f", *p90Float),
		OverProvisioned: overProvisioned,
	}, nil
}

func getFirstValue(res []client.Result) (interface{}, error) {
	if len(res) > 0 && len(res[0].Series) > 0 && len(res[0].Series[0].Values) > 0 && len(res[0].Series[0].Values[0]) > 1 {
		return res[0].Series[0].Values[0][1], nil
	} else {
		return nil, fmt.Errorf("length of result is 0")
	}
}
