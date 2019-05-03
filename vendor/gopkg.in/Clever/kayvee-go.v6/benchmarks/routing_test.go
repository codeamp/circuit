package benchmarks

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"gopkg.in/Clever/kayvee-go.v6/logger"
	"gopkg.in/Clever/kayvee-go.v6/router"
)

type logline struct {
	Title string                 `json:"title"`
	Data  map[string]interface{} `json:"data"`
}

var basicCorpus []logline
var pathologicalCorpus []logline
var realisticCorpus []logline

var noRouting logger.KayveeLogger
var basicRouting logger.KayveeLogger
var pathoRouting logger.KayveeLogger
var realRouting logger.KayveeLogger

func loadJSON(path string, o interface{}) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, o)
}

type noopWriter struct{}

func (n *noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func init() {
	var err error

	err = loadJSON("./data/corpus-basic.json", &basicCorpus)
	if err != nil {
		log.Fatal(err)
	}

	err = loadJSON("./data/corpus-pathological.json", &pathologicalCorpus)
	if err != nil {
		log.Fatal(err)
	}

	err = loadJSON("./data/corpus-realistic.json", &realisticCorpus)
	if err != nil {
		log.Fatal(err)
	}

	noRouting = logger.New("perf")

	basicRouting = logger.New("perf")
	basicRouter, err := router.NewFromConfig("./data/kvconfig-basic.yml")
	if err != nil {
		log.Fatal(err)
	}
	basicRouting.SetRouter(basicRouter)

	pathoRouting = logger.New("perf")
	pathoRouter, err := router.NewFromConfig("./data/kvconfig-pathological.yml")
	if err != nil {
		log.Fatal(err)
	}
	pathoRouting.SetRouter(pathoRouter)

	realRouting = logger.New("perf")
	realRouter, err := router.NewFromConfig("./data/kvconfig-realistic.yml")
	if err != nil {
		log.Fatal(err)
	}
	realRouting.SetRouter(realRouter)

	output := &noopWriter{}
	formatter := func(noop map[string]interface{}) string { return "" }

	noRouting.SetConfig("perf", logger.Debug, formatter, output)
	basicRouting.SetConfig("perf", logger.Debug, formatter, output)
	pathoRouting.SetConfig("perf", logger.Debug, formatter, output)
	realRouting.SetConfig("perf", logger.Debug, formatter, output)
}

// No routing
func BenchmarkNoRoutingWithBasicCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(basicCorpus); i++ {
			noRouting.Info(basicCorpus[i].Title)
		}
	}
}
func BenchmarkNoRoutingWithPathologicalCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(pathologicalCorpus); i++ {
			log := pathologicalCorpus[i]
			noRouting.InfoD(log.Title, log.Data)
		}
	}
}
func BenchmarkNoRoutingWithRealisticCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(realisticCorpus); i++ {
			log := realisticCorpus[i]
			noRouting.InfoD(log.Title, log.Data)
		}
	}
}

// Basic routing
func BenchmarkBasicRoutingWithBasicCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(basicCorpus); i++ {
			basicRouting.Info(basicCorpus[i].Title)
		}
	}
}
func BenchmarkBasicRoutingWithPathologicalCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(pathologicalCorpus); i++ {
			log := pathologicalCorpus[i]
			basicRouting.InfoD(log.Title, log.Data)
		}
	}
}
func BenchmarkBasicRoutingWithRealisticCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(realisticCorpus); i++ {
			log := realisticCorpus[i]
			basicRouting.InfoD(log.Title, log.Data)
		}
	}
}

// Pathological routing
func BenchmarkPathologicalRoutingWithBasicCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(basicCorpus); i++ {
			pathoRouting.Info(basicCorpus[i].Title)
		}
	}
}
func BenchmarkPathologicalRoutingWithPathologicalCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(pathologicalCorpus); i++ {
			log := pathologicalCorpus[i]
			pathoRouting.InfoD(log.Title, log.Data)
		}
	}
}
func BenchmarkPathologicalRoutingWithRealisticCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(realisticCorpus); i++ {
			log := realisticCorpus[i]
			pathoRouting.InfoD(log.Title, log.Data)
		}
	}
}

// Realistic routing
func BenchmarkRealisticRoutingWithBasicCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(basicCorpus); i++ {
			realRouting.Info(basicCorpus[i].Title)
		}
	}
}
func BenchmarkRealisticRoutingWithPathologicalCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(pathologicalCorpus); i++ {
			log := pathologicalCorpus[i]
			realRouting.InfoD(log.Title, log.Data)
		}
	}
}
func BenchmarkRealisticRoutingWithRealisticCorpus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(realisticCorpus); i++ {
			log := realisticCorpus[i]
			realRouting.InfoD(log.Title, log.Data)
		}
	}
}
