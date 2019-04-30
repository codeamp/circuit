package metrics

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

type PostgresCollectorOpts struct {
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
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		return float64(0)
	}

	return float64(1)
}
