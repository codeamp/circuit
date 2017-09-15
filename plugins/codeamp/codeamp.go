package codeamp

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/codeamp/circuit/plugins"
	codeamp_models "github.com/codeamp/circuit/plugins/codeamp/models"
	codeamp_schema "github.com/codeamp/circuit/plugins/codeamp/schema"
	codeamp_schema_resolvers "github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/gorilla/handlers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/spf13/viper"
)

var _schema *graphql.Schema

func init() {
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return NewCodeAmp() })
}

type CodeAmp struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	DB             *gorm.DB
	Schema         *graphql.Schema
}

func NewCodeAmp() *CodeAmp {
	return &CodeAmp{}
}

func (x *CodeAmp) Migrate() {
	var err error

	x.DB, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		log.Fatal(err)
	}

	x.DB.Exec(fmt.Sprintf("CREATE DATABASE %s", viper.GetString("plugins.codeamp.postgres.dbname")))
	x.DB.Close()

	x.DB, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))

	x.DB.Exec("CREATE EXTENSION \"uuid-ossp\"")

	x.DB.AutoMigrate(
		&codeamp_models.User{},
		&codeamp_models.UserPermission{},
		&codeamp_models.Project{},
		&codeamp_models.Release{},
		&codeamp_models.Feature{},
	)

	hashedPassword, _ := utils.HashPassword("password")
	user := codeamp_models.User{
		Email:    "admin@codeamp.com",
		Password: hashedPassword,
	}
	x.DB.Create(&user)

	userPermission := codeamp_models.UserPermission{
		UserId: user.Model.ID,
		Value:  "admin",
	}
	x.DB.Create(&userPermission)

	defer x.DB.Close()
}

func (x *CodeAmp) Listen() {
	_, filename, _, _ := runtime.Caller(0)
	fs := http.FileServer(http.Dir(path.Join(path.Dir(filename), "static/")))
	http.Handle("/", fs)
	http.Handle("/query", utils.CorsMiddleware(utils.AuthMiddleware(&relay.Handler{Schema: x.Schema})))

	log.Info(fmt.Sprintf("running GraphQL server on %v", x.ServiceAddress))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *CodeAmp) Start(events chan transistor.Event) error {
	var err error

	x.Events = events

	x.DB, err = gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	//defer x.DB.Close()

	schema, err := codeamp_schema.Schema()
	if err != nil {
		log.Fatal(err)
	}

	x.Schema, err = graphql.ParseSchema(string(schema), &codeamp_schema_resolvers.Resolver{DB: x.DB})
	if err != nil {
		panic(err)
	}

	go x.Listen()

	return nil
}

func (x *CodeAmp) Stop() {
	log.Info("stopping CodeAmp service")
}

func (x *CodeAmp) Subscribe() []string {
	return []string{
		"plugins.GitPing",
		"plugins.GitCommit",
		"plugins.GitStatus",
		"plugins.DockerBuild:status",
		"plugins.HeartBeat",
		"plugins.LoadBalancer:status",
		"plugins.DockerDeploy:status",
		"plugins.Route53",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	log.InfoWithFields("process CodeAmp event", log.Fields{
		"event_name": e.Name,
	})

	if e.Name == "plugins.HeartBeat" {
		heartBeat := e.Payload.(plugins.HeartBeat)
		switch heartBeat.Tick {
		case "minute":
			x.HeartBeat("minute")
		}
		return nil
	}

	if e.Name == "plugins.GitCommit" {
		payload := e.Payload.(plugins.GitCommit)
		x.GitCommit(payload)
	}
	return nil
}
