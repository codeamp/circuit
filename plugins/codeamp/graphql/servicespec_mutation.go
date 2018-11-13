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

// CreateServiceSpec
func (r *ServiceSpecResolverMutation) CreateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	// make create operation atomic
	tx := r.DB.Begin()
	currentDefault := model.ServiceSpec{}

	/*
	* Find existing default; if input.default = true,
	* set existing default spec = false.
	*/
	if args.ServiceSpec.IsDefault {
		if err := tx.Where("is_default = ?", true).First(&currentDefault).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("could not find default service spec")
		}
		
		currentDefault.IsDefault = false			
		if err := tx.Save(&currentDefault).Error; err != nil {
			tx.Rollback()
			return nil, err
		}		
	}

	serviceSpec := model.ServiceSpec{
		Name:                   args.ServiceSpec.Name,
		CpuRequest:             args.ServiceSpec.CpuRequest,
		CpuLimit:               args.ServiceSpec.CpuLimit,
		MemoryRequest:          args.ServiceSpec.MemoryRequest,
		MemoryLimit:            args.ServiceSpec.MemoryLimit,
		TerminationGracePeriod: args.ServiceSpec.TerminationGracePeriod,
		IsDefault: 				args.ServiceSpec.IsDefault,
	}

	if err := r.DB.Create(&serviceSpec).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
}

// UpdateServiceSpec
func (r *ServiceSpecResolverMutation) UpdateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	serviceSpec := model.ServiceSpec{}
	currentDefault := model.ServiceSpec{}
	isDefault := args.ServiceSpec.IsDefault

	serviceSpecID, err := uuid.FromString(*args.ServiceSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("missing argument id")
	}

	if r.DB.Where("id = ?", serviceSpecID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("serviceSpec not found with given argument id")
	}

	tx := r.DB.Begin()
	if err := tx.Where("is_default = ?", true).First(&currentDefault).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("could not find default service spec")
	}

	/*
	* Find existing default; if input.default = true,
	* set existing default spec = false.
	*/
	if args.ServiceSpec.IsDefault {
		if err := tx.Where("is_default = ?", true).First(&currentDefault).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("could not find default service spec")
		}
		
		currentDefault.IsDefault = false			
		if err := tx.Save(&currentDefault).Error; err != nil {
			tx.Rollback()
			return nil, err
		}		
	}

	// check if currentDefault is the same as serviceSpec
	// if so, isDefault must always be true
	if serviceSpec.Model.ID.String() == currentDefault.Model.ID.String() {
		isDefault = true
	}

	serviceSpec.Name = args.ServiceSpec.Name
	serviceSpec.CpuLimit = args.ServiceSpec.CpuLimit
	serviceSpec.CpuRequest = args.ServiceSpec.CpuRequest
	serviceSpec.MemoryLimit = args.ServiceSpec.MemoryLimit
	serviceSpec.MemoryRequest = args.ServiceSpec.MemoryRequest
	serviceSpec.TerminationGracePeriod = args.ServiceSpec.TerminationGracePeriod
	serviceSpec.IsDefault = isDefault


	if err := tx.Save(&serviceSpec).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
}


// DeleteServiceSpec
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
