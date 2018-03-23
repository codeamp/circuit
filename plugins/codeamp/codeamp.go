package codeamp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/assets"
	"github.com/codeamp/circuit/plugins"
	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	redis "github.com/go-redis/redis"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	sioredis "github.com/satyakb/go-socket.io-redis"
	"github.com/spf13/viper"
)

func init() {
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return NewCodeAmp() })
}

type CodeAmp struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	Schema         *graphql.Schema
	SocketIO       *socketio.Server
	DB             *gorm.DB
	Redis          *redis.Client
}

func NewCodeAmp() *CodeAmp {
	return &CodeAmp{}
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

	r := &resolvers.Resolver{DB: x.DB, Events: x.Events, Redis: x.Redis}
	http.Handle("/query", resolvers.CorsMiddleware(r.AuthMiddleware(&relay.Handler{Schema: x.Schema})))

	log.Info(fmt.Sprintf("running GraphQL server on %v", x.ServiceAddress))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *CodeAmp) Start(events chan transistor.Event) error {
	var err error

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	//defer x.DB.Close()

	schema, err := assets.Asset("plugins/codeamp/schema.graphql")
	if err != nil {
		log.Panic(err)
	}

	parsedSchema, err := graphql.ParseSchema(string(schema), &resolvers.Resolver{DB: db, Events: events})
	if err != nil {
		log.Panic(err)
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

	x.Events = events
	x.SocketIO = sio
	x.Schema = parsedSchema
	x.Redis = redisClient

	// DEBUG
	db.LogMode(false)

	x.DB = db

	go x.Listen()

	return nil
}

func (x *CodeAmp) Stop() {
	log.Info("stopping CodeAmp service")
}

func (x *CodeAmp) Subscribe() []string {
	return []string{
		"plugins.GitCommit",
		"plugins.HeartBeat",
		"plugins.WebsocketMsg",
		"plugins.ProjectExtension:status",
		"plugins.ProjectExtension:update",
		"plugins.ProjectExtension:complete",
		"plugins.ReleaseExtension:status",
		"plugins.Release:status",
		"plugins.Release:complete",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	log.InfoWithFields("process CodeAmp event", log.Fields{
		"event_name": e.Name,
	})

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

func (x *CodeAmp) HeartBeatEventHandler(e transistor.Event) {
	payload := e.Payload.(plugins.HeartBeat)

	var projects []resolvers.Project

	x.DB.Find(&projects)
	for _, project := range projects {
		switch payload.Tick {
		case "minute":
			x.GitSync(&project)
		}
	}
}

func (x *CodeAmp) GitCommitEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.GitCommit)

	var project resolvers.Project
	var feature resolvers.Feature

	if x.DB.Where("repository = ?", payload.Repository).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"repository": payload.Repository,
		})
		return nil
	}

	if x.DB.Where("project_id = ? AND hash = ?", project.ID, payload.Hash).First(&feature).RecordNotFound() {
		feature = resolvers.Feature{
			ProjectID:  project.ID,
			Message:    payload.Message,
			User:       payload.User,
			Hash:       payload.Hash,
			ParentHash: payload.ParentHash,
			Ref:        payload.Ref,
			Created:    payload.Created,
		}
		x.DB.Save(&feature)
	} else {
		log.InfoWithFields("feature already exists", log.Fields{
			"repository": payload.Repository,
			"hash":       payload.Hash,
		})
	}

	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event: fmt.Sprintf("projects/%s/features", project.Slug),
	}, nil)

	return nil
}

func (x *CodeAmp) WebsocketMsgEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.WebsocketMsg)

	if payload.Channel == "" {
		payload.Channel = "general"
	}

	x.SocketIO.BroadcastTo(payload.Channel, payload.Event, payload.Payload, nil)

	return nil
}

func (x *CodeAmp) ProjectExtensionEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.ProjectExtension)
	var extension resolvers.ProjectExtension
	var project resolvers.Project

	if e.Matches("plugins.ProjectExtension:status") {
		if x.DB.Where("id = ?", payload.ID).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": payload.ID,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because ProjectExtension not found given payload id: %s.", payload.ID))
		}

		if x.DB.Where("id = ?", extension.ProjectID).Find(&project).RecordNotFound() {
			log.InfoWithFields("project not found", log.Fields{
				"id": extension.ProjectID,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because Project not found given payload id: %s.", extension.ProjectID))
		}

		extension.State = payload.State
		extension.StateMessage = payload.StateMessage
		//extension.Artifacts = postgres.Jsonb{marshalledArtifacts}
		x.DB.Save(&extension)

		x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/extensions", project.Slug),
			Payload: extension,
		}, nil)

		if payload.State == plugins.GetState("complete") {
			// TODO: handle extension init complete
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.ReleaseExtension)

	var releaseExtension resolvers.ReleaseExtension
	var release resolvers.Release

	if e.Matches("plugins.ReleaseExtension:status") {
		if x.DB.Where("id = ?", payload.Release.ID).Find(&release).RecordNotFound() {
			log.InfoWithFields("release", log.Fields{
				"id": payload.Release.ID,
			})
			return nil
		}

		if x.DB.Where("id = ?", payload.ID).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.ID,
			})
			return nil
		}

		releaseExtension.State = payload.State
		releaseExtension.StateMessage = payload.StateMessage
		marshalledReArtifacts, err := json.Marshal(payload.Artifacts)
		if err != nil {
			log.Info(err.Error(), log.Fields{})
		}

		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		x.DB.Save(&releaseExtension)

		if payload.State == plugins.GetState("complete") {
			// append release extension artifacts to Release
			mergedArtifacts := make(map[string]interface{})
			err = json.Unmarshal(release.Artifacts.RawMessage, &mergedArtifacts)
			if err != nil {
				log.Info(err.Error())
				return err
			}

			if len(mergedArtifacts) > 0 {
				for key, value := range payload.Artifacts {
					mergedArtifacts[key] = value
				}
			} else {
				mergedArtifacts = payload.Artifacts
			}

			marshalledArtifacts, err := json.Marshal(mergedArtifacts)
			if err != nil {
				log.InfoWithFields(err.Error(), log.Fields{})
				return err
			}

			release.Artifacts = postgres.Jsonb{marshalledArtifacts}
			x.DB.Save(&release)

			x.ReleaseExtensionCompleted(&releaseExtension)
		}

		if payload.State == plugins.GetState("failed") {
			release.State = plugins.GetState("failed")
			release.StateMessage = payload.StateMessage
			x.DB.Save(&release)
		}
	}

	return nil
}

func (x *CodeAmp) GitSync(project *resolvers.Project) error {
	var feature resolvers.Feature
	var release resolvers.Release
	var headFeature resolvers.Feature
	hash := ""

	// Get latest release and deployed feature hash
	if x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&release).RecordNotFound() {
		// get latest feature if there is no releases
		x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&feature)
		hash = feature.Hash
	} else {
		if x.DB.Where("id = ?", release.HeadFeatureID).Find(&headFeature).RecordNotFound() {
			log.InfoWithFields("can not find head feature", log.Fields{
				"id": release.HeadFeatureID,
			})
		}
		hash = headFeature.Hash
	}

	// get branches of entire environments
	projectSettingsCollection := []resolvers.ProjectSettings{}
	if x.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectSettingsCollection).RecordNotFound() {
		gitSync := plugins.GitSync{
			Action: plugins.GetAction("update"),
			State:  plugins.GetState("waiting"),
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Repository: project.Repository,
			},
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        "master",
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			From: hash,
		}

		x.Events <- transistor.NewEvent(gitSync, nil)
	} else {
		for _, projectSettings := range projectSettingsCollection {
			gitSync := plugins.GitSync{
				Action: plugins.GetAction("update"),
				State:  plugins.GetState("waiting"),
				Project: plugins.Project{
					ID:         project.Model.ID.String(),
					Repository: project.Repository,
				},
				Git: plugins.Git{
					Url:           project.GitUrl,
					Protocol:      project.GitProtocol,
					Branch:        projectSettings.GitBranch,
					RsaPrivateKey: project.RsaPrivateKey,
					RsaPublicKey:  project.RsaPublicKey,
				},
				From: hash,
			}

			x.Events <- transistor.NewEvent(gitSync, nil)
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionCompleted(re *resolvers.ReleaseExtension) {
	project := resolvers.Project{}
	release := resolvers.Release{}
	releaseExtensions := []resolvers.ReleaseExtension{}

	if x.DB.Where("id = ?", re.ReleaseID).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	if x.DB.Where("release_id = ?", re.ReleaseID).Find(&releaseExtensions).RecordNotFound() {
		log.InfoWithFields("release extensions not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/releases/reCompleted", project.Slug),
		Payload: release,
	}, nil)

	// loop through and check if all same-type elease extensions are completed
	done := true
	for _, fre := range releaseExtensions {
		if fre.Type == re.Type && fre.State != plugins.GetState("complete") {
			done = false
		}
	}

	if done {
		switch re.Type {
		case plugins.GetType("workflow"):
			x.WorkflowProjectExtensionsCompleted(&release)
		case plugins.GetType("deployment"):
			x.DeploymentProjectExtensionsCompleted(&release)
		}
	}
}

func (x *CodeAmp) WorkflowProjectExtensionsCompleted(release *resolvers.Release) {
	// find all related deployment extensions
	depProjectExtensions := []resolvers.ProjectExtension{}
	aggregateReleaseExtensionArtifacts := make(map[string]interface{})
	found := false

	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectID, release.EnvironmentID).Find(&depProjectExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
	}

	for _, de := range depProjectExtensions {
		var ext resolvers.Extension
		if x.DB.Where("id = ?", de.ExtensionID).First(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension spec": de,
			})
		}
		if ext.Type == plugins.GetType("workflow") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND project_extension_id = ? AND state = ?", release.Model.ID, de.Model.ID, string(plugins.GetState("complete"))).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":           release.Model.ID,
					"project_extension_id": de.Model.ID,
					"state":                plugins.GetState("complete"),
				})
			}

			// put all releaseextension artifacts inside release artifacts
			unmarshalledArtifacts := make(map[string]interface{})
			err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &unmarshalledArtifacts)
			if err != nil {
				log.InfoWithFields(err.Error(), log.Fields{})
				return
			}

			for k, v := range unmarshalledArtifacts {
				key := fmt.Sprintf("%s_%s", strings.ToUpper(ext.Key), strings.ToUpper(k))
				aggregateReleaseExtensionArtifacts[key] = v
			}
		}

		if ext.Type == plugins.GetType("deployment") {
			found = true
		}
	}

	// persist workflow artifacts
	// release.Artifacts = plugins.MapStringStringToHstore(releaseExtensionArtifacts)
	// x.DB.Save(&release)

	// if there are no deployment workflows, then release is complete
	if !found {
		x.ReleaseCompleted(release)
		return
	}

	project := resolvers.Project{}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	headFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := resolvers.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	projectSettings := resolvers.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var secrets []resolvers.Secret
	err := json.Unmarshal(release.Secrets.RawMessage, &secrets)
	if err != nil {
		log.Info(err.Error(), log.Fields{})
		return
	}

	var services []resolvers.Service
	err = json.Unmarshal(release.Services.RawMessage, &services)
	if err != nil {
		log.Info(err.Error(), log.Fields{})
		return
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec resolvers.ServiceSpec
		if x.DB.Where("id = ?", service.ServiceSpecID).First(&spec).RecordNotFound() {
			log.InfoWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecID,
			})
			return
		}

		count, _ := strconv.ParseInt(service.Count, 10, 64)
		terminationGracePeriod, _ := strconv.ParseInt(spec.TerminationGracePeriod, 10, 64)

		pluginServices = append(pluginServices, plugins.Service{
			ID:        service.Model.ID.String(),
			Action:    plugins.GetAction("create"),
			State:     plugins.GetState("waiting"),
			Name:      service.Name,
			Command:   service.Command,
			Listeners: []plugins.Listener{},
			Replicas:  count,
			Spec: plugins.ServiceSpec{
				ID:                            spec.Model.ID.String(),
				CpuRequest:                    fmt.Sprintf("%sm", spec.CpuRequest),
				CpuLimit:                      fmt.Sprintf("%sm", spec.CpuLimit),
				MemoryRequest:                 fmt.Sprintf("%sMi", spec.MemoryRequest),
				MemoryLimit:                   fmt.Sprintf("%sMi", spec.MemoryLimit),
				TerminationGracePeriodSeconds: terminationGracePeriod,
			},
			Type: string(service.Type),
		})
	}

	var pluginSecrets []plugins.Secret
	for _, secret := range secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
	}

	releaseEvent := plugins.Release{
		Action:       plugins.GetAction("create"),
		State:        plugins.GetState("waiting"),
		Environment:  environment.Key,
		Artifacts:    aggregateReleaseExtensionArtifacts,
		StateMessage: "create release event",
		ID:           release.Model.ID.String(),
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
		User: "",
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
		Secrets:  pluginSecrets,
		Services: pluginServices,
	}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range depProjectExtensions {
		ext := resolvers.Extension{}
		if x.DB.Where("id= ?", extension.ExtensionID).Find(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension": extension,
			})
		}

		if plugins.Type(ext.Type) == plugins.GetType("workflow") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, extension.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": extension.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}
		}

		if plugins.Type(ext.Type) == plugins.GetType("deployment") {

			// create ReleaseExtension
			releaseExtension := resolvers.ReleaseExtension{
				ReleaseID:          release.Model.ID,
				FeatureHash:        "",
				ServicesSignature:  "",
				SecretsSignature:   "",
				ProjectExtensionID: extension.Model.ID,
				State:              plugins.GetState("waiting"),
				Type:               plugins.GetType("deployment"),
				StateMessage:       "initialized",
			}

			x.DB.Save(&releaseExtension)
			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			config, err := resolvers.ExtractConfig(unmarshalledConfig, ext.Key, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			eventAction := plugins.GetAction("create")
			eventState := plugins.GetState("waiting")
			artifacts := map[string]interface{}{}

			// check if the last release extension has the same
			// ServicesSignature and SecretsSignature. If so,
			// mark the action as completed before sending the event
			lastReleaseExtension := resolvers.ReleaseExtension{}

			x.DB.Where("project_extension_id = ? and services_signature = ? and secrets_signature = ?", releaseExtension.ProjectExtensionID, releaseExtension.ServicesSignature,
				releaseExtension.SecretsSignature).Order("created_at desc").First(&lastReleaseExtension)

			if lastReleaseExtension.Model.ID.String() != "" {
				eventAction = plugins.GetAction("status")
				eventState = lastReleaseExtension.State
				err := json.Unmarshal(lastReleaseExtension.Artifacts.RawMessage, &artifacts)
				if err != nil {
					log.Info(err.Error())
				}
			}

			releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
				ID:           releaseExtension.Model.ID.String(),
				Action:       eventAction,
				Slug:         ext.Key,
				State:        eventState,
				Config:       config,
				Artifacts:    map[string]interface{}{},
				Release:      releaseEvent,
				StateMessage: releaseExtension.StateMessage,
			})

		}
	}

	// send out release extension event for each re
	for _, re := range releaseExtensionEvents {
		//re.Release.Artifacts = aggregateReleaseExtensionArtifacts
		x.Events <- transistor.NewEvent(re, nil)
	}
}

func (x *CodeAmp) DeploymentProjectExtensionsCompleted(release *resolvers.Release) {
	// find all related deployment extensions
	depProjectExtensions := []resolvers.ProjectExtension{}
	// releaseExtensionArtifacts := map[string]string{}

	if x.DB.Where("project_id = ?", release.ProjectID).Find(&depProjectExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
		return
	}

	for _, de := range depProjectExtensions {
		var ext resolvers.Extension
		if x.DB.Where("id = ?", de.ExtensionID).First(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": de.ExtensionID,
			})
		}

		if ext.Type == plugins.GetType("deployment") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND project_extension_id = ? AND state = ?", release.Model.ID, de.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":           release.Model.ID,
					"project_extension_id": de.Model.ID,
					"state":                plugins.GetState("complete"),
				})
			}
		}
	}

	x.ReleaseCompleted(release)
}

func (x *CodeAmp) ReleaseCompleted(release *resolvers.Release) {
	project := resolvers.Project{}
	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	// mark release as complete
	release.State = plugins.GetState("complete")
	release.StateMessage = "Completed"

	x.DB.Save(release)
}

func (x *CodeAmp) ReleaseCreated(release *resolvers.Release) {
	project := resolvers.Project{}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	// loop through extensions and send ReleaseWorkflow events
	projectExtensions := []resolvers.ProjectExtension{}
	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectID, release.EnvironmentID).Find(&projectExtensions).RecordNotFound() {
		log.InfoWithFields("project has no extensions", log.Fields{
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
		})
	}

	services := []resolvers.Service{}
	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectID, release.EnvironmentID).Find(&services).RecordNotFound() {
		log.InfoWithFields("project has no services", log.Fields{
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
		})
	}

	headFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := resolvers.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	unmarshalledArtifacts := make(map[string]interface{})
	err := json.Unmarshal(release.Artifacts.RawMessage, &unmarshalledArtifacts)
	if err != nil {
		log.Info(err.Error(), log.Fields{})
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	projectSettings := resolvers.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	releaseEvent := plugins.Release{
		ID:          release.Model.ID.String(),
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		Environment: environment.Key,
		HeadFeature: plugins.Feature{
			ID:         headFeature.Model.ID.String(),
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
		User: release.User.Email,
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
		Artifacts: unmarshalledArtifacts,
	}
	for _, extension := range projectExtensions {
		ext := resolvers.Extension{}
		if x.DB.Where("id= ?", extension.ExtensionID).Find(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": extension.ExtensionID,
			})
		}

		// ONLY SEND WORKFLOW TYPE, EVENTs
		if plugins.Type(ext.Type) == plugins.GetType("workflow") {
			// create ReleaseExtension
			releaseExtension := resolvers.ReleaseExtension{
				ReleaseID:          release.Model.ID,
				FeatureHash:        "",
				ServicesSignature:  "",
				SecretsSignature:   "",
				ProjectExtensionID: extension.Model.ID,
				State:              plugins.GetState("waiting"),
				Type:               plugins.GetType("workflow"),
			}

			x.DB.Save(&releaseExtension)

			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			config, err := resolvers.ExtractConfig(unmarshalledConfig, ext.Key, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			x.Events <- transistor.NewEvent(plugins.ReleaseExtension{
				ID:        releaseExtension.Model.ID.String(),
				Action:    plugins.GetAction("create"),
				Slug:      ext.Key,
				State:     releaseExtension.State,
				Release:   releaseEvent,
				Config:    config,
				Artifacts: map[string]interface{}{},
			}, nil)
		}
	}

	// send web socket message notifying release has been created
	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/releases", project.Slug),
		Payload: release,
	}, nil)
}
