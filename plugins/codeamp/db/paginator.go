package db_resolver

import (
	"reflect"

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
	ReleaseList    []model.Release
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

func EntryHelper(params model.PaginatorInput, entries interface{}) (interface{}, error) {
	filteredRows := []interface{}{}
	cursorRowIdx := 0

	// filter on things after cursor_id
	reflectedEntries := reflect.ValueOf(entries)

	cursorParamUUID := uuid.FromStringOrNil(*params.Cursor)

	if cursorParamUUID != uuid.Nil {
		for idx := 0; idx < reflectedEntries.Len(); idx++ {
			if params.Cursor != nil &&
				structs.Map(reflectedEntries.Index(idx).Interface())["Model"].(map[string]interface{})["ID"] == cursorParamUUID {
				cursorRowIdx = idx
				break
			}
		}
	}

	i := cursorRowIdx
	for {
		if len(filteredRows) == int(params.Limit) ||
			len(entries.([]interface{})) == i {
			break
		}
		filteredRows = append(filteredRows, entries.([]interface{})[i])
		i++
	}

	return filteredRows, nil
}

// Releases
func (r *ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	var results []*ReleaseResolver

	filteredRows, err := EntryHelper(*r.PaginatorInput, r.ReleaseList)
	if err != nil {
		return nil, err
	}

	for _, row := range filteredRows.([]model.Release) {
		results = append(results, &ReleaseResolver{
			DB:      r.DB,
			Release: row,
		})
	}
	return results, nil
}

func (r *ReleaseListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *ReleaseListResolver) Count() int32 {
	return int32(len(r.ReleaseList))
}

func (r *ReleaseListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ReleaseList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.ReleaseList) >= nextCursorIdx {
		return r.ReleaseList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *ReleaseListResolver) getPage() int32 {
	for idx, row := range r.ReleaseList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}

// SECRETS

// SecretResolver resolver for Release
type SecretListResolver struct {
	SecretList     []model.Secret
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Secrets
func (r *SecretListResolver) Entries() []*SecretResolver {
	var filteredRows []model.Secret
	var results []*SecretResolver

	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.SecretList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	i := cursorRowIdx
	for {
		if len(filteredRows) == int(r.PaginatorInput.Limit) ||
			len(r.SecretList) == i {
			break
		}
		filteredRows = append(filteredRows, r.SecretList[i])
		i++
	}

	for _, row := range filteredRows {
		var secretValue model.SecretValue

		r.DB.Where("secret_id = ?", row.Model.ID).Order("created_at desc").First(&secretValue)

		results = append(results, &SecretResolver{
			DB:          r.DB,
			Secret:      row,
			SecretValue: secretValue,
		})
	}
	return results
}

func (r *SecretListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *SecretListResolver) Count() int32 {
	return int32(len(r.SecretList))
}

func (r *SecretListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.SecretList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.SecretList) >= nextCursorIdx {
		return r.SecretList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *SecretListResolver) getPage() int32 {
	for idx, row := range r.SecretList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}

// SERVICES

// SecretResolver resolver for Release
type ServiceListResolver struct {
	ServiceList    []model.Service
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

// Services
func (r *ServiceListResolver) Entries() []*ServiceResolver {
	var filteredRows []model.Service
	var results []*ServiceResolver

	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ServiceList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	i := cursorRowIdx
	for {
		if len(filteredRows) == int(r.PaginatorInput.Limit) ||
			len(r.ServiceList) == i {
			break
		}
		filteredRows = append(filteredRows, r.ServiceList[i])
		i++
	}

	for _, row := range filteredRows {
		results = append(results, &ServiceResolver{
			DB:      r.DB,
			Service: row,
		})
	}
	return results
}

func (r *ServiceListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *ServiceListResolver) Count() int32 {
	return int32(len(r.ServiceList))
}

func (r *ServiceListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ServiceList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.ServiceList) >= nextCursorIdx {
		return r.ServiceList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *ServiceListResolver) getPage() int32 {
	for idx, row := range r.ServiceList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}

// FEATURES

// SecretResolver resolver for Release
type FeatureListResolver struct {
	FeatureList    []model.Feature
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

func (r *FeatureListResolver) Entries() []*FeatureResolver {
	var filteredRows []model.Feature
	var results []*FeatureResolver

	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.FeatureList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	i := cursorRowIdx
	for {
		if len(filteredRows) == int(r.PaginatorInput.Limit) ||
			len(r.FeatureList) == i {
			break
		}
		filteredRows = append(filteredRows, r.FeatureList[i])
		i++
	}

	for _, row := range filteredRows {
		results = append(results, &FeatureResolver{
			DB:      r.DB,
			Feature: row,
		})
	}
	return results
}

func (r *FeatureListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *FeatureListResolver) Count() int32 {
	return int32(len(r.FeatureList))
}

func (r *FeatureListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.FeatureList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.FeatureList) >= nextCursorIdx {
		return r.FeatureList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *FeatureListResolver) getPage() int32 {
	for idx, row := range r.FeatureList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}

// PROJECTS

// SecretResolver resolver for Release
type ProjectListResolver struct {
	ProjectList    []model.Project
	PaginatorInput *model.PaginatorInput
	DB             *gorm.DB
}

func (r *ProjectListResolver) Entries() []*ProjectResolver {
	var filteredRows []model.Project
	var results []*ProjectResolver

	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ProjectList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	i := cursorRowIdx
	for {
		if len(filteredRows) == int(r.PaginatorInput.Limit) ||
			len(r.ProjectList) == i {
			break
		}
		filteredRows = append(filteredRows, r.ProjectList[i])
		i++
	}

	for _, row := range filteredRows {
		results = append(results, &ProjectResolver{
			DB:      r.DB,
			Project: row,
		})
	}
	return results
}

func (r *ProjectListResolver) Page() int32 {
	// get page # from count / itemsPerPage
	return r.getPage()
}

func (r *ProjectListResolver) Count() int32 {
	return int32(len(r.ProjectList))
}

func (r *ProjectListResolver) NextCursor() string {
	cursorRowIdx := 0

	// filter on things after cursor_id
	for idx, row := range r.ProjectList {
		if r.PaginatorInput.Cursor != nil && row.Model.ID.String() == *r.PaginatorInput.Cursor {
			cursorRowIdx = idx
			break
		}
	}

	nextCursorIdx := cursorRowIdx + int(r.PaginatorInput.Limit) + 1
	if len(r.ProjectList) >= nextCursorIdx {
		return r.ProjectList[nextCursorIdx].Model.ID.String()
	} else {
		return ""
	}
}

func (r *ProjectListResolver) getPage() int32 {
	for idx, row := range r.ProjectList {
		if row.Model.ID.String() == *r.PaginatorInput.Cursor {
			return int32(idx)/r.PaginatorInput.Limit + int32(1)
		}
	}

	return 1
}
