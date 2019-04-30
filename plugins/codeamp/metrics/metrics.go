package metrics

import (
	"log"
	"net/http"
	"strconv"

	collector "github.com/codeamp/circuit/plugins/codeamp/metrics/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

func Handler() http.Handler {
	r := prometheus.NewRegistry()
	registerCollector(r)
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	return handler
}

func registerCollector(r *prometheus.Registry) {
	redisDb, err := strconv.Atoi(viper.GetString("redis.database"))
	if err != nil {
		log.Fatal(err)
	}

	Collector := collector.NewCollector(
		collector.RedisCollectorOpts{
			Host:     viper.GetString("redis.server"),
			Password: viper.GetString("redis.password"),
			DB:       redisDb,
		},
		collector.PostgresCollectorOpts{
			Host:     viper.GetString("plugins.codeamp.postgres.host"),
			Port:     viper.GetString("plugins.codeamp.postgres.port"),
			User:     viper.GetString("plugins.codeamp.postgres.user"),
			Password: viper.GetString("plugins.codeamp.postgres.password"),
			DB:       viper.GetString("plugins.codeamp.postgres.dbname"),
			SSLMode:  viper.GetString("plugins.codeamp.postgres.sslmode"),
		},
	)

	r.MustRegister(Collector)
}
