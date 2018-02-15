package resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/scalar"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

func (r *Resolver) Release(ctx context.Context, args *struct{ ID graphql.ID }) *ReleaseResolver {
	release := models.Release{}
	return &ReleaseResolver{db: r.db, Release: release}
}

type ReleaseResolver struct {
	db      *gorm.DB
	Release models.Release
}

type ReleaseInput struct {
	ID            *string
	ProjectId     string
	HeadFeatureId string
	EnvironmentId string
}

func createSnapshot(db *gorm.DB, args *struct{ Release *ReleaseInput }) (interface{}, error) {
	snapshot := map[string]interface{}{
		"environmentVariables": []map[string]interface{}{},
		"services":             []map[string]interface{}{},
	}

	// get all the env vars related to this release and store
	projectEnvVars := []models.EnvironmentVariable{}
	if db.Where("environment_id = ? and project_id = ? and scope = ?", args.Release.EnvironmentId, args.Release.ProjectId, "project").Find(&projectEnvVars).RecordNotFound() {
		log.InfoWithFields("no project env vars found", log.Fields{
			"environment_id": args.Release.EnvironmentId,
			"project_id":     args.Release.ProjectId,
			"scope":          "project",
		})
	}

	for _, val := range projectEnvVars {
		evValue := models.EnvironmentVariableValue{}
		if db.Where("environment_variable_id = ?", val.Model.ID.String()).Order("created_at desc").First(&evValue).RecordNotFound() {
			log.InfoWithFields("envvar value not found", log.Fields{
				"id": val.Model.ID.String(),
			})
		} else {
			snapshot["environmentVariables"] = append(snapshot["environmentVariables"].([]map[string]interface{}), map[string]interface{}{
				"value":     evValue.Value,
				"type":      val.Type,
				"key":       val.Key,
				"is_secret": val.IsSecret,
			})
		}
	}

	// get admin env vars
	adminEnvVars := []models.EnvironmentVariable{}
	if db.Where("scope = ?", "global").Find(&projectEnvVars).RecordNotFound() {
		log.InfoWithFields("no admin env vars found", log.Fields{
			"scope": "global",
		})
	}
	for _, val := range adminEnvVars {
		evValue := models.EnvironmentVariableValue{}
		if db.Where("environment_variable_id = ?", val.Model.ID.String()).Order("created_at desc").First(&evValue).RecordNotFound() {
			log.InfoWithFields("envvar value not found", log.Fields{
				"id": val.Model.ID.String(),
			})
		} else {
			snapshot["environmentVariables"] = append(snapshot["environmentVariables"].([]map[string]interface{}), map[string]interface{}{
				"value":     evValue.Value,
				"type":      val.Type,
				"key":       val.Key,
				"is_secret": val.IsSecret,
			})
		}
	}

	// get all services
	services := []models.Service{}
	if db.Where("project_id = ? and environment_id = ?", args.Release.ProjectId, args.Release.EnvironmentId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found", log.Fields{
			"project_id": args.Release.ProjectId,
		})
		return nil, fmt.Errorf("no services found")
	}

	env := models.Environment{}
	if db.Where("id = ?", args.Release.EnvironmentId).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.Release.EnvironmentId,
		})
		return nil, fmt.Errorf("no env found")
	}

	for _, service := range services {
		serviceSpec := models.ServiceSpec{}
		if db.Where("id = ?", service.ServiceSpecId).Find(&serviceSpec).RecordNotFound() {
			log.InfoWithFields("no service spec found", log.Fields{
				"id": service.ServiceSpecId,
			})
			return nil, fmt.Errorf("no service spec found")
		}
		serviceMap := map[string]interface{}{
			"id": service.Model.ID.String(),
			"service_spec": map[string]interface{}{
				"id":                       serviceSpec.Model.ID.String(),
				"name":                     serviceSpec.Name,
				"cpu_request":              serviceSpec.CpuRequest,
				"cpu_limit":                serviceSpec.CpuLimit,
				"memory_request":           serviceSpec.MemoryRequest,
				"memory_limit":             serviceSpec.MemoryLimit,
				"termination_grace_period": serviceSpec.TerminationGracePeriod,
			},
			"container_ports": []map[string]interface{}{},
			"command":         service.Command,
			"name":            service.Name,
			"type":            service.Type,
			"count":           service.Count,
			"environment": map[string]interface{}{
				"name": env.Name,
			},
		}

		listeners := []models.ContainerPort{}
		if db.Where("service_id = ?", service.Model.ID).Find(&listeners).RecordNotFound() {
			log.InfoWithFields("container ports not found", log.Fields{
				"service_id": service.Model.ID,
			})
		}
		for _, listener := range listeners {
			serviceMap["container_ports"] = append(serviceMap["container_ports"].([]map[string]interface{}), map[string]interface{}{
				"port":     listener.Port,
				"protocol": listener.Protocol,
			})
		}
		snapshot["services"] = append(snapshot["services"].([]map[string]interface{}), serviceMap)
	}
	return snapshot, nil
}

func (r *Resolver) RollbackRelease(ctx context.Context, args *struct{ ReleaseId graphql.ID }) (*ReleaseResolver, error) {
	/*
		Rollback's purpose is to deploy a feature with a previous configuration state of the project.
		We find the corresponding release object, get the Snapshot var to get the configuration of the project at the moment
		the release was created. We then create a new release object and insert the old release's info into the new release.
	*/
	release := models.Release{}
	if r.db.Where("id = ?", string(args.ReleaseId)).Find(&release).RecordNotFound() {
		errMsg := fmt.Sprintf("Could not find release with given id %s", string(args.ReleaseId))
		log.Info(errMsg)
		return nil, fmt.Errorf(errMsg)
	}
	// create new release object with snapshot from found release
	newRelease := models.Release{
		ProjectId:     release.ProjectId,
		EnvironmentId: release.EnvironmentId,
		UserID:        release.UserID,
		HeadFeatureID: release.HeadFeatureID,
		TailFeatureID: release.TailFeatureID,
		State:         plugins.GetState("waiting"),
		StateMessage:  "Release created and rolled back.",
		Snapshot:      release.Snapshot,
	}
	r.db.Create(&newRelease)
	r.actions.ReleaseCreated(&newRelease)

	return &ReleaseResolver{db: r.db, Release: release}, nil
}

func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *ReleaseInput }) (*ReleaseResolver, error) {
	var tailFeatureId uuid.UUID
	var currentRelease models.Release

	// only filled if ID is passed in args
	// used during redeploy fn
	var releaseFromId models.Release

	if args.Release.ID != nil {
		if r.db.Where("id = ?", *args.Release.ID).Find(&releaseFromId).RecordNotFound() {
			errMsg := fmt.Sprintf("release from id does not exist %s", args.Release.ID)
			return &ReleaseResolver{}, fmt.Errorf(errMsg)
		} else {
			snapshot, err := createSnapshot(r.db, args)
			if err != nil {
				return &ReleaseResolver{}, err
			}

			marshalledSnapshot, err := json.Marshal(snapshot)
			if err != nil {
				log.Info(err.Error())
				return nil, err
			}

			forkedRelease := models.Release{
				ProjectId:     releaseFromId.ProjectId,
				EnvironmentId: releaseFromId.EnvironmentId,
				UserID:        releaseFromId.UserID,
				HeadFeatureID: releaseFromId.HeadFeatureID,
				TailFeatureID: releaseFromId.TailFeatureID,
				State:         plugins.GetState("waiting"),
				StateMessage:  "Release created",
				Snapshot:      postgres.Jsonb{marshalledSnapshot},
			}
			r.db.Create(&forkedRelease)
			r.actions.ReleaseCreated(&forkedRelease)
			return &ReleaseResolver{db: r.db, Release: forkedRelease}, nil
		}
	} else {
		projectId, err := uuid.FromString(args.Release.ProjectId)
		if err != nil {
			log.InfoWithFields("Couldn't parse projectId", log.Fields{
				"args": args,
			})
			return nil, fmt.Errorf("Couldn't parse projectId")
		}
		headFeatureId, err := uuid.FromString(args.Release.HeadFeatureId)
		if err != nil {
			log.InfoWithFields("Couldn't parse headFeatureId", log.Fields{
				"args": args,
			})
			return nil, fmt.Errorf("Couldn't parse headFeatureId")
		}
		environmentId, err := uuid.FromString(args.Release.EnvironmentId)
		if err != nil {
			log.InfoWithFields("Couldn't parse environmentId", log.Fields{
				"args": args,
			})
			return nil, fmt.Errorf("Couldn't parse environmentId")
		}

		// the tail feature id is the current release's head feature id
		if r.db.Where("state = ? and project_id = ? and environment_id = ?", plugins.GetState("complete"), args.Release.ProjectId, environmentId).Find(&currentRelease).Order("created_at desc").Limit(1).RecordNotFound() {
			// get first ever feature in project if current release doesn't exist yet
			var firstFeature models.Feature
			if r.db.Where("project_id = ?", args.Release.ProjectId).Find(&firstFeature).Order("created_at asc").Limit(1).RecordNotFound() {
				log.InfoWithFields("CreateRelease", log.Fields{
					"release": r,
				})
				return nil, fmt.Errorf("No features found.")
			}
			tailFeatureId = firstFeature.ID
		} else {
			tailFeatureId = currentRelease.HeadFeatureID
		}

		userIdString, err := utils.CheckAuth(ctx, []string{})
		if err != nil {
			return &ReleaseResolver{}, err
		}
		userId := uuid.FromStringOrNil(userIdString)

		snapshot, err := createSnapshot(r.db, args)
		if err != nil {
			return &ReleaseResolver{}, err
		}

		marshalledSnapshot, err := json.Marshal(snapshot)
		if err != nil {
			log.Info(err.Error())
			return nil, err
		}

		release := models.Release{
			ProjectId:     projectId,
			EnvironmentId: environmentId,
			UserID:        userId,
			HeadFeatureID: headFeatureId,
			TailFeatureID: tailFeatureId,
			State:         plugins.GetState("waiting"),
			StateMessage:  "Release created",
			Snapshot:      postgres.Jsonb{marshalledSnapshot},
		}

		r.db.Create(&release)
		r.actions.ReleaseCreated(&release)

		return &ReleaseResolver{db: r.db, Release: release}, nil
	}
}

func (r *ReleaseResolver) ID() graphql.ID {
	return graphql.ID(r.Release.Model.ID.String())
}

func (r *ReleaseResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project

	r.db.Model(r.Release).Related(&project)

	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *ReleaseResolver) User(ctx context.Context) (*UserResolver, error) {
	var user models.User

	r.db.Model(r.Release).Related(&user)

	return &UserResolver{db: r.db, User: user}, nil
}

func (r *ReleaseResolver) HeadFeature() (*FeatureResolver, error) {
	var feature models.Feature

	r.db.Where("id = ?", r.Release.HeadFeatureID).First(&feature)

	return &FeatureResolver{db: r.db, Feature: feature}, nil
}

func (r *ReleaseResolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	var rows []models.ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.db.Where("release_id = ?", r.Release.ID).Find(&rows)
	for _, re := range rows {
		results = append(results, &ReleaseExtensionResolver{db: r.db, ReleaseExtension: re})
	}
	return results, nil

}

func (r *ReleaseResolver) TailFeature() (*FeatureResolver, error) {
	var feature models.Feature

	r.db.Where("id = ?", r.Release.TailFeatureID).First(&feature)

	return &FeatureResolver{db: r.db, Feature: feature}, nil
}

func (r *ReleaseResolver) State() string {
	return string(r.Release.State)
}

func (r *ReleaseResolver) StateMessage() string {
	return r.Release.StateMessage
}

func (r *ReleaseResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var environment models.Environment
	if r.db.Where("id = ?", r.Release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"service": r.Release,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}

func (r *ReleaseResolver) Artifacts() scalar.Json {
	return scalar.Json{r.Release.Artifacts.RawMessage}
}

func (r *ReleaseResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Release.Model.CreatedAt}
}
