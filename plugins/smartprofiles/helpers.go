package smartprofiles

import (
	"strconv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/olekukonko/tablewriter"
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
	Name                  string     `json:"name"`
	Namespace             string     `json:"namespace"`
	CurrentState          State      `json:"currentState"`
	RecommendedState 	  State      `json:"recommendedState"`
	StateDifference       State      `json:"difference"`
}

type State struct {
	Memory Resource `json:"memory"`
	CPU    Resource `json:"cpu"` 	
}

type Resource struct {
	Request string `json:"request"`
	Current string `json:"current"`	
	P90 string `json:"p90"`
	Limit string `json:"limit"`
	OverProvisioned bool `json:"overProvisioned"`
}

type InfluxClienter struct {
	Client client.Client
}

func RenderResourceMonitor() *tablewriter.Table {
	fmt.Println("main")

	// initialize influx client connection
	influxClient, err := InitInfluxClient()
	if err != nil {
		panic(err)
	}

	spew.Dump("initialized influx client")
	res, err := influxClient.queryDB("select * from kubernetes_pod_container where time > now() - " + os.Args[1] + "")
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "Service Name",
		"Memory OP", "CPU OP",
		"Memory Request (gb)", "Memory Current Usage", "Memory P90 Usage", "Memory Limit",
		"CPU Request (cpu)", "CPU Current Usage (cpu)", "CPU P90 Usage", "CPU Limit (cpu)"})

	// var wg sync.WaitGroup
	// wg.Add(len(res[0].Series[0].Values))

	fmt.Println(fmt.Sprintf("[...] starting"))

	visited := map[string]bool{}
	// values := res[0].Series[0].Values[0:20]

	values := [][]interface{}{}

	for _, row := range res[0].Series[0].Values {
		containerName := row[2].(string)
		if visited[containerName] {
			continue
		} else {
			visited[containerName] = true
			values = append(values, row)
		}
	}

	// wg.Add(len(values))

	ch := make(chan *Service)

	for _, row := range values {
		containerName := row[2].(string)
		namespace := row[15].(string)

		// go func(containerName string, namespace string, influxClient *InfluxClienter) {

		// defer wg.Done()
		go influxClient.GetService(containerName, namespace, "72h", ch)
		// }(containerName, namespace, influxClient)
	}

	// for range values {
	// 	var svc *Service
	// 	svc = <-ch

	// 	table.Append([]string{
	// 		svc.Namespace, svc.Name,
	// 		svc.Costs.Memory.OverProvisioned, svc.Costs.CPU.OverProvisioned,
	// 		svc.Costs.Memory.Min, svc.Costs.Memory.Current, svc.Costs.Memory.P90, svc.Costs.Memory.Max,
	// 		svc.Costs.CPU.Min, svc.Costs.CPU.Current, svc.Costs.CPU.P90, svc.Costs.CPU.Max,
	// 	})

	// 	fmt.Println(fmt.Sprintf("[+] appended %s - %s", svc.Name, svc.Namespace))
	// }

	// wg.Wait()

	return table
}

func InitInfluxClient() (*InfluxClienter, error) {
	// spew.Dump(os.Getenv("INFLUX_HOST"))
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "",
		// Username:           os.Getenv("INFLUX_USERNAME"),
		// Password:           os.Getenv("INFLUX_PASSWORD"),
	})
	if err != nil {
		return nil, err
	}

	defer c.Close()

	return &InfluxClienter{Client: c}, nil
}

func (ic *InfluxClienter) GetService(name string, namespace string, timeRange string, svcChan chan *Service) {
	fmt.Println(fmt.Sprintf("[...] appending %s - %s", name, namespace))

	memoryCost, err := ic.getServiceMemoryCost(name, namespace, timeRange)
	if err != nil {
		spew.Dump(err)
		panic(err.Error())
	}

	cpuCost, err := ic.getServiceCPUCost(name, namespace, timeRange)
	if err != nil {
		panic(err.Error())
	}

	memRecommendation := ic.getResourceRecommendation(memoryCost)
	cpuRecommendation := ic.getResourceRecommendation(cpuCost)

	memDifference := ic.getResourceDiff(*memoryCost, *memRecommendation)
	cpuDifference := ic.getResourceDiff(*cpuCost, *cpuRecommendation)

	svcChan <- &Service{
		Name:      name,
		Namespace: namespace,
		CurrentState: State{
			Memory: *memoryCost,
			CPU:    *cpuCost,
		},
		RecommendedState: State{
			Memory: *memRecommendation,
			CPU: *cpuRecommendation,
		},
		StateDifference: State{
			Memory: *memDifference,
			CPU: *cpuDifference,
		},
	}
}

func (ic *InfluxClienter) getResourceDiff(current Resource, recommended Resource) *Resource {
	r := Resource{}

	// get request diff
	recRequestFloat, _ := strconv.ParseFloat(recommended.Request, 64) 
	curRequestFloat, _ := strconv.ParseFloat(current.Request, 64)
	r.Request = fmt.Sprintf("%2f", recRequestFloat - curRequestFloat)
	
	// get limit diff
	recLimitFloat, _ := strconv.ParseFloat(recommended.Limit, 64)
	curLimitFloat, _ := strconv.ParseFloat(current.Limit, 64)		
	r.Limit =  fmt.Sprintf("%2f", recLimitFloat - curLimitFloat)

	return &r
}

func (ic *InfluxClienter) getResourceRecommendation(cost *Resource) *Resource {
	r := Resource{}
	// request recommendation - 20% above usage
	currentFloat, _ := strconv.ParseFloat(cost.Current, 64)
	r.Request = fmt.Sprintf("%f", float64(1.2) * currentFloat)

	// limit recommendation - 20% above p90
	p90Float, _ := strconv.ParseFloat(cost.P90, 64)
	r.Limit = fmt.Sprintf("%f", float64(1.2) * p90Float)
	r.OverProvisioned = cost.OverProvisioned

	return &r
}

func (ic *InfluxClienter) getServiceMemoryCost(serviceName string, namespace string, timeRange string) (*Resource, error) {
	// memory_current_cost --> telegraf.autogen.kubernetes_pod_container.memory_usage_bytes
	// memory_min_cost     --> telegraf.autogen.prom_kube_pod_container_resource_requests_memory_bytes.gauge
	// memory_max_cost     --> telegraf.autogen.prom_kube_pod_container_resource_limits_memory_bytes.gauge

	minCost := json.Number("")
	currentCost := json.Number("")
	maxCost := json.Number("")
	p90 := json.Number("")

	// min > current by atleast 20%
	overProvisioned := false

	// get current cost
	res, err := ic.queryDB(fmt.Sprintf("select mean(memory_usage_bytes)/1000000000 from kubernetes_pod_container where time > now() - " + timeRange + " and container_name = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err := getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		currentCost = resFirstValue.(json.Number)
	}

	// get min cost
	res, err = ic.queryDB(fmt.Sprintf("select mean(gauge)/1000000000 from prom_kube_pod_container_resource_requests_memory_bytes where time > now() - " + timeRange + " and container = '%s' and namespace ='%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err = getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		minCost = resFirstValue.(json.Number)
	}

	// get max cost
	res, err = ic.queryDB(fmt.Sprintf("select mean(gauge)/1000000000 from prom_kube_pod_container_resource_limits_memory_bytes where time > now() - " + timeRange + " and container = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err = getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		maxCost = resFirstValue.(json.Number)
	}

	// get p90
	res, err = ic.queryDB(fmt.Sprintf("select percentile(memory_usage_bytes, 90)/1000000000 from kubernetes_pod_container where time > now() - " + timeRange + " and container_name = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err = getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		p90 = resFirstValue.(json.Number)
	}	

	// check if overprovisioned
	maxCostFloat, _ := maxCost.Float64()
	minCostFloat, _ := minCost.Float64()
	currentCostFloat, _ := currentCost.Float64()
	p90Float, _ := p90.Float64()

	if minCostFloat*0.8 > currentCostFloat {
		overProvisioned = true
	}

	return &Resource{
		Limit:             fmt.Sprintf("%.2f", maxCostFloat),
		Request:             fmt.Sprintf("%.2f", minCostFloat),
		Current:         fmt.Sprintf("%.2f", currentCostFloat),
		P90:         	 fmt.Sprintf("%.2f", p90Float),
		OverProvisioned: overProvisioned,
	}, nil
}

func (ic *InfluxClienter) getServiceCPUCost(serviceName string, namespace string, timeRange string) (*Resource, error) {
	// cpu_current_cost    --> telegraf.autogen.kubernetes_pod_container.cpu_usage_nanocores
	// cpu_min_cost        --> telegraf.autogen.prom_kube_pod_container_resource_requests_cpu_cores.gauge
	// cpu_max_cost        --> telegraf.autogen.prom_kube_pod_container_resource_limits_cpu_cores.gauge
	minCost := json.Number("")
	currentCost := json.Number("")
	maxCost := json.Number("")
	p90 := json.Number("")

	// min > current by atleast 20%
	overProvisioned := false

	// get current cost
	res, err := ic.queryDB(fmt.Sprintf("select mean(cpu_usage_nanocores)/100000000 from kubernetes_pod_container where time > now() - " + timeRange + " and container_name = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err := getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		currentCost = resFirstValue.(json.Number)
	}

	// get min cost
	res, err = ic.queryDB(fmt.Sprintf("select mean(gauge) from prom_kube_pod_container_resource_requests_cpu_cores where time > now() - " + timeRange + " and container = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err = getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		minCost = resFirstValue.(json.Number)
	}

	// get maxCost
	res, err = ic.queryDB(fmt.Sprintf("select mean(gauge) from prom_kube_pod_container_resource_limits_cpu_cores where time > now() - " + timeRange + " and container = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err = getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		maxCost = resFirstValue.(json.Number)
	}

	// get p90
	res, err = ic.queryDB(fmt.Sprintf("select percentile(cpu_usage_nanocores, 90)/1000000000 from kubernetes_pod_container where time > now() - " + timeRange + " and container_name = '%s' and namespace = '%s'", serviceName, namespace))
	if err != nil {
		return nil, err
	}

	resFirstValue, err = getFirstValue(res)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		p90 = resFirstValue.(json.Number)
	}		

	maxCostFloat, _ := maxCost.Float64()
	// check if overprovisioned
	minCostFloat, _ := minCost.Float64()
	currentCostFloat, _ := currentCost.Float64()
	p90Float, _ := p90.Float64()

	if minCostFloat*0.8 > currentCostFloat {
		overProvisioned = true
	}

	return &Resource{
		Limit:             fmt.Sprintf("%.2f", maxCostFloat),
		Request:             fmt.Sprintf("%.2f", minCostFloat),
		Current:         fmt.Sprintf("%.2f", currentCostFloat),
		P90:             fmt.Sprintf("%.2f", p90Float),
		OverProvisioned: overProvisioned,
	}, nil
}

// queryDB convenience function to query the database
func (ic *InfluxClienter) queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: "telegraf",
	}
	if response, err := ic.Client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func getFirstValue(res []client.Result) (interface{}, error) {
	if len(res) > 0 && len(res[0].Series) > 0 && len(res[0].Series[0].Values) > 0 && len(res[0].Series[0].Values[0]) > 1 {
		return res[0].Series[0].Values[0][1], nil
	} else {
		return nil, fmt.Errorf("length of result is 0")
	}
}
