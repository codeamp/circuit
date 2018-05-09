package codeamp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	socketio "github.com/azhao1981/go-socket.io"
	sioredis "github.com/azhao1981/go-socket.io-redis"
	"github.com/codeamp/circuit/assets"
	"github.com/codeamp/circuit/plugins"
	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	redis "github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
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
	Resolver       *resolvers.Resolver
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
	x.Resolver = r
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
		"plugins.Release:create",
		"plugins.Release:complete",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	log.Info("Processing CodeAmp event")
	e.Dump()

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
	var environment resolvers.Environment
	var feature resolvers.Feature
	var projectSettings []resolvers.ProjectSettings

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

		if payload.Head {
			if x.DB.Where("continuous_deploy = ? and project_id = ?", true, project.Model.ID).Find(&projectSettings).RecordNotFound() {
				log.InfoWithFields("No continuous deploys found", log.Fields{
					"continuous_deploy": true,
					"project_id":        project.Model.ID,
				})
			} else {
				// call CreateRelease for each env that has cd turned on
				for _, setting := range projectSettings {
					if setting.ContinuousDeploy && fmt.Sprintf("refs/heads/%s", setting.GitBranch) == feature.Ref {
						adminContext := context.WithValue(context.Background(), "jwt", resolvers.Claims{
							UserID:      uuid.FromStringOrNil("codeamp").String(),
							Email:       "codeamp",
							Permissions: []string{"admin"},
						})

						x.Resolver.CreateRelease(adminContext, &struct{ Release *resolvers.ReleaseInput }{
							Release: &resolvers.ReleaseInput{
								HeadFeatureID: feature.Model.ID.String(),
								ProjectID:     setting.ProjectID.String(),
								EnvironmentID: setting.EnvironmentID.String(),
							},
						})
					}

					if x.DB.Where("id = ?", setting.EnvironmentID).First(&environment).RecordNotFound() {
						log.InfoWithFields("Environment not found", log.Fields{
							"id": setting.EnvironmentID,
						})
					}

					x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
						Event: fmt.Sprintf("projects/%s/%s/features", project.Slug, environment.Key),
					}, nil)
				}
			}
		}
	} else {
		log.InfoWithFields("feature already exists", log.Fields{
			"repository": payload.Repository,
			"hash":       payload.Hash,
		})
	}

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

		if payload.State == plugins.GetState("complete") {
			if len(e.Artifacts) > 0 {
				marshalledReArtifacts, err := json.Marshal(e.Artifacts)
				if err != nil {
					log.Info(err.Error(), log.Fields{})
				}
				extension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
			}
		}

		x.DB.Save(&extension)

		x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/%s/extensions", project.Slug, payload.Environment),
			Payload: extension,
		}, nil)
	}

	return nil
}

func (x *CodeAmp) ReleaseEventHandler(e transistor.Event) error {
	var err error
	payload := e.Payload.(plugins.Release)
	release := resolvers.Release{}
	releaseExtensions := []resolvers.ReleaseExtension{}

	if x.DB.Where("id = ?", payload.ID).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"id": payload.ID,
		})
		return fmt.Errorf("release %s not found", payload.ID)
	}

	if e.Matches("plugins.Release:create") {
		x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions)

		for _, releaseExtension := range releaseExtensions {
			projectExtension := resolvers.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.InfoWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return fmt.Errorf("project extension %s not found", releaseExtension.ProjectExtensionID)
			}

			extension := resolvers.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.InfoWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return fmt.Errorf("extension %s not found", projectExtension.ExtensionID)
			}

			if plugins.Type(extension.Type) == plugins.GetType("workflow") {
				// check if the last release extension has the same
				// ServicesSignature and SecretsSignature. If so,
				// mark the action as completed before sending the event
				lastReleaseExtension := resolvers.ReleaseExtension{}
				artifacts := []transistor.Artifact{}

				eventAction := plugins.GetAction("create")
				eventState := plugins.GetState("waiting")

				// check if can cache workflows
				if !release.ForceRebuild && !x.DB.Where("project_extension_id = ? and services_signature = ? and secrets_signature = ? and feature_hash = ? and state in (?)", projectExtension.Model.ID, releaseExtension.ServicesSignature, releaseExtension.SecretsSignature, releaseExtension.FeatureHash, []string{"complete", "failed"}).Order("created_at desc").First(&lastReleaseExtension).RecordNotFound() {
					eventAction = plugins.GetAction("status")
					eventState = lastReleaseExtension.State

					err := json.Unmarshal(lastReleaseExtension.Artifacts.RawMessage, &artifacts)
					if err != nil {
						log.Info(err.Error())
					}
				} else {
					artifacts, err = resolvers.ExtractArtifacts(projectExtension, extension, x.DB)
					if err != nil {
						log.Info(err.Error())
					}
				}

				ev := transistor.NewEvent(plugins.ReleaseExtension{
					ID:      releaseExtension.Model.ID.String(),
					Action:  eventAction,
					Slug:    extension.Key,
					State:   eventState,
					Release: payload,
				}, nil)

				ev.Artifacts = artifacts
				x.Events <- ev
			}
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
			return fmt.Errorf("Release %s not found", payload.Release.ID)
		}

		if x.DB.Where("id = ?", payload.ID).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.ID,
			})
			return fmt.Errorf("Release extension %s not found", payload.ID)
		}

		releaseExtension.State = payload.State
		releaseExtension.StateMessage = payload.StateMessage
		marshalledReArtifacts, err := json.Marshal(e.Artifacts)
		if err != nil {
			log.Info(err.Error(), log.Fields{})
			return err
		}

		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		x.DB.Save(&releaseExtension)

		if payload.State == plugins.GetState("complete") {
			x.ReleaseExtensionCompleted(&releaseExtension)
		}

		if payload.State == plugins.GetState("failed") {
			x.ReleaseFailed(&release, payload.StateMessage)
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
	environment := resolvers.Environment{}
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

	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("Environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
	}

	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases/reCompleted", project.Slug, environment.Key),
		Payload: release,
	}, nil)

	// loop through and check if all same-type elease extensions are completed
	done := true
	for _, releaseExtension := range releaseExtensions {
		if releaseExtension.Type == re.Type && releaseExtension.State != plugins.GetState("complete") {
			done = false
		}
	}

	if done {
		switch re.Type {
		case plugins.GetType("workflow"):
			x.WorkflowReleaseExtensionsCompleted(&release)
		case plugins.GetType("deployment"):
			x.ReleaseCompleted(&release)
		}
	}
}

func (x *CodeAmp) WorkflowReleaseExtensionsCompleted(release *resolvers.Release) {
	project := resolvers.Project{}
	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": release.ProjectID,
		})
		return
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

	user := resolvers.User{}
	if x.DB.Where("id = ?", release.UserID).First(&user).RecordNotFound() {
		log.InfoWithFields("user not found", log.Fields{
			"id": release.UserID,
		})
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

	// insert CodeAmp envs
	slugSecret := plugins.Secret{
		Key:   "CODEAMP_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, slugSecret)

	hashSecret := plugins.Secret{
		Key:   "CODEAMP_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, hashSecret)

	timeSecret := plugins.Secret{
		Key:   "CODEAMP_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, timeSecret)

	// insert Codeflow envs - remove later
	_slugSecret := plugins.Secret{
		Key:   "CODEFLOW_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _slugSecret)

	_hashSecret := plugins.Secret{
		Key:   "CODEFLOW_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _hashSecret)

	_timeSecret := plugins.Secret{
		Key:   "CODEFLOW_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _timeSecret)

	releaseExtensionDeploymentsCount := 0
	releaseExtensions := []resolvers.ReleaseExtension{}
	artifacts := []transistor.Artifact{}

	x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions)
	for _, releaseExtension := range releaseExtensions {
		// collect workflow artifacts
		if releaseExtension.Type == plugins.GetType("workflow") {
			projectExtension := resolvers.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.InfoWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := resolvers.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.InfoWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			var unmarshalledArtifacts []transistor.Artifact
			err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &unmarshalledArtifacts)
			if err != nil {
				log.InfoWithFields(err.Error(), log.Fields{})
				return
			}

			for _, artifact := range unmarshalledArtifacts {
				artifact.Source = extension.Key
				artifacts = append(artifacts, artifact)
			}
		}
	}

	for _, releaseExtension := range releaseExtensions {
		releaseExtensionAction := plugins.Action("create")
		if releaseExtension.Type == plugins.GetType("deployment") {
			_artifacts := artifacts

			// Fail deployment if the release is in a failed state
			if release.State == plugins.GetState("failed") {
				releaseExtensionAction = plugins.GetAction("status")
				releaseExtension.State = plugins.GetState("failed")
				releaseExtension.StateMessage = release.StateMessage
			}

			projectExtension := resolvers.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.InfoWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := resolvers.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.InfoWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			projectExtensionArtifacts, err := resolvers.ExtractArtifacts(projectExtension, extension, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			for _, artifact := range projectExtensionArtifacts {
				_artifacts = append(_artifacts, artifact)
			}

			releaseExtensionEvent := plugins.ReleaseExtension{
				ID:     releaseExtension.Model.ID.String(),
				Action: releaseExtensionAction,
				Slug:   extension.Key,
				State:  releaseExtension.State,
				Release: plugins.Release{
					ID:          release.Model.ID.String(),
					State:       release.State,
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
					User: user.Email,
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
				},
				StateMessage: releaseExtension.StateMessage,
			}

			ev := transistor.NewEvent(releaseExtensionEvent, nil)
			ev.Artifacts = _artifacts
			x.Events <- ev
			releaseExtensionDeploymentsCount++
		}
	}

	if releaseExtensionDeploymentsCount == 0 {
		x.ReleaseCompleted(release)
	}
}

func (x *CodeAmp) RunQueuedReleases(release *resolvers.Release) error {
	var nextQueuedRelease resolvers.Release

	if x.DB.Where("id != ? and state = ? and project_id = ? and environment_id = ? and created_at > ?", release.Model.ID, "waiting", release.ProjectID, release.EnvironmentID, release.CreatedAt).Order("created_at asc").First(&nextQueuedRelease).RecordNotFound() {
		log.InfoWithFields("No queued releases found.", log.Fields{
			"id":             release.Model.ID,
			"state":          "waiting",
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
			"created_at":     release.CreatedAt,
		})
		return nil
	}

	var project resolvers.Project
	var services []resolvers.Service
	var secrets []resolvers.Secret

	projectSecrets := []resolvers.Secret{}
	// get all the env vars related to this release and store
	x.DB.Where("environment_id = ? AND project_id = ? AND scope = ?", nextQueuedRelease.EnvironmentID, nextQueuedRelease.ProjectID, "project").Find(&projectSecrets)
	for _, secret := range projectSecrets {
		var secretValue resolvers.SecretValue
		x.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	globalSecrets := []resolvers.Secret{}
	x.DB.Where("environment_id = ? AND scope = ?", nextQueuedRelease.EnvironmentID, "global").Find(&globalSecrets)
	for _, secret := range globalSecrets {
		var secretValue resolvers.SecretValue
		x.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	x.DB.Where("project_id = ? and environment_id = ?", nextQueuedRelease.ProjectID, nextQueuedRelease.EnvironmentID).Find(&services)
	if len(services) == 0 {
		log.InfoWithFields("no services found", log.Fields{
			"project_id": nextQueuedRelease.ProjectID,
		})
	}

	if x.DB.Where("id = ?", nextQueuedRelease.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	for i, service := range services {
		ports := []resolvers.ServicePort{}
		x.DB.Where("service_id = ?", service.Model.ID).Find(&ports)
		services[i].Ports = ports
	}

	if x.DB.Where("id = ?", nextQueuedRelease.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	// get all branches relevant for the project
	var branch string
	var projectSettings resolvers.ProjectSettings

	if x.DB.Where("environment_id = ? and project_id = ?", nextQueuedRelease.EnvironmentID, nextQueuedRelease.ProjectID).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var environment resolvers.Environment
	if x.DB.Where("id = ?", nextQueuedRelease.EnvironmentID).Find(&environment).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": nextQueuedRelease.EnvironmentID,
		})
		return nil
	}

	var headFeature resolvers.Feature
	if x.DB.Where("id = ?", nextQueuedRelease.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": nextQueuedRelease.HeadFeatureID,
		})
		return nil
	}

	var tailFeature resolvers.Feature
	if x.DB.Where("id = ?", nextQueuedRelease.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": nextQueuedRelease.TailFeatureID,
		})
		return nil
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec resolvers.ServiceSpec
		if x.DB.Where("id = ?", service.ServiceSpecID).First(&spec).RecordNotFound() {
			log.InfoWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecID,
			})
			return nil
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

	// insert CodeAmp envs
	slugSecret := plugins.Secret{
		Key:   "CODEAMP_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, slugSecret)

	hashSecret := plugins.Secret{
		Key:   "CODEAMP_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, hashSecret)

	timeSecret := plugins.Secret{
		Key:   "CODEAMP_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, timeSecret)

	// insert Codeflow envs - remove later
	_slugSecret := plugins.Secret{
		Key:   "CODEFLOW_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _slugSecret)

	_hashSecret := plugins.Secret{
		Key:   "CODEFLOW_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _hashSecret)

	_timeSecret := plugins.Secret{
		Key:   "CODEFLOW_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _timeSecret)

	releaseEvent := plugins.Release{
		ID:          nextQueuedRelease.Model.ID.String(),
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
		User: nextQueuedRelease.User.Email,
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
		Secrets: pluginSecrets,
	}

	x.Events <- transistor.NewEvent(releaseEvent, nil)
	return nil
}

func (x *CodeAmp) ReleaseFailed(release *resolvers.Release, stateMessage string) {
	release.State = plugins.GetState("failed")
	release.StateMessage = stateMessage
	x.DB.Save(release)

	x.RunQueuedReleases(release)
}

func (x *CodeAmp) ReleaseCompleted(release *resolvers.Release) {
	project := resolvers.Project{}
	environment := resolvers.Environment{}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("Environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
	}

	// mark release as complete
	release.State = plugins.GetState("complete")
	release.StateMessage = "Completed"

	x.DB.Save(release)

	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key),
		Payload: release,
	}, nil)

	x.RunQueuedReleases(release)
}
