package codeamp

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	socketio "github.com/azhao1981/go-socket.io"
	sioredis "github.com/azhao1981/go-socket.io-redis"
	"github.com/codeamp/circuit/assets"
	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/helpers"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	redis "github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"

	uuid "github.com/satori/go.uuid"
)

// This UUID is utilized by the release process and the slack notifier plugin
// this will handle the case when a deployment is automated and there is
// no explicit user assigned to the release. Anywhere there is a check
// involving the user, we check to see if it's this one first.
// If it is, we proceed as if it was an automated deployment
const ContinuousDeployUUID = "59ee0229-9814-4d9b-be83-7921fe6069c1"

func init() {
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin {
		return &CodeAmp{}
	},
		plugins.Project{},
		plugins.HeartBeat{},
		plugins.GitSync{},
		plugins.WebsocketMsg{},
		plugins.ProjectExtension{},
		plugins.ReleaseExtension{},
		plugins.Release{})
}

type CodeAmp struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	Schema         *graphql.Schema
	SocketIO       *socketio.Server
	DB             *gorm.DB
	Redis          *redis.Client
	Resolver       *graphql_resolver.Resolver
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

func (x *CodeAmp) GraphQLListen() {
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
	fs := http.FileServer(http.Dir(path.Join(path.Dir(filename), "graphql/static/")))
	http.Handle("/", fs)

	middleware := graphql_resolver.Middleware{x.Resolver}
	http.Handle("/query", middleware.Cors(middleware.Auth(&relay.Handler{Schema: x.Schema})))

	log.Info(fmt.Sprintf("Running GraphQL server on %v", x.ServiceAddress))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *CodeAmp) initPostGres() (*gorm.DB, error) {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		return nil, err
	}

	// DEBUG
	//db.LogMode(false)

	x.DB = db
	return db, nil
}

func (x *CodeAmp) InitGraphQL(resolver interface{}) (*graphql.Schema, error) {
	schema, err := assets.Asset("plugins/codeamp/graphql/schema.graphql")
	if err != nil {
		return nil, err
	}

	return graphql.ParseSchema(string(schema), resolver)
}

func (x *CodeAmp) initRedis() {
	// Socket-io
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	split := strings.Split(viper.GetString("redis.server"), ":")
	host, port := split[0], split[1]

	opts := map[string]string{
		"host":   host,
		"port":   port,
		"prefix": "socket.io",
	}
	sio.SetAdaptor(sioredis.Redis(opts))

	redisDb, err := strconv.Atoi(viper.GetString("redis.database"))
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.server"),
		Password: viper.GetString("redis.password"),
		DB:       redisDb,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		log.Fatal(err)
	}

	x.SocketIO = sio
	x.Redis = redisClient
}

func (x *CodeAmp) Start(events chan transistor.Event) error {
	_, err := x.initPostGres()
	if err != nil {
		return err
	}

	x.initRedis()

	x.Events = events

	x.Resolver = &graphql_resolver.Resolver{DB: x.DB, Events: x.Events, Redis: x.Redis}
	x.Schema, err = x.InitGraphQL(x.Resolver)

	if err != nil {
		return err
	}

	go x.GraphQLListen()

	log.Info("Starting CodeAmp service")
	return nil
}

func (x *CodeAmp) Stop() {
	log.Info("Stopping CodeAmp service")
}

func (x *CodeAmp) Subscribe() []string {
	return []string{
		"gitsync:status",
		"heartbeat",
		"websocket",
		"project",
		"release",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	methodName := fmt.Sprintf("%sEventHandler", strings.Split(e.PayloadModel, ".")[1])

	if _, ok := reflect.TypeOf(x).MethodByName(methodName); ok {
		reflect.ValueOf(x).MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(e)})
	} else {
		log.InfoWithFields("*EventHandler not implemented", log.Fields{
			"event_name":    e.Name,
			"payload_model": e.PayloadModel,
			"method_name":   methodName,
		})
	}

	return nil
}

// SendNotifications dispatches notification events to registered project extension of type notification
func (x *CodeAmp) SendNotifications(releaseState string, release *model.Release, project *model.Project) error {
	var projectExtensions []model.ProjectExtension

	environment := model.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return fmt.Errorf("environment not found")
	}

	if x.DB.Where("project_id = ? and environment_id = ?", project.ID, environment.ID).Find(&projectExtensions).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": project.ID,
		})
		return fmt.Errorf("project not found")
	}

	var headFeature model.Feature
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.WarnWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return nil
	}

	var tailFeature model.Feature
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.WarnWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return nil
	}

	projectSettings := model.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.WarnWithFields("no env project branch found", log.Fields{})
		return nil
	}

	user := model.User{}
	if release.UserID != uuid.FromStringOrNil(ContinuousDeployUUID) {
		if x.DB.Where("id = ?", release.UserID).First(&user).RecordNotFound() {
			log.InfoWithFields("NOTIFICATIONS - user not found", log.Fields{
				"id": release.UserID,
			})
			return nil
		}
	} else {
		user.Email = "Automated Deployment"
	}

	projectModel := plugins.Project{
		ID:         project.Model.ID.String(),
		Slug:       project.Slug,
		Repository: project.Repository,
	}

	for _, pe := range projectExtensions {
		extension := model.Extension{}
		if x.DB.Where("id = ? and type = ?", pe.ExtensionID, plugins.GetType("notification")).Find(&extension).RecordNotFound() == false {

			projectExtensionArtifacts, _ := helpers.ExtractArtifacts(pe, extension, x.DB)
			_artifacts := []transistor.Artifact{}

			for _, artifact := range projectExtensionArtifacts {
				_artifacts = append(_artifacts, artifact)
			}

			notificationEvent := plugins.NotificationExtension{
				ID: pe.Model.ID.String(),
				Release: plugins.Release{
					ID:          release.Model.ID.String(),
					Environment: environment.Key,
					HeadFeature: plugins.Feature{
						Hash:       headFeature.Hash,
						ParentHash: headFeature.ParentHash,
						User:       headFeature.User,
						Message:    headFeature.Message,
						Created:    headFeature.Created,
					},
					TailFeature: plugins.Feature{
						ID:         tailFeature.Model.ID.String(),
						Hash:       tailFeature.Hash,
						ParentHash: tailFeature.ParentHash,
						User:       tailFeature.User,
						Message:    tailFeature.Message,
						Created:    tailFeature.Created,
					},
					User:    user.Email,
					Project: projectModel,
					Git: plugins.Git{
						Url:           project.GitUrl,
						Branch:        projectSettings.GitBranch,
						RsaPrivateKey: project.RsaPrivateKey,
					},
				},
				Project:     projectModel,
				Environment: environment.Key,
			}

			event := transistor.NewEvent(transistor.EventName(fmt.Sprintf("%s:notify", extension.Key)), transistor.GetAction("status"), notificationEvent)
			event.Artifacts = _artifacts
			event.AddArtifact("MESSAGE", releaseState, false)
			x.Events <- event
		}
	}

	return nil
}
