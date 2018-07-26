package graphql_resolver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
)

func AppendPluginService(pluginServices []plugins.Service, service model.Service, spec model.ServiceSpec) []plugins.Service {
	terminationGracePeriod, _ := strconv.ParseInt(spec.TerminationGracePeriod, 10, 64)

	listeners := []plugins.Listener{}
	for _, l := range service.Ports {
		listener := plugins.Listener{
			Port:     l.Port,
			Protocol: l.Protocol,
		}
		listeners = append(listeners, listener)
	}

	readinessHeaders := []plugins.HealthProbeHttpHeader{}
	for _, h := range service.ReadinessProbe.HttpHeaders {
		header := plugins.HealthProbeHttpHeader{
			Name:  h.Name,
			Value: h.Value,
		}
		readinessHeaders = append(readinessHeaders, header)
	}

	livenessHeaders := []plugins.HealthProbeHttpHeader{}
	for _, h := range service.LivenessProbe.HttpHeaders {
		header := plugins.HealthProbeHttpHeader{
			Name:  h.Name,
			Value: h.Value,
		}
		livenessHeaders = append(livenessHeaders, header)
	}

	return append(pluginServices, plugins.Service{
		ID:        service.Model.ID.String(),
		Action:    transistor.GetAction("create"),
		State:     transistor.GetState("waiting"),
		Name:      service.Name,
		Command:   service.Command,
		Listeners: listeners,
		Replicas:  int64(service.Count),
		Spec: plugins.ServiceSpec{
			ID:                            spec.Model.ID.String(),
			CpuRequest:                    fmt.Sprintf("%sm", spec.CpuRequest),
			CpuLimit:                      fmt.Sprintf("%sm", spec.CpuLimit),
			MemoryRequest:                 fmt.Sprintf("%sMi", spec.MemoryRequest),
			MemoryLimit:                   fmt.Sprintf("%sMi", spec.MemoryLimit),
			TerminationGracePeriodSeconds: terminationGracePeriod,
		},
		Type: string(service.Type),
		DeploymentStrategy: plugins.DeploymentStrategy{
			Type:           service.DeploymentStrategy.Type,
			MaxUnavailable: service.DeploymentStrategy.MaxUnavailable,
			MaxSurge:       service.DeploymentStrategy.MaxSurge,
		},
		ReadinessProbe: plugins.ServiceHealthProbe{
			ServiceID:           service.ReadinessProbe.ServiceID,
			Type:                service.ReadinessProbe.Type,
			Method:              service.ReadinessProbe.Method,
			Command:             service.ReadinessProbe.Command,
			Port:                service.ReadinessProbe.Port,
			Scheme:              service.ReadinessProbe.Scheme,
			Path:                service.ReadinessProbe.Path,
			InitialDelaySeconds: service.ReadinessProbe.InitialDelaySeconds,
			PeriodSeconds:       service.ReadinessProbe.PeriodSeconds,
			TimeoutSeconds:      service.ReadinessProbe.TimeoutSeconds,
			SuccessThreshold:    service.ReadinessProbe.SuccessThreshold,
			FailureThreshold:    service.ReadinessProbe.FailureThreshold,
			HttpHeaders:         readinessHeaders,
		},
		LivenessProbe: plugins.ServiceHealthProbe{
			ServiceID:           service.LivenessProbe.ServiceID,
			Type:                service.LivenessProbe.Type,
			Method:              service.LivenessProbe.Method,
			Command:             service.LivenessProbe.Command,
			Port:                service.LivenessProbe.Port,
			Scheme:              service.LivenessProbe.Scheme,
			Path:                service.LivenessProbe.Path,
			InitialDelaySeconds: service.LivenessProbe.InitialDelaySeconds,
			PeriodSeconds:       service.LivenessProbe.PeriodSeconds,
			TimeoutSeconds:      service.LivenessProbe.TimeoutSeconds,
			SuccessThreshold:    service.LivenessProbe.SuccessThreshold,
			FailureThreshold:    service.LivenessProbe.FailureThreshold,
			HttpHeaders:         livenessHeaders,
		},
	})
}

func (r *Resolver) setupServices(services []model.Service) ([]plugins.Service, error) {
	var pluginServices []plugins.Service
	for _, service := range services {
		var spec model.ServiceSpec
		if r.DB.Where("id = ?", service.ServiceSpecID).First(&spec).RecordNotFound() {
			log.WarnWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecID,
			})
			return []plugins.Service{}, fmt.Errorf("ServiceSpec not found")
		}

		pluginServices = AppendPluginService(pluginServices, service, spec)
	}

	return pluginServices, nil
}

func BuildReleasePayload(release model.Release, project model.Project, environment model.Environment, branch string, headFeature model.Feature, tailFeature model.Feature, services []plugins.Service, secrets []plugins.Secret) plugins.Release {
	return plugins.Release{
		ID:          release.Model.ID.String(),
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
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
		Secrets:    secrets,
		Services:   services,
		IsRollback: release.IsRollback,
	}
}

func GetProjectExtensionsWithRoute53Subdomain(subdomain string, db *gorm.DB) []model.ProjectExtension {
	var existingProjectExtensions []model.ProjectExtension

	if db.Where("custom_config ->> 'subdomain' ilike ?", subdomain).Find(&existingProjectExtensions).RecordNotFound() {
		return []model.ProjectExtension{}
	}

	return existingProjectExtensions
}

func (r *Resolver) handleExtensionRoute53(args *struct{ ProjectExtension *model.ProjectExtensionInput }, projectExtension *model.ProjectExtension) error {
	extension := model.Extension{}

	// HOTFIX: check for existing subdomains for route53
	unmarshaledCustomConfig := make(map[string]interface{})
	err := json.Unmarshal(args.ProjectExtension.CustomConfig.RawMessage, &unmarshaledCustomConfig)
	if err != nil {
		return fmt.Errorf("Could not unmarshal custom config")
	}

	artifacts, err := ExtractArtifacts(*projectExtension, extension, r.DB)
	if err != nil {
		return err
	}

	hostedZoneId := ""
	for _, artifact := range artifacts {
		if artifact.Key == "HOSTED_ZONE_ID" {
			hostedZoneId = strings.ToUpper(artifact.Value.(string))
			break
		}
	}

	existingProjectExtensions := GetProjectExtensionsWithRoute53Subdomain(strings.ToUpper(unmarshaledCustomConfig["subdomain"].(string)), r.DB)
	for _, existingProjectExtension := range existingProjectExtensions {
		if existingProjectExtension.Model.ID.String() != "" {
			// check if HOSTED_ZONE_ID is the same
			var tmpExtension model.Extension

			r.DB.Where("id = ?", existingProjectExtension.ExtensionID).First(&tmpExtension)

			tmpExtensionArtifacts, err := ExtractArtifacts(existingProjectExtension, tmpExtension, r.DB)
			if err != nil {
				return err
			}

			for _, artifact := range tmpExtensionArtifacts {
				if artifact.Key == "HOSTED_ZONE_ID" &&
					strings.ToUpper(artifact.Value.(string)) == hostedZoneId {
					errMsg := "There is a route53 project extension with inputted subdomain already."
					log.InfoWithFields(errMsg, log.Fields{
						"project_extension_id":          projectExtension.Model.ID.String(),
						"existing_project_extension_id": existingProjectExtension.Model.ID.String(),
						"environment_id":                projectExtension.EnvironmentID.String(),
						"hosted_zone_id":                hostedZoneId,
					})
					return fmt.Errorf(errMsg)
				}
			}
		}
	}

	return nil
}
