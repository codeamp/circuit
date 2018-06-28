package db_resolver

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"

	"github.com/fatih/structs"
	uuid "github.com/satori/go.uuid"
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

func getCursorRowIdx(params model.PaginatorInput, entries []interface{}) (int, error) {
	cursorRowIdx := 0
	if params.Cursor != nil {
		uuid.FromStringOrNil(*params.Cursor)
	}

	cursorParamUUID := uuid.Nil
	// get cursor index to know upper bound of slice
	if cursorParamUUID != uuid.Nil {
		for idx := 0; idx < len(entries); idx++ {
			if params.Cursor != nil &&
				structs.Map(entries[idx])["Model"].(map[string]interface{})["ID"] == cursorParamUUID {
				cursorRowIdx = idx
				break
			}

			if idx == len(entries)-1 {
				return 0, fmt.Errorf("Could not find given cursor.")
			}
		}
	}

	return cursorRowIdx, nil
}

func EntryHelper(params model.PaginatorInput, query *gorm.DB, db *gorm.DB, rows *interface{}) error {
	cursorRow := make(map[string]interface{})
	spew.Dump(reflect.ValueOf(&rows).Type())

	if query == nil {
		return fmt.Errorf("Query attribute cannot be empty")
	}

	if params.Cursor != nil && len(*params.Cursor) > 0 {
		db.Where("id = ?", *params.Cursor).First(&cursorRow)
		// query.Order("created_at desc").Where("created_at < ?", cursorRow["model"].(map[string]interface{})["created_at"].(string)).Limit(int(*params.Limit)).Find(rows)
	} else {
		query.Order("created_at desc").Limit(int(*params.Limit)).Find(rows)
	}

	return nil
}

func NextCursorHelper(params model.PaginatorInput, entries interface{}) (string, error) {
	return "", nil
}

func PageHelper(params model.PaginatorInput, entries interface{}) (int32, error) {
	return int32(0), nil
}

// Entries
func (r *ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	var rows []model.Release
	var cursorRow model.Release
	var results []*ReleaseResolver

	if r.Query == nil {
		return nil, fmt.Errorf("Query attribute cannot be empty")
	}

	if r.PaginatorInput.Cursor != nil && len(*r.PaginatorInput.Cursor) > 0 {
		r.DB.Where("id = ?", *r.PaginatorInput.Cursor).First(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at < ?", cursorRow.Model.CreatedAt).Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(int(*r.PaginatorInput.Limit)).Find(&rows)
	}

	for _, row := range rows {
		results = append(results, &ReleaseResolver{
			Release: row,
			DB:      r.DB,
		})
	}

	return results, nil
}

func (r *ReleaseListResolver) Page() (int32, error) {
	var cursorRow model.Release
	var rows []model.Release

	limit := int(*r.PaginatorInput.Limit)
	index := 0

	if r.PaginatorInput.Cursor != nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at >= ?", cursorRow.Model.CreatedAt).Find(&rows).Count(&index)
	}

	if r.PaginatorInput.Cursor == nil || r.PaginatorInput.Limit == nil {
		return int32(1), nil
	} else {
		return int32((index / limit) + 1), nil
	}
}

func (r *ReleaseListResolver) NextCursor() (string, error) {
	var rows []model.Release
	var cursorRow model.Release

	nextCursorIdx := int(*r.PaginatorInput.Limit) + 1

	if r.PaginatorInput.Cursor != nil {
		r.DB.Where("id = ?", r.PaginatorInput.Cursor).Find(&cursorRow)
		r.Query.Order("created_at desc").Where("created_at < ?", cursorRow.Model.CreatedAt).Limit(nextCursorIdx).Find(&rows)
	} else {
		r.Query.Order("created_at desc").Limit(nextCursorIdx).Find(&rows)
	}

	if len(rows) == nextCursorIdx {
		return rows[nextCursorIdx-1].Model.ID.String(), nil
	} else {
		return "", nil
	}
}

func (r *ReleaseListResolver) Count() (int32, error) {
	var rows []model.Release

	r.Query.Find(&rows)

	return int32(len(rows)), nil
}

// SECRETS

// SecretResolver resolver for Release
type SecretListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
}

// Secrets
func (r *SecretListResolver) Entries() ([]*SecretResolver, error) {
	return []*SecretResolver{}, nil
}

func (r *SecretListResolver) Page() (int32, error) {
	return int32(0), nil
}

func (r *SecretListResolver) Count() (int32, error) {
	return int32(0), nil
}
func (r *SecretListResolver) NextCursor() (string, error) {
	return "", nil
}

// SERVICES

// ServiceListResolver
type ServiceListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
}

// Services
func (r *ServiceListResolver) Entries() ([]*ServiceResolver, error) {
	return []*ServiceResolver{}, nil
}

func (r *ServiceListResolver) Page() (int32, error) {
	return int32(0), nil
}

func (r *ServiceListResolver) NextCursor() (string, error) {
	return "", nil
}

func (r *ServiceListResolver) Count() (int32, error) {
	return int32(0), nil
}

// FEATURES

// FeatureListResolver
type FeatureListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
}

func (r *FeatureListResolver) Entries() ([]*FeatureResolver, error) {
	var results []*FeatureResolver
	return results, nil
}

func (r *FeatureListResolver) Page() (int32, error) {
	return int32(0), nil
}

func (r *FeatureListResolver) NextCursor() (string, error) {
	return "", nil
}

func (r *FeatureListResolver) Count() (int32, error) {
	return int32(0), nil
}

// PROJECTS

// ProjectListResolver
type ProjectListResolver struct {
	PaginatorInput *model.PaginatorInput
	Query          *gorm.DB
}

func (r *ProjectListResolver) Entries() ([]*ProjectResolver, error) {
	var results []*ProjectResolver
	return results, nil
}

func (r *ProjectListResolver) Page() (int32, error) {
	return int32(0), nil
}

func (r *ProjectListResolver) NextCursor() (string, error) {
	return "", nil
}

func (r *ProjectListResolver) Count() (int32, error) {
	return int32(0), nil
}
