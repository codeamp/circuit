package codeamp

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/satyakb/go-socket.io-redis"
	"github.com/spf13/viper"
)

var _schema *graphql.Schema

func init() {
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return NewCodeAmp() })
}

type CodeAmp struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	Schema         *graphql.Schema
	Actions        *actions.Actions
	SocketIO       *socketio.Server
	Db             *gorm.DB
}

func NewCodeAmp() *CodeAmp {
	return &CodeAmp{}
}

func (x *CodeAmp) Migrate() {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		log.Fatal(err)
	}

	db.LogMode(true)

	db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.Project{},
		&models.Release{},
		&models.Feature{},
		&models.Service{},
		&models.ContainerPort{},
		&models.ServiceSpec{},
		&models.ExtensionSpec{},
		&models.Extension{},
		&models.EnvironmentVariable{},
		&models.ExtensionSpecEnvironmentVariable{},
		&models.ReleaseExtension{},
		&models.Environment{},
	)

	hashedPassword, _ := utils.HashPassword("password")

	user := models.User{
		Email:    "admin@codeamp.com",
		Password: hashedPassword,
	}

	db.FirstOrInit(&user, models.User{
		Email: "admin@codeamp.com",
	})
	db.Save(&user)

	userPermission := models.UserPermission{
		UserId: user.Model.ID,
		Value:  "admin",
	}
	db.FirstOrInit(&userPermission, userPermission)
	db.Save(&userPermission)

	developmentEnv := models.Environment{
		Name: "development",
	}
	db.FirstOrInit(&developmentEnv, developmentEnv)
	db.Save(&developmentEnv)

	productionEnv := models.Environment{
		Name: "production",
	}
	db.FirstOrInit(&productionEnv, productionEnv)
	db.Save(&productionEnv)

	serviceSpec := models.ServiceSpec{
		Name:                   "default",
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	db.FirstOrInit(&serviceSpec, models.ServiceSpec{
		Name: "default",
	})
	db.Save(&serviceSpec)

	extensionSpec := models.ExtensionSpec{
		Type:      plugins.Workflow,
		Key:       "dockerbuilder",
		Name:      "Docker Builder",
		Component: "",
		FormSpec: plugins.MapStringStringToHstore(map[string]string{
			"REGISTRY": "required|string",
			"USERNAME": "required|string",
			"PASSWORD": "required|string",
		}),
	}
	db.FirstOrInit(&extensionSpec, models.ExtensionSpec{
		Key: extensionSpec.Key,
	})
	db.Save(&extensionSpec)

	extensionSpec = models.ExtensionSpec{
		Type:      plugins.Deployment,
		Key:       "kubernetesdeployment",
		Name:      "Kubernetes",
		Component: "",
	}
	db.FirstOrInit(&extensionSpec, models.ExtensionSpec{
		Key: extensionSpec.Key,
	})
	db.Save(&extensionSpec)

	defer db.Close()
}

//Custom server which basically only contains a socketio variable
//But we need it to enhance it with functions
type socketIOServer struct {
	Server *socketio.Server
}

//Header handling, this is necessary to adjust security and/or header settings in general
//Please keep in mind to adjust that later on in a productive environment!
//Access-Control-Allow-Origin will be set to whoever will call the server
func (s *socketIOServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

func (x *CodeAmp) Listen() {
	x.SocketIO.On("connection", func(so socketio.Socket) {
		so.Join("general")
	})

	x.SocketIO.On("error", func(so socketio.Socket, err error) {
		log.Println("socket-io error:", err)
	})

	sIOServer := new(socketIOServer)
	sIOServer.Server = x.SocketIO
	http.Handle("/socket.io/", sIOServer)

	_, filename, _, _ := runtime.Caller(0)
	fs := http.FileServer(http.Dir(path.Join(path.Dir(filename), "static/")))
	http.Handle("/", fs)
	http.Handle("/query", utils.CorsMiddleware(utils.AuthMiddleware(&relay.Handler{Schema: x.Schema}, x.Db)))

	log.Info(fmt.Sprintf("running GraphQL server on %v", x.ServiceAddress))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *CodeAmp) Start(events chan transistor.Event) error {
	var err error

	x.Events = events

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	//defer x.DB.Close()

	actions := actions.NewActions(events, db)
	resolver := resolvers.NewResolver(events, db, actions)

	s, err := schema.Schema()
	if err != nil {
		log.Fatal(err)
	}

	parsedSchema, err := graphql.ParseSchema(string(s), resolver)
	if err != nil {
		panic(err)
	}

	// Socket-io
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	split := strings.Split(viper.GetString("redis.server"), ":")
	host, port := split[0], split[1]

	opts := map[string]string{
		"host": host,
		"port": port,
	}
	sio.SetAdaptor(redis.Redis(opts))

	x.SocketIO = sio
	x.Actions = actions
	x.Schema = parsedSchema

	// DEBUG
	db.LogMode(true)

	x.Db = db

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
		"plugins.HeartBeat",
		"plugins.WebsocketMsg",
		"plugins.Extension:status",
		"plugins.Extension:complete",
		"plugins.ReleaseExtension:status",
		"plugins.Release:status",
		"plugins.Release:complete",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	log.InfoWithFields("process CodeAmp event", log.Fields{
		"event_name": e.Name,
	})

	if e.Matches("plugins.HeartBeat") {
		heartBeat := e.Payload.(plugins.HeartBeat)
		switch heartBeat.Tick {
		case "minute":
			x.Actions.HeartBeat("minute")
		}
		return nil
	}

	if e.Matches("plugins.GitCommit") {
		payload := e.Payload.(plugins.GitCommit)
		x.Actions.GitCommit(payload)
	}

	if e.Matches("plugins.WebsocketMsg") {
		payload := e.Payload.(plugins.WebsocketMsg)

		if payload.Channel == "" {
			payload.Channel = "general"
		}

		x.SocketIO.BroadcastTo(payload.Channel, payload.Event, payload.Payload, nil)
	}

	if e.Matches("plugins.Extension:complete") {
		payload := e.Payload.(plugins.Extension)
		var extension models.Extension

		if x.Db.Where("id = ?", payload.Id).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": payload.Id,
			})
			return nil
		}

		extension.State = plugins.Complete
		extension.Artifacts = plugins.MapStringStringToHstore(payload.Artifacts)
		x.Db.Save(&extension)

		x.Actions.ExtensionInitCompleted(&extension)
	}

	if e.Matches("plugins.ReleaseExtension:status") {
		payload := e.Payload.(plugins.ReleaseExtension)
		var releaseExtension models.ReleaseExtension

		if x.Db.Where("id = ?", payload.Id).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.Id,
			})
			return nil
		}

		releaseExtension.State = payload.State
		releaseExtension.Artifacts = plugins.MapStringStringToHstore(payload.Artifacts)
		x.Db.Save(&releaseExtension)

		if payload.State == plugins.Complete {
			x.Actions.ReleaseExtensionCompleted(&releaseExtension)
		}
	}

	return nil
}
