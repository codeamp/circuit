package collector

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	upGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "ext_service",
		Name:      "up",
		Help:      "To show if we can connect to external service",
	}, []string{"service"})

	environmentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "project",
		Name:      "environment",
		Help:      "To show if project uses environment",
	}, []string{"project", "environment"})

	continuousDeploymentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "project",
		Name:      "continuous_deployment",
		Help:      "To show if project uses continuous deployment",
	}, []string{"project", "environment"})
)

type RedisCollectorOpts struct {
	Host     string
	Password string
	DB       int
}

type PostgresCollectorOpts struct {
	Host     string
	Port     string
	User     string
	DB       string
	SSLMode  string
	Password string
}

type Collector struct {
	RedisOpts    RedisCollectorOpts
	PostgresOpts PostgresCollectorOpts
}

func NewCollector(redisOpts RedisCollectorOpts, postgresOpts PostgresCollectorOpts) *Collector {
	exporter := &Collector{
		RedisOpts:    redisOpts,
		PostgresOpts: postgresOpts,
	}

	return exporter
}

func (exporter *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(exporter, ch)
}

func (exporter *Collector) Collect(ch chan<- prometheus.Metric) {
	upGauge.Reset()
	environmentGauge.Reset()

	exporter.collectRedis()
	exporter.collectPostgres()

	upGauge.Collect(ch)
	environmentGauge.Collect(ch)
}

func (exporter *Collector) collectRedis() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     exporter.RedisOpts.Host,
		Password: exporter.RedisOpts.Password,
		DB:       exporter.RedisOpts.DB,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		upGauge.WithLabelValues("redis").Set(0)
		return
	}

	upGauge.WithLabelValues("redis").Set(1)
}

func (exporter *Collector) collectPostgres() {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		exporter.PostgresOpts.Host,
		exporter.PostgresOpts.Port,
		exporter.PostgresOpts.User,
		exporter.PostgresOpts.DB,
		exporter.PostgresOpts.SSLMode,
		exporter.PostgresOpts.Password,
	))
	if err != nil {
		upGauge.WithLabelValues("postgres").Set(float64(0))
		return
	}

	upGauge.WithLabelValues("postgres").Set(float64(1))

	//db.LogMode(true)

	exporter.collectEnvironments(db)
}

func (exporter *Collector) collectEnvironments(db *gorm.DB) {
	var projects []model.Project
	var environments []model.Environment

	db.Find(&projects)
	db.Find(&environments)

	for _, project := range projects {
		for _, environment := range environments {
			var projectEnvironment model.ProjectEnvironment
			if db.Where("environment_id = ? AND project_id = ?", environment.ID.String(), project.ID.String()).First(&projectEnvironment).RecordNotFound() {
				environmentGauge.WithLabelValues(project.Name, environment.Name).Set(float64(0))
			} else {
				environmentGauge.WithLabelValues(project.Name, environment.Name).Set(float64(1))
			}
		}
	}
}
