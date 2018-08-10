package db_resolver

import (
	"fmt"
	"math"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type Paginator struct {
	Page       int32  `json:"page"`
	Count      int32  `json:"count"`
	NextCursor string `json:"nextCursor"`
}

type PaginatorResolver interface {
	Page() int32
	Count() int32
	Entries() []interface{}
	NextCursor() string
}

// ReleaseResolver resolver for Release
type ReleaseListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
	DB             *gorm.DB
}

// Entries
func (r *ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	var rows []model.Release
	var cursorRow model.Release
	var results []*ReleaseResolver

	if r.Query == nil {
		return nil, fmt.Errorf("Query attribute cannot be empty")
	}

	if r.PaginatorInput != nil && r.PaginatorInput.Limit != nil {
		if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
			r.DB.Where("id = ?", *r.PaginatorInput.Cursor).First(&cursorRow)
			r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		} else {
			r.Query.Order("created_at desc").Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		}
	} else {
		r.Query.Order("created_at desc").Find(&rows)
	}

	for _, row := range rows {
		results = append(results, &ReleaseResolver{
			Release: row,
			DB:      r.DB,
		})
	}

	return results, nil
}

// Page
func (r *ReleaseListResolver) Page() (int32, error) {
	var cursorRow model.Release
	var rows []model.Release

	if r.PaginatorInput == nil {
		return int32(1), nil
	}

	index := 0

	if (r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0) || r.PaginatorInput.Limit == nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at >= ?", cursorRow.Model.CreatedAt).Find(&rows).Count(&index)
		return int32(math.Ceil((float64(index) / float64(*r.PaginatorInput.Limit)))), nil
	} else {
		return int32(1), nil
	}
}

// NextCursor
func (r *ReleaseListResolver) NextCursor() (string, error) {
	var rows []model.Release
	var cursorRow model.Release

	if r.PaginatorInput == nil {
		return "", nil
	}

	nextCursorIdx := int(*r.PaginatorInput.Limit) + 1

	if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(nextCursorIdx).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(nextCursorIdx).Find(&rows)
	}

	if len(rows) == nextCursorIdx {
		return rows[nextCursorIdx-1].Model.ID.String(), nil
	} else {
		return "", nil
	}
}

// Count
func (r *ReleaseListResolver) Count() (int32, error) {
	var count int
	r.Query.Model(&model.Release{}).Count(&count)

	return int32(count), nil
}

// SECRETS

// SecretResolver resolver for Secret
type SecretListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
	DB             *gorm.DB
}

// Entries
func (r *SecretListResolver) Entries() ([]*SecretResolver, error) {
	var rows []model.Secret
	var cursorRow model.Secret
	var results []*SecretResolver

	if r.Query == nil {
		return nil, fmt.Errorf("Query attribute cannot be empty")
	}

	if r.PaginatorInput != nil && r.PaginatorInput.Limit != nil {
		if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
			r.DB.Where("id = ?", *r.PaginatorInput.Cursor).First(&cursorRow)
			r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		} else {
			r.Query.Order("created_at desc").Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		}
	} else {
		r.Query.Order("created_at desc").Find(&rows)
	}

	for _, row := range rows {
		results = append(results, &SecretResolver{
			Secret: row,
			DB:     r.DB,
		})
	}

	return results, nil
}

// Page
func (r *SecretListResolver) Page() (int32, error) {
	var cursorRow model.Secret
	var rows []model.Secret

	if r.PaginatorInput == nil {
		return int32(1), nil
	}

	index := 0

	if (r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0) || r.PaginatorInput.Limit == nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at >= ?", cursorRow.Model.CreatedAt).Find(&rows).Count(&index)
		return int32(math.Ceil((float64(index) / float64(*r.PaginatorInput.Limit)))), nil
	} else {
		return int32(1), nil
	}
}

// NextCursor
func (r *SecretListResolver) NextCursor() (string, error) {
	var rows []model.Secret
	var cursorRow model.Secret

	if r.PaginatorInput == nil {
		return "", nil
	}

	nextCursorIdx := int(*r.PaginatorInput.Limit) + 1

	if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(nextCursorIdx).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(nextCursorIdx).Find(&rows)
	}

	if len(rows) == nextCursorIdx {
		return rows[nextCursorIdx-1].Model.ID.String(), nil
	} else {
		return "", nil
	}
}

// Count
func (r *SecretListResolver) Count() (int32, error) {
	var count int
	r.Query.Model(&model.Secret{}).Count(&count)

	return int32(count), nil
}

// SERVICES

// ServiceResolver resolver for Secret
type ServiceListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
	DB             *gorm.DB
}

// Entries
func (r *ServiceListResolver) Entries() ([]*ServiceResolver, error) {
	var rows []model.Service
	var cursorRow model.Service
	var results []*ServiceResolver

	if r.Query == nil {
		return nil, fmt.Errorf("Query attribute cannot be empty")
	}

	if r.PaginatorInput != nil && r.PaginatorInput.Limit != nil {
		if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
			r.DB.Where("id = ?", *r.PaginatorInput.Cursor).First(&cursorRow)
			r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		} else {
			r.Query.Order("created_at desc").Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		}
	} else {
		r.Query.Order("created_at desc").Find(&rows)
	}

	for _, row := range rows {
		results = append(results, &ServiceResolver{
			Service: row,
			DB:      r.DB,
		})
	}

	return results, nil
}

// Page
func (r *ServiceListResolver) Page() (int32, error) {
	var cursorRow model.Service
	var rows []model.Service

	if r.PaginatorInput == nil {
		return int32(1), nil
	}

	index := 0

	if (r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0) || r.PaginatorInput.Limit == nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at >= ?", cursorRow.Model.CreatedAt).Find(&rows).Count(&index)
		return int32(math.Ceil((float64(index) / float64(*r.PaginatorInput.Limit)))), nil
	} else {
		return int32(1), nil
	}
}

// NextCursor
func (r *ServiceListResolver) NextCursor() (string, error) {
	var rows []model.Service
	var cursorRow model.Service

	if r.PaginatorInput == nil {
		return "", nil
	}

	nextCursorIdx := int(*r.PaginatorInput.Limit) + 1

	if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(nextCursorIdx).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(nextCursorIdx).Find(&rows)
	}

	if len(rows) == nextCursorIdx {
		return rows[nextCursorIdx-1].Model.ID.String(), nil
	} else {
		return "", nil
	}
}

// Count
func (r *ServiceListResolver) Count() (int32, error) {
	var count int
	r.Query.Model(&model.Service{}).Count(&count)

	return int32(count), nil
}

// FEATURES

// FeatureResolver resolver for Secret
type FeatureListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
	DB             *gorm.DB
}

// Entries
func (r *FeatureListResolver) Entries() ([]*FeatureResolver, error) {
	var rows []model.Feature
	var cursorRow model.Feature
	var results []*FeatureResolver

	if r.Query == nil {
		return nil, fmt.Errorf("Query attribute cannot be empty")
	}

	if r.PaginatorInput != nil && r.PaginatorInput.Limit != nil {
		if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
			r.DB.Where("id = ?", *r.PaginatorInput.Cursor).First(&cursorRow)
			r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		} else {
			r.Query.Order("created_at desc").Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		}
	} else {
		r.Query.Order("created_at desc").Find(&rows)
	}

	for _, row := range rows {
		results = append(results, &FeatureResolver{
			Feature: row,
			DB:      r.DB,
		})
	}

	return results, nil
}

// Page
func (r *FeatureListResolver) Page() (int32, error) {
	var cursorRow model.Feature
	var rows []model.Feature

	if r.PaginatorInput == nil {
		return int32(1), nil
	}

	index := 0

	if (r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0) || r.PaginatorInput.Limit == nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at >= ?", cursorRow.Model.CreatedAt).Find(&rows).Count(&index)
		return int32(math.Ceil((float64(index) / float64(*r.PaginatorInput.Limit)))), nil
	} else {
		return int32(1), nil
	}
}

// NextCursor
func (r *FeatureListResolver) NextCursor() (string, error) {
	var rows []model.Feature
	var cursorRow model.Feature

	if r.PaginatorInput == nil {
		return "", nil
	}

	nextCursorIdx := int(*r.PaginatorInput.Limit) + 1

	if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(nextCursorIdx).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(nextCursorIdx).Find(&rows)
	}

	if len(rows) == nextCursorIdx {
		return rows[nextCursorIdx-1].Model.ID.String(), nil
	} else {
		return "", nil
	}
}

// Count
func (r *FeatureListResolver) Count() (int32, error) {
	var count int
	r.Query.Model(&model.Feature{}).Count(&count)

	return int32(count), nil
}

// PROJECTS

// ProjectResolver resolver for Secret
type ProjectListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
	DB             *gorm.DB
}

// Entries
func (r *ProjectListResolver) Entries() ([]*ProjectResolver, error) {
	var rows []model.Project
	var cursorRow model.Project
	var results []*ProjectResolver

	if r.Query == nil {
		return nil, fmt.Errorf("Query attribute cannot be empty")
	}

	if r.PaginatorInput != nil && r.PaginatorInput.Limit != nil {
		if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
			r.DB.Where("id = ?", *r.PaginatorInput.Cursor).First(&cursorRow)
			r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		} else {
			r.Query.Order("created_at desc").Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
		}
	} else {
		r.Query.Order("created_at desc").Find(&rows)
	}

	for _, row := range rows {
		results = append(results, &ProjectResolver{
			Project: row,
			DB:      r.DB,
		})
	}

	return results, nil
}

// Page
func (r *ProjectListResolver) Page() (int32, error) {
	var cursorRow model.Project
	var rows []model.Project

	if r.PaginatorInput == nil {
		return int32(1), nil
	}

	index := 0

	if (r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0) || r.PaginatorInput.Limit == nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at >= ?", cursorRow.Model.CreatedAt).Find(&rows).Count(&index)
		return int32(math.Ceil((float64(index) / float64(*r.PaginatorInput.Limit)))), nil
	} else {
		return int32(1), nil
	}
}

// NextCursor
func (r *ProjectListResolver) NextCursor() (string, error) {
	var rows []model.Project
	var cursorRow model.Project

	if r.PaginatorInput == nil {
		return "", nil
	}

	nextCursorIdx := int(*r.PaginatorInput.Limit) + 1

	if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at <= ?", cursorRow.Model.CreatedAt).Limit(nextCursorIdx).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(nextCursorIdx).Find(&rows)
	}

	if len(rows) == nextCursorIdx {
		return rows[nextCursorIdx-1].Model.ID.String(), nil
	} else {
		return "", nil
	}
}

// Count
func (r *ProjectListResolver) Count() (int32, error) {
	var count int
	r.Query.Model(&model.Project{}).Count(&count)

	return int32(count), nil
}
