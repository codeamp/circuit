package metrics

import (
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
)

type RedisCollectorOpts struct {
	Host     string
	Password string
	DB       int
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
		Addr:     exporter.Opts.Host,
		Password: exporter.Opts.Password,
		DB:       exporter.Opts.DB,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		return float64(0)
	}

	return float64(1)
}
