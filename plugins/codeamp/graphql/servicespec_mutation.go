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
		if err := tx.Where("is_default = ?", true).First(&currentDefault).Error; err == nil {
			currentDefault.IsDefault = false			
			if err := tx.Save(&currentDefault).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
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

	if err := tx.Create(&serviceSpec).Error; err != nil {
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
	isDefault := false

	tx := r.DB.Begin()

	if err := tx.Where("id = ?", args.ServiceSpec.ID).First(&serviceSpec).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("serviceSpec not found with given argument id")
	}

<<<<<<< HEAD
	/*
	* Find existing default; if input.default = true,
	* set existing default spec = false.
	* Condition: service spec cannot have a service mapped to it in order to be a default.
	*/
	if args.ServiceSpec.IsDefault && uuid.Equal(serviceSpec.ServiceID, uuid.Nil) {
		if err := tx.Where("is_default = ?", true).First(&currentDefault).Error; err == nil {
			currentDefault.IsDefault = false			
			if err := tx.Save(&currentDefault).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		isDefault = true
	}

	// check if currentDefault is the same as serviceSpec
	// if so, isDefault must always be true
	if serviceSpec.Model.ID.String() == currentDefault.Model.ID.String() {
		isDefault = true
	}

	/*
	* Find existing default; if input.default = true,
	* set existing default spec = false.
	* Condition: service spec cannot have a service mapped to it in order to be a default.
	*/
	if args.ServiceSpec.IsDefault && uuid.Equal(serviceSpec.ServiceID, uuid.Nil) {
		if err := tx.Where("is_default = ?", true).First(&currentDefault).Error; err == nil {
			currentDefault.IsDefault = false			
			if err := tx.Save(&currentDefault).Error; err != nil {
				tx.Rollback()
				return nil, err
			}	
		}

		isDefault = true		
	}

	// check if currentDefault is the same as serviceSpec
	// if so, isDefault must always be true
	if serviceSpec.Model.ID.String() == currentDefault.Model.ID.String() {
		isDefault = true
=======
	if r.DB.Where("id = ?", serviceSpecID).Find(&serviceSpec).RecordNotFound() {
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
>>>>>>> Add default property for service spec profiles
	}
	
	// if IsDefault is True, check which one is the current default
	if args.ServiceSpec.IsDefault {
		var currentDefault model.ServiceSpec
		if err := r.DB.Where("is_default = ?", true).First(&currentDefault).Error; err == nil {
			currentDefault.IsDefault = false			
			r.DB.Save(&currentDefault)
		}
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
	tx := r.DB.Begin()

	serviceSpec := model.ServiceSpec{}
	if err := tx.Where("id=?", args.ServiceSpec.ID).First(&serviceSpec).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("ServiceSpec not found with given argument id")
	} else {
		services := []model.Service{}
		if err := tx.Where("id = ?", serviceSpec.ServiceID).Find(&services).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		
		if serviceSpec.IsDefault {
			tx.Rollback()
			return nil, fmt.Errorf("Select another service spec to be a default before deleting this.")
		}
		
		if len(services) == 0 {
			if err := tx.Delete(&serviceSpec).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			//r.ServiceSpecDeleted(&serviceSpec)

			return &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}}, nil
		} else {
			return nil, fmt.Errorf("Delete all project-services using this service spec first.")
		}
	}
}
