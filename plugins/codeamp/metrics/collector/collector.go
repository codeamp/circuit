package collector

import (
	"fmt"
	"unsafe"

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
		Name:      "continuous_deploy",
		Help:      "To show if project uses continuous deployment",
	}, []string{"project", "environment"})

	onMasterGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "project",
		Name:      "on_master_branch",
		Help:      "To show if project uses master branch",
	}, []string{"project", "environment", "branch"})
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
	c := &Collector{
		RedisOpts:    redisOpts,
		PostgresOpts: postgresOpts,
	}

	return c
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	upGauge.Reset()
	environmentGauge.Reset()
	continuousDeploymentGauge.Reset()
	onMasterGauge.Reset()

	c.collectRedis()
	c.collectPostgres()

	upGauge.Collect(ch)
	environmentGauge.Collect(ch)
	continuousDeploymentGauge.Collect(ch)
	onMasterGauge.Collect(ch)
}

func (c *Collector) collectRedis() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.RedisOpts.Host,
		Password: c.RedisOpts.Password,
		DB:       c.RedisOpts.DB,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		upGauge.WithLabelValues("redis").Set(0)
		return
	}

	upGauge.WithLabelValues("redis").Set(1)
}

func (c *Collector) collectPostgres() {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		c.PostgresOpts.Host,
		c.PostgresOpts.Port,
		c.PostgresOpts.User,
		c.PostgresOpts.DB,
		c.PostgresOpts.SSLMode,
		c.PostgresOpts.Password,
	))
	if err != nil {
		upGauge.WithLabelValues("postgres").Set(float64(0))
		return
	}

	upGauge.WithLabelValues("postgres").Set(float64(1))

	//db.LogMode(true)

	c.collectEnvironments(db)
	c.collectProjectSettings(db)
}

func (c *Collector) collectEnvironments(db *gorm.DB) {
	var projects []model.Project
	var environments []model.Environment

	db.Find(&projects)
	db.Find(&environments)

	for _, project := range projects {
		for _, environment := range environments {
			var projectEnvironment model.ProjectEnvironment
			if db.Where("environment_id = ? AND project_id = ?", environment.ID.String(), project.ID.String()).First(&projectEnvironment).RecordNotFound() {
				environmentGauge.WithLabelValues(project.Name, environment.Key).Set(float64(0))
			} else {
				environmentGauge.WithLabelValues(project.Name, environment.Key).Set(float64(1))
			}
		}
	}
}

func (c *Collector) collectProjectSettings(db *gorm.DB) {
	var projects []model.Project
	var environments []model.Environment

	db.Find(&projects)
	db.Find(&environments)

	for _, project := range projects {
		for _, environment := range environments {
			var projectSettings model.ProjectSettings
			if db.Where("environment_id = ? AND project_id = ?", environment.ID.String(), project.ID.String()).First(&projectSettings).RecordNotFound() {
				continuousDeploymentGauge.WithLabelValues(project.Name, environment.Key).Set(float64(-1))
				onMasterGauge.WithLabelValues(project.Name, environment.Key, "").Set(float64(-1))
			} else {
				continuousDeploymentGauge.WithLabelValues(project.Name, environment.Key).Set(float64(bool2int(projectSettings.ContinuousDeploy)))
				onMasterGauge.WithLabelValues(project.Name, environment.Key, projectSettings.GitBranch).Set(float64(onMaster2int(projectSettings.GitBranch)))
			}
		}
	}
}

func bool2int(a bool) uint64 {
	return *(*uint64)(unsafe.Pointer(&a)) & 1
}

func onMaster2int(a string) uint64 {
	if a == "master" {
		return uint64(1)
	} else {
		return uint64(0)
	}
}
