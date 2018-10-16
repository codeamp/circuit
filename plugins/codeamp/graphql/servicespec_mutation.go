package graphql_resolver

import (
	"fmt"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// ServiceSpec Resolver Mutation
type ServiceSpecResolverMutation struct {
	DB *gorm.DB
}

func (r *ServiceSpecResolverMutation) CreateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{
		Name:                   args.ServiceSpec.Name,
		CpuRequest:             args.ServiceSpec.CpuRequest,
		CpuLimit:               args.ServiceSpec.CpuLimit,
		MemoryRequest:          args.ServiceSpec.MemoryRequest,
		MemoryLimit:            args.ServiceSpec.MemoryLimit,
		TerminationGracePeriod: args.ServiceSpec.TerminationGracePeriod,
	}

	r.DB.Create(&serviceSpec)

	return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
}

func (r *ServiceSpecResolverMutation) UpdateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{}

	serviceSpecID, err := uuid.FromString(*args.ServiceSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("UpdateServiceSpec: Missing argument id")
	}

	if r.DB.Where("id=?", serviceSpecID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	}

	serviceSpec.Name = args.ServiceSpec.Name
	serviceSpec.CpuLimit = args.ServiceSpec.CpuLimit
	serviceSpec.CpuRequest = args.ServiceSpec.CpuRequest
	serviceSpec.MemoryLimit = args.ServiceSpec.MemoryLimit
	serviceSpec.MemoryRequest = args.ServiceSpec.MemoryRequest
	serviceSpec.TerminationGracePeriod = args.ServiceSpec.TerminationGracePeriod

	r.DB.Save(&serviceSpec)

	//r.ServiceSpecUpdated(&serviceSpec)

	return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
}

func (r *ServiceSpecResolverMutation) DeleteServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{}
	if r.DB.Where("id=?", args.ServiceSpec.ID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	} else {
		services := []model.Service{}
		r.DB.Where("service_spec_id = ?", serviceSpec.Model.ID).Find(&services)
		if len(services) == 0 {
			r.DB.Delete(&serviceSpec)

			//r.ServiceSpecDeleted(&serviceSpec)

			return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
		} else {
			return nil, fmt.Errorf("Delete all project-services using this service spec first.")
		}
	}
}
