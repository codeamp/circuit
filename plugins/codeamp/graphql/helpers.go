package graphql_resolver

import (
	"fmt"
	"strconv"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
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
		PreStopHook: service.PreStopHook,
	})
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
		Secrets:             secrets,
		Services:            services,
		IsRollback:          release.IsRollback,
		Redeployable:        release.Redeployable,
		RedeployableMessage: release.RedeployableMessage,
	}
}

func GetProjectExtensionsWithRoute53Subdomain(subdomain string, db *gorm.DB) []model.ProjectExtension {
	var existingProjectExtensions []model.ProjectExtension

	if db.Where("custom_config ->> 'subdomain' ilike ?", subdomain).Find(&existingProjectExtensions).RecordNotFound() {
		return []model.ProjectExtension{}
	}

	return existingProjectExtensions
}
