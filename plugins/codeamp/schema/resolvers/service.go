package resolvers

import (
	"context"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ServiceInput struct {
	ID             *string
	Name           string
	Command        string
	ServiceSpec    string
	OneShot        bool
	Count          string
	ContainerPorts *[]models.ContainerPort
	ProjectId      string
}

type ServiceResolver struct {
	db      *gorm.DB
	Service models.Service
}

func (r *Resolver) Service(ctx context.Context, args *struct{ ID graphql.ID }) (*ServiceResolver, error) {
	service := models.Service{}
	if err := r.db.Where("id = ?", args.ID).First(&service).Error; err != nil {
		return nil, err
	}

	return &ServiceResolver{db: r.db, Service: service}, nil
}

func (r *Resolver) CreateService(args *struct{ Service *ServiceInput }) (*ServiceResolver, error) {
	projectId, err := uuid.FromString(args.Service.ProjectId)
	if err != nil {
		return &ServiceResolver{}, err
	}
	spew.Dump(args.Service)
	service := models.Service{
		Name:        args.Service.Name,
		Command:     args.Service.Command,
		ServiceSpec: args.Service.ServiceSpec,
		OneShot:     args.Service.OneShot,
		Count:       args.Service.Count,
		ProjectId:   projectId,
		Created:     time.Now(),
	}

	r.db.Create(&service)
	spew.Dump(*args.Service.ContainerPorts)

	if args.Service.ContainerPorts != nil {
		for _, cp := range *args.Service.ContainerPorts {
			containerPort := models.ContainerPort{
				ServiceId: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			r.db.Create(&containerPort)
		}
	}

	r.actions.ServiceCreated(&service)

	return &ServiceResolver{db: r.db, Service: service}, nil
}

func (r *ServiceResolver) ID() graphql.ID {
	return graphql.ID(r.Service.Model.ID.String())
}

func (r *ServiceResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project

	r.db.Model(r.Service).Related(&project)

	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *ServiceResolver) Name() string {
	return r.Service.Name
}

func (r *ServiceResolver) Command() string {
	return r.Service.Command
}

func (r *ServiceResolver) ServiceSpec() string {
	return r.Service.ServiceSpec
}

func (r *ServiceResolver) Count() string {
	return r.Service.Count
}

func (r *ServiceResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Service.Created}
}

func (r *ServiceResolver) OneShot() bool {
	return r.Service.OneShot
}

func (r *ServiceResolver) ContainerPorts(ctx context.Context) ([]*ContainerPortResolver, error) {
	var rows []models.ContainerPort
	var results []*ContainerPortResolver

	r.db.Where("service_id = ?", r.Service.ID).Order("created_at desc").Find(&rows)
	for _, cPort := range rows {
		spew.Dump(cPort)
		results = append(results, &ContainerPortResolver{ContainerPort: cPort})
	}
	spew.Dump(results)
	return results, nil
}

type ContainerPortResolver struct {
	ContainerPort models.ContainerPort
}

func (r *ContainerPortResolver) Port() string {
	return r.ContainerPort.Port
}

func (r *ContainerPortResolver) Protocol() string {
	return r.ContainerPort.Protocol
}
