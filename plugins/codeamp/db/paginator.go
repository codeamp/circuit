package db_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type Paginator struct {
	Page  int32 `json:"page"`
	Count int32 `json:"count"`
}

func GetPaginatorLimitAndOffset(paginatorInput *model.PaginatorInput) (string, int) {
	limit := "ALL"
	offset := 0

	if paginatorInput != nil {
		if paginatorInput.Limit != nil {
			inputLimit := int(*paginatorInput.Limit)

			if paginatorInput.Page != nil {
				offset = inputLimit * int(*paginatorInput.Page)
			}

			limit = fmt.Sprintf("%d", inputLimit)
		}
	}

	return limit, offset
}

type PaginatorResolver interface {
	Page() int32
	Count() int32
	Entries() []interface{}
}

// ReleaseResolver resolver for Release
type ReleaseListResolver struct {
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Entries
func (r *ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	var rows []model.Release
	var results []*ReleaseResolver

	limit, offset := GetPaginatorLimitAndOffset(r.PaginatorInput)
	err := r.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	unscoped_db := r.DB.New()
	for _, row := range rows {
		results = append(results, &ReleaseResolver{
			Release: row,
			DB:      unscoped_db,
		})
	}

	return results, nil
}

// Count
func (r *ReleaseListResolver) Count() (int32, error) {
	var count int
	r.DB.Model(&model.Release{}).Count(&count)

	return int32(count), nil
}

// SECRETS

// SecretResolver resolver for Secret
type SecretListResolver struct {
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Entries
func (r *SecretListResolver) Entries() ([]*SecretResolver, error) {
	var rows []model.Secret
	var results []*SecretResolver

	limit, offset := GetPaginatorLimitAndOffset(r.PaginatorInput)
	err := r.DB.Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	unscoped_db := r.DB.New()
	for _, row := range rows {
		results = append(results, &SecretResolver{
			Secret: row,
			DB:     unscoped_db,
		})
	}

	return results, nil
}

// Count
func (r *SecretListResolver) Count() (int32, error) {
	var count int
	r.DB.Model(&model.Secret{}).Count(&count)

	return int32(count), nil
}

// SERVICES

// ServiceResolver resolver for Secret
type ServiceListResolver struct {
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Entries
func (r *ServiceListResolver) Entries() ([]*ServiceResolver, error) {
	var rows []model.Service
	var results []*ServiceResolver

	limit, offset := GetPaginatorLimitAndOffset(r.PaginatorInput)
	err := r.DB.Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	unscoped_db := r.DB.New()
	for _, row := range rows {
		results = append(results, &ServiceResolver{
			Service: row,
			DB:      unscoped_db,
		})
	}

	return results, nil
}

// Count
func (r *ServiceListResolver) Count() (int32, error) {
	var count int
	r.DB.Model(&model.Service{}).Count(&count)

	return int32(count), nil
}

// FEATURES

// FeatureResolver resolver for Secret
type FeatureListResolver struct {
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Entries
func (r *FeatureListResolver) Entries() ([]*FeatureResolver, error) {
	var rows []model.Feature
	var results []*FeatureResolver

	limit, offset := GetPaginatorLimitAndOffset(r.PaginatorInput)
	err := r.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	unscoped_db := r.DB.New()
	for _, row := range rows {
		results = append(results, &FeatureResolver{
			Feature: row,
			DB:      unscoped_db,
		})
	}

	return results, nil
}

// Count
func (r *FeatureListResolver) Count() (int32, error) {
	var count int
	r.DB.Model(&model.Feature{}).Count(&count)

	return int32(count), nil
}

// PROJECTS

// ProjectResolver resolver for Secret
type ProjectListResolver struct {
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Entries
func (r *ProjectListResolver) Entries() ([]*ProjectResolver, error) {
	var rows []model.Project
	var results []*ProjectResolver

	limit, offset := GetPaginatorLimitAndOffset(r.PaginatorInput)
	err := r.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	unscoped_db := r.DB.New()
	for _, row := range rows {
		results = append(results, &ProjectResolver{
			Project: row,
			DB:      unscoped_db,
		})
	}

	return results, nil
}

// Count
func (r *ProjectListResolver) Count() (int32, error) {
	var count int
	r.DB.Model(&model.Project{}).Count(&count)

	return int32(count), nil
}
