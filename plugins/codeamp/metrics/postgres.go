package metrics

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
)

type PostgresCollectorOpts struct {
	Host     string
	Port     string
	User     string
	DB       string
	SSLMode  string
	Password string
}

type PostgresCollector struct {
	Opts PostgresCollectorOpts
}

func NewPostgresCollector(opts PostgresCollectorOpts) *PostgresCollector {
	exporter := &PostgresCollector{
		Opts: opts,
	}

	return exporter
}

func (exporter *PostgresCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(exporter, ch)
}

func (exporter *PostgresCollector) Collect(ch chan<- prometheus.Metric) {
	postgresUpGauge.WithLabelValues().Set(exporter.up())
	postgresUpGauge.Collect(ch)
	postgresUpGauge.Reset()
}

func (exporter *PostgresCollector) up() float64 {
	_, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		exporter.Opts.Host,
		exporter.Opts.Port,
		exporter.Opts.User,
		exporter.Opts.DB,
		exporter.Opts.SSLMode,
		exporter.Opts.Password,
	))
	if err != nil {
		return float64(0)
	}

	return float64(1)
}
