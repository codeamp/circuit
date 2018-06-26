package db_resolver

import (
	"fmt"
	"log"

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

func EntryHelper(params model.PaginatorInput, entries interface{}) (interface{}, error) {
	filteredRows := []interface{}{}
	return filteredRows, nil
}

func NextCursorHelper(params model.PaginatorInput, entries interface{}) (string, error) {
	return "", nil
}

func PageHelper(params model.PaginatorInput, entries interface{}) (int32, error) {
	return int32(0), nil
}

// Releases
func (r *ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	var results []*ReleaseResolver

	whereString := ""
	limitString := ""
	log.Println(whereString, limitString)

	// r.Query.Where(whereString).Limit(limitString)

	return results, nil
}

func (r *ReleaseListResolver) Page() (int32, error) {
	return int32(0), nil
}

func (r *ReleaseListResolver) NextCursor() (string, error) {
	return "", nil
}

func (r *ReleaseListResolver) Count() (int32, error) {
	return int32(0), nil
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
