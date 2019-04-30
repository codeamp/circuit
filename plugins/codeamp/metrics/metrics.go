package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	postgresUpGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "postgres",
		Name:      "up",
		Help:      "To show if we can connect to postgres service",
	}, []string{})

	redisUpGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "redis",
		Name:      "up",
		Help:      "To show if we can connect to redis service",
	}, []string{})
)
