package graphql

import (
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type GraphQL struct {
	Events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("graphql", func() transistor.Plugin {
		return &GraphQL{}
	}, plugins.Project{})
}

func (x *GraphQL) GraphQLListen() {
	// x.SocketIO.On("connection", func(so socketio.Socket) {
	// 	so.Join("general")
	// })

	// x.SocketIO.On("error", func(so socketio.Socket, err error) {
	// 	log.Println("socket-io error:", err)
	// })

	// sIOServer := new(socketIOServer)
	// sIOServer.Server = x.SocketIO
	// http.Handle("/socket.io/", sIOServer)

	// _, filename, _, _ := runtime.Caller(0)
	// fs := http.FileServer(http.Dir(path.Join(path.Dir(filename), "static/")))
	// http.Handle("/", fs)

	// http.Handle("/query", resolver.CorsMiddleware(x.Resolver.AuthMiddleware(&relay.Handler{Schema: x.Schema})))

	// log.Info(fmt.Sprintf("Running GraphQL server on %v", x.ServiceAddress))
	// log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *GraphQL) Start(events chan transistor.Event) error {
	x.Events = events

	log.Info("Starting GraphQL service")
	return nil
}

func (x *GraphQL) Stop() {
	log.Info("Stopping GraphQL service")
}

func (x *GraphQL) Subscribe() []string {
	return []string{
		"gitsync:status",
		"heartbeat",
		"websocket",
		"project",
		"release",
	}
}

func (x *GraphQL) Process(e transistor.Event) error {
	log.DebugWithFields("Processing GraphQL event", log.Fields{"event": e})

	return nil
}
