package codeamp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	"github.com/jinzhu/gorm/dialects/postgres"
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

	db.LogMode(false)

	db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.Project{},
		&models.Release{},
		&models.Feature{},
		&models.GitBranch{},
		&models.Service{},
		&models.ContainerPort{},
		&models.ServiceSpec{},
		&models.ExtensionSpec{},
		&models.Extension{},
		&models.EnvironmentVariable{},
		&models.EnvironmentVariableValue{},
		&models.ReleaseExtension{},
		&models.Environment{},
		&models.EnvironmentBasedProjectBranch{},
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

	// loop through secrets
	secretsContent, err := ioutil.ReadFile("./configs/secrets.json")
	if err != nil {
		log.Info("Could not read file ./configs/secrets.json")
	}
	secretsMap := make(map[string]interface{})
	err = json.Unmarshal(secretsContent, &secretsMap)
	if err != nil {
		log.Info("Could not unmarshal config. Please look at your secrets to ensure valid JSON format.")
	}

	// hosted zone id
	// hosted zone name
	hostedZoneId := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneId, hostedZoneId)
	db.Save(&hostedZoneId)

	eValue := models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneId.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneIdDev := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneIdDev, hostedZoneIdDev)
	db.Save(&hostedZoneIdDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneIdDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// hosted zone name
	hostedZone := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZone, hostedZone)
	db.Save(&hostedZone)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZone.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneDev := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneDev, hostedZoneDev)
	db.Save(&hostedZoneDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws secret key
	awsSecretKey := models.EnvironmentVariable{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&awsSecretKey, awsSecretKey)
	db.Save(&awsSecretKey)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsSecretKey.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsSecretKeyDev := models.EnvironmentVariable{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&awsSecretKeyDev, awsSecretKeyDev)
	db.Save(&awsSecretKeyDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsSecretKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws access key
	awsAccessKey := models.EnvironmentVariable{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&productionEnv, productionEnv)
	db.Save(&productionEnv)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsAccessKey.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsAccessKeyDev := models.EnvironmentVariable{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKeyDev, awsAccessKeyDev)
	db.Save(&awsAccessKeyDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsAccessKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker org
	dockerOrg := models.EnvironmentVariable{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrg, dockerOrg)
	db.Save(&dockerOrg)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerOrg.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerOrgDev := models.EnvironmentVariable{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrgDev, dockerOrgDev)
	db.Save(&dockerOrgDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerOrgDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker host
	dockerHost := models.EnvironmentVariable{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHost, dockerHost)
	db.Save(&dockerHost)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerHost.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerHostDev := models.EnvironmentVariable{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHostDev, dockerHostDev)
	db.Save(&dockerHostDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerHostDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// s3 bucket
	s3Bucket := models.EnvironmentVariable{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&s3Bucket, s3Bucket)
	db.Save(&s3Bucket)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: s3Bucket.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	s3BucketDev := models.EnvironmentVariable{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&s3BucketDev, s3BucketDev)
	db.Save(&s3BucketDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: s3BucketDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// ssl arn
	sslArn := models.EnvironmentVariable{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&sslArn, sslArn)
	db.Save(&sslArn)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: sslArn.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	sslArnDev := models.EnvironmentVariable{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&sslArnDev, sslArnDev)
	db.Save(&sslArnDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: sslArnDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// kubeconfig
	kubeConfig := models.EnvironmentVariable{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfig, kubeConfig)
	db.Save(&kubeConfig)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: kubeConfig.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	kubeConfigDev := models.EnvironmentVariable{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfigDev, kubeConfigDev)
	db.Save(&kubeConfigDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: kubeConfigDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

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

	// docker builder
	dbConfig := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "ORG", "value": dockerOrg.Model.ID.String()},
		map[string]interface{}{"key": "HOST", "value": dockerHost.Model.ID.String()},
		map[string]interface{}{"key": "USER", "value": ""},
		map[string]interface{}{"key": "EMAIL", "value": ""},
		map[string]interface{}{"key": "PASSWORD", "value": ""},
	}
	marshalledDbConfig, err := json.Marshal(dbConfig)
	if err != nil {
		log.Info("could not marshal dockerbuilder config")
	}
	extensionSpec := models.ExtensionSpec{
		Type:          plugins.GetType("workflow"),
		Key:           "dockerbuilder",
		Name:          "Docker Builder",
		Component:     "",
		EnvironmentId: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledDbConfig},
	}
	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	dbConfig = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "ORG", "value": dockerOrgDev.Model.ID.String()},
		map[string]interface{}{"key": "HOST", "value": dockerHostDev.Model.ID.String()},
		map[string]interface{}{"key": "USER", "value": ""},
		map[string]interface{}{"key": "EMAIL", "value": ""},
		map[string]interface{}{"key": "PASSWORD", "value": ""},
	}
	marshalledDbConfig, err = json.Marshal(dbConfig)
	if err != nil {
		log.Info("could not marshal dockerbuilder config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("workflow"),
		Key:           "dockerbuilder",
		Name:          "Docker Builder",
		Component:     "",
		EnvironmentId: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledDbConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	// load balancer
	lbConfig := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArn.Model.ID.String()},
		map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3Bucket.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneId.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_NAME", "value": hostedZone.Model.ID.String()},
		map[string]interface{}{"key": "AWS_ACCESS_KEY_ID", "value": awsAccessKey.Model.ID.String()},
		map[string]interface{}{"key": "AWS_SECRET_KEY", "value": awsAccessKey.Model.ID.String()},
	}
	marshalledLbConfig, err := json.Marshal(lbConfig)
	if err != nil {
		log.Info("could not marshal loadbalancer config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("once"),
		Key:           "kubernetesloadbalancers",
		Name:          "Load Balancer",
		Component:     "LoadBalancer",
		EnvironmentId: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledLbConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	lbConfig = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArnDev.Model.ID.String()},
		map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3BucketDev.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneIdDev.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_NAME", "value": hostedZoneDev.Model.ID.String()},
		map[string]interface{}{"key": "AWS_ACCESS_KEY_ID", "value": awsAccessKeyDev.Model.ID.String()},
		map[string]interface{}{"key": "AWS_SECRET_KEY", "value": awsAccessKeyDev.Model.ID.String()},
	}
	marshalledLbConfig, err = json.Marshal(lbConfig)
	if err != nil {
		log.Info("could not marshal loadbalancer config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("once"),
		Key:           "kubernetesloadbalancers",
		Name:          "Load Balancer",
		Component:     "LoadBalancer",
		EnvironmentId: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledLbConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	// kubernetes
	kubeConfigSpec := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "USER", "value": ""},
		map[string]interface{}{"key": "PASSWORD", "value": ""},
		map[string]interface{}{"key": "EMAIL", "value": ""},
		map[string]interface{}{"key": "HOST", "value": dockerHost.Model.ID.String()},
	}
	marshalledKubeConfig, err := json.Marshal(kubeConfigSpec)
	if err != nil {
		log.Info("could not marshal kube config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("deployment"),
		Key:           "kubernetesdeployments",
		Name:          "Kubernetes",
		Component:     "",
		EnvironmentId: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledKubeConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	kubeConfigSpec = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "USER", "value": ""},
		map[string]interface{}{"key": "PASSWORD", "value": ""},
		map[string]interface{}{"key": "EMAIL", "value": ""},
		map[string]interface{}{"key": "HOST", "value": dockerHostDev.Model.ID.String()},
	}
	marshalledKubeConfig, err = json.Marshal(kubeConfigSpec)
	if err != nil {
		log.Info("could not marshal kube config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("deployment"),
		Key:           "kubernetesdeployments",
		Name:          "Kubernetes",
		Component:     "",
		EnvironmentId: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledKubeConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
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
	db.LogMode(false)

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
		"plugins.GitBranch",
		"plugins.GitStatus",
		"plugins.HeartBeat",
		"plugins.WebsocketMsg",
		"plugins.Extension:status",
		"plugins.Extension:update",
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

	if e.Matches("plugins.GitBranch") {
		payload := e.Payload.(plugins.GitBranch)
		x.Actions.GitBranch(payload)
	}

	if e.Matches("plugins.WebsocketMsg") {
		payload := e.Payload.(plugins.WebsocketMsg)

		if payload.Channel == "" {
			payload.Channel = "general"
		}

		x.SocketIO.BroadcastTo(payload.Channel, payload.Event, payload.Payload, nil)
	}

	if e.Matches("plugins.Extension:status") {
		payload := e.Payload.(plugins.Extension)
		var extension models.Extension

		if x.Db.Where("id = ?", payload.Id).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": payload.Id,
			})
			return nil
		}

		marshalledArtifacts, err := json.Marshal(payload.Artifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{})
			return nil
		}

		extension.State = plugins.GetState("complete")
		extension.Artifacts = postgres.Jsonb{marshalledArtifacts}

		x.Db.Save(&extension)

		if payload.State == plugins.GetState("complete") {
			x.Actions.ExtensionInitCompleted(&extension)
		}
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

		marshalledArtifacts, err := json.Marshal(payload.Artifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{})
			return nil
		}

		extension.State = plugins.GetState("complete")
		extension.Artifacts = postgres.Jsonb{marshalledArtifacts}
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
		marshalledReArtifacts, err := json.Marshal(payload.Artifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{})
			return nil
		}

		releaseExtension.State = payload.State
		releaseExtension.StateMessage = payload.StateMessage
		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		// releaseExtension.Artifacts = plugins.MapStringStringToHstore(payload.Artifacts)
		x.Db.Save(&releaseExtension)

		if payload.State == plugins.GetState("complete") {
			x.Actions.ReleaseExtensionCompleted(&releaseExtension)
		}

		if payload.State == plugins.GetState("failed") {
			var release models.Release

			if x.Db.Where("id = ?", payload.Release.Id).Find(&release).RecordNotFound() {
				log.InfoWithFields("release", log.Fields{
					"id": payload.Release.Id,
				})
				return nil
			}
			release.State = plugins.GetState("failed")
			release.StateMessage = payload.StateMessage
			x.Db.Save(&release)
		}
	}

	return nil
}
