package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type ServiceSpecInput struct {
	ID                     *string
	Name                   string
	CpuRequest             string
	CpuLimit               string
	MemoryRequest          string
	MemoryLimit            string
	TerminationGracePeriod string
}

type ServiceSpecResolver struct {
	db          *gorm.DB
	ServiceSpec models.ServiceSpec
}

func (r *Resolver) ServiceSpec(ctx context.Context, args *struct{ ID graphql.ID }) (*ServiceSpecResolver, error) {
	serviceSpec := models.ServiceSpec{}
	if err := r.db.Where("id = ?", args.ID).First(&serviceSpec).Error; err != nil {
		return nil, err
	}

	return &ServiceSpecResolver{db: r.db, ServiceSpec: serviceSpec}, nil
}

func (r *Resolver) CreateServiceSpec(args *struct{ ServiceSpec *ServiceSpecInput }) (*ServiceSpecResolver, error) {
	var err error

	if err != nil {
		return &ServiceSpecResolver{}, err
	}

	spew.Dump(args.ServiceSpec)
	serviceSpec := models.ServiceSpec{
		Name:                   args.ServiceSpec.Name,
		CpuRequest:             args.ServiceSpec.CpuRequest,
		CpuLimit:               args.ServiceSpec.CpuLimit,
		MemoryRequest:          args.ServiceSpec.MemoryRequest,
		MemoryLimit:            args.ServiceSpec.MemoryLimit,
		TerminationGracePeriod: args.ServiceSpec.TerminationGracePeriod,
	}

	r.db.Create(&serviceSpec)

	// r.actions.ServiceCreated(&serviceSpec)

	return &ServiceSpecResolver{db: r.db, ServiceSpec: serviceSpec}, nil
}

func (r *ServiceSpecResolver) ID() graphql.ID {
	return graphql.ID(r.ServiceSpec.Model.ID.String())
}

func (r *ServiceSpecResolver) Name(ctx context.Context) string {
	return r.ServiceSpec.Name
}

func (r *ServiceSpecResolver) CpuRequest(ctx context.Context) string {
	return r.ServiceSpec.CpuRequest
}

func (r *ServiceSpecResolver) CpuLimit(ctx context.Context) string {
	return r.ServiceSpec.CpuLimit
}

func (r *ServiceSpecResolver) MemoryLimit(ctx context.Context) string {
	return r.ServiceSpec.MemoryLimit
}

func (r *ServiceSpecResolver) MemoryRequest(ctx context.Context) string {
	return r.ServiceSpec.MemoryRequest
}

func (r *ServiceSpecResolver) TerminationGracePeriod(ctx context.Context) string {
	return r.ServiceSpec.TerminationGracePeriod
}

func (r *ServiceSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ServiceSpec.Created}
}
