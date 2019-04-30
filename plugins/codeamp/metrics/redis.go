package metrics

import (
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

type RedisCollectorOpts struct {
}

type RedisCollector struct {
	Opts RedisCollectorOpts
}

func NewRedisCollector(opts RedisCollectorOpts) *RedisCollector {
	exporter := &RedisCollector{
		Opts: opts,
	}

	return exporter
}

func (exporter *RedisCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(exporter, ch)
}

func (exporter *RedisCollector) Collect(ch chan<- prometheus.Metric) {
	redisUpGauge.WithLabelValues().Set(exporter.up())
	redisUpGauge.Collect(ch)
	redisUpGauge.Reset()
}

func (exporter *RedisCollector) up() float64 {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.server"),
		Password: viper.GetString("redis.password"),
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		return float64(0)
	}

	return float64(1)
}
