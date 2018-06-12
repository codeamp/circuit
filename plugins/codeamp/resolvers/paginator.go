package codeamp_resolvers

<<<<<<< HEAD
import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type Paginator struct {
	Page       int32  `json:"page"`
	Count      int32  `json:"count"`
	NextCursor string `json:"nextCursor"`
=======
type Paginator struct {
	Page        int32 `json:"page"`
	Count       int32 `json:"count"`
	HasNextPage bool  `json:"hasNextPage"`
>>>>>>> Add pagination for Releases
}

type PaginatorResolver interface {
	Page() int32
	Count() int32
<<<<<<< HEAD
	Entries() []interface{}
	NextCursor() string
}

// ReleaseResolver resolver for Release
type ReleaseListResolver struct {
	ReleaseList    []Release
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

func convertToInterfaceSlice(entries interface{}) ([]interface{}, error) {
	val := reflect.ValueOf(entries)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("Entries must be a slice.")
	}

	out := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		out[i] = val.Index(i).Interface()
	}
	
	return out, nil
}

func getCursorRowIdx(params PaginatorInput, entries []interface{}) (int, error) {
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

			if idx == len(entries) - 1 {
				return 0, fmt.Errorf("Could not find given cursor.")
			}			
		}
	}

	return cursorRowIdx, nil
}

func EntryHelper(params PaginatorInput, entries interface{}) (interface{}, error) {
	out, err := convertToInterfaceSlice(entries)
	if err != nil {
		return nil, err
	}

	cursorRowIdx, err := getCursorRowIdx(params, out)
	if err != nil {
		return nil, err
	}
	
	filteredRows := []interface{}{}
	if params.Limit == nil {
		return out, nil
	} else {
		i := cursorRowIdx
		for {
			if len(filteredRows) == int(*params.Limit) || len(out) == i {
				break
			}		
			
			filteredRows = append(filteredRows, out[i])
			i++
		}
	
		return filteredRows, nil
	}
}

func NextCursorHelper(params PaginatorInput, entries interface{}) (string, error) {
	out, err := convertToInterfaceSlice(entries)
	if err != nil {
		return "", err
	}

	cursorRowIdx, err := getCursorRowIdx(params, out)
	if err != nil {
		return "", err
	}

	if params.Limit == nil {
		return "", nil
	} else {
		nextCursorIdx := cursorRowIdx + int(*params.Limit) + 1
		if len(out) > nextCursorIdx {
			return structs.Map(out[nextCursorIdx])["Model"].(map[string]interface{})["ID"].(uuid.UUID).String(), nil
		} else {
			return "", nil
		}
	}
}


func PageHelper(params PaginatorInput, entries interface{}) (int32, error) {
	if params.Limit == nil {
		return int32(1), nil
	}

	out, err := convertToInterfaceSlice(entries)
	if err != nil {
		return int32(1), err
	}

	cursorRowIdx, err := getCursorRowIdx(params, out)
	if err != nil {
		return int32(1), err
	}
	
	return int32(cursorRowIdx)/(*params.Limit) + int32(1), nil
}

// Releases
func (r *ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	var results []*ReleaseResolver

	filteredRows, err := EntryHelper(*r.PaginatorInput, r.ReleaseList)
	if err != nil {
		return nil, err
	}

	for _, row := range filteredRows.([]interface{}) {
		release := Release{}
		releaseBytes, _ := json.Marshal(row)
		json.Unmarshal(releaseBytes, &release)

		results = append(results, &ReleaseResolver{
			DB:      r.DB,
			Release: release,
		})
	}
	return results, nil
}

func (r *ReleaseListResolver) Page() (int32, error) {
	return PageHelper(*r.PaginatorInput, r.ReleaseList)
}

func (r *ReleaseListResolver) Count() int32 {
	return int32(len(r.ReleaseList))
}

func (r *ReleaseListResolver) NextCursor() (string, error) {
	return NextCursorHelper(*r.PaginatorInput, r.ReleaseList)
}

// SECRETS

// SecretListResolver
type SecretListResolver struct {
	SecretList     []Secret
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

// Secrets
func (r *SecretListResolver) Entries() ([]*SecretResolver, error) {
	var results []*SecretResolver

	filteredRows, err := EntryHelper(*r.PaginatorInput, r.SecretList)
	if err != nil {
		return nil, err
	}

	for _, row := range filteredRows.([]interface{}) {
		secret := Secret{}
		secretBytes, _ := json.Marshal(row)
		json.Unmarshal(secretBytes, &secret)

		results = append(results, &SecretResolver{
			DB:      r.DB,
			Secret: secret,
		})
	}
	
	return results, nil
}

func (r *SecretListResolver) Page() (int32, error) {
	return PageHelper(*r.PaginatorInput, r.SecretList)
}

func (r *SecretListResolver) Count() int32 {
	return int32(len(r.SecretList))
}

func (r *SecretListResolver) NextCursor() (string, error) {
	return NextCursorHelper(*r.PaginatorInput, r.SecretList)
}

// SERVICES

// ServiceListResolver
type ServiceListResolver struct {
	ServiceList    []Service
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

// Services
func (r *ServiceListResolver) Entries() ([]*ServiceResolver, error) {
	var results []*ServiceResolver

	filteredRows, err := EntryHelper(*r.PaginatorInput, r.ServiceList)
	if err != nil {
		return nil, err
	}

	for _, row := range filteredRows.([]interface{}) {
		service := Service{}
		serviceBytes, _ := json.Marshal(row)
		json.Unmarshal(serviceBytes, &service)

		results = append(results, &ServiceResolver{
			DB:      r.DB,
			Service: service,
		})
	}
	
	return results, nil
}

func (r *ServiceListResolver) Page() (int32, error) {
	return PageHelper(*r.PaginatorInput, r.ServiceList)
}

func (r *ServiceListResolver) Count() int32 {
	return int32(len(r.ServiceList))
}

func (r *ServiceListResolver) NextCursor() (string, error) {
	return NextCursorHelper(*r.PaginatorInput, r.ServiceList)
}

// FEATURES

// FeatureListResolver
type FeatureListResolver struct {
	FeatureList    []Feature
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

func (r *FeatureListResolver) Entries() ([]*FeatureResolver, error) {
	var results []*FeatureResolver

	filteredRows, err := EntryHelper(*r.PaginatorInput, r.FeatureList)
	if err != nil {
		return nil, err
	}

	for _, row := range filteredRows.([]interface{}) {
		feature := Feature{}
		featureBytes, _ := json.Marshal(row)
		json.Unmarshal(featureBytes, &feature)

		results = append(results, &FeatureResolver{
			DB:      r.DB,
			Feature: feature,
		})
	}
	
	return results, nil
}

func (r *FeatureListResolver) Page() (int32, error) {
	return PageHelper(*r.PaginatorInput, r.FeatureList)
}

func (r *FeatureListResolver) Count() int32 {
	return int32(len(r.FeatureList))
}

func (r *FeatureListResolver) NextCursor() (string, error) {
	return NextCursorHelper(*r.PaginatorInput, r.FeatureList)
}

// PROJECTS

// ProjectListResolver
type ProjectListResolver struct {
	ProjectList    []Project
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

func (r *ProjectListResolver) Entries() ([]*ProjectResolver, error) {
	var results []*ProjectResolver

	filteredRows, err := EntryHelper(*r.PaginatorInput, r.ProjectList)
	if err != nil {
		return nil, err
	}

	for _, row := range filteredRows.([]interface{}) {
		project := Project{}
		projectBytes, _ := json.Marshal(row)
		json.Unmarshal(projectBytes, &project)

		results = append(results, &ProjectResolver{
			DB:      r.DB,
			Project: project,
		})
	}
	
	return results, nil
}

func (r *ProjectListResolver) Page() (int32, error) {
	return PageHelper(*r.PaginatorInput, r.ProjectList)
}

func (r *ProjectListResolver) Count() int32 {
	return int32(len(r.ProjectList))
}

func (r *ProjectListResolver) NextCursor() (string, error) {
	return NextCursorHelper(*r.PaginatorInput, r.ProjectList)
=======
	HasNextPage() int32
>>>>>>> Add pagination for Releases
}
