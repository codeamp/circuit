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
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	redis "github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

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
	fs := http.FileServer(http.Dir(path.Join(path.Dir(filename), "static/")))
	http.Handle("/", fs)

	http.Handle("/query", graphql_resolver.CorsMiddleware(x.Resolver.AuthMiddleware(&relay.Handler{Schema: x.Schema})))

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

func (x *CodeAmp) initGraphQL(resolver interface{}) {
	schema, err := assets.Asset("plugins/codeamp/schema.graphql")
	if err != nil {
		log.Panic(err)
	}

	parsedSchema, err := graphql.ParseSchema(string(schema), resolver)
	if err != nil {
		log.Panic(err)
	}

	x.Schema = parsedSchema
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

	x.Resolver = &graphql_resolver.Resolver{DB: x.DB, Events: x.Events, Redis: x.Redis}
	x.initGraphQL(x.Resolver)

	x.Events = events
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
	log.DebugWithFields("Processing CodeAmp event", log.Fields{"event": e})

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
