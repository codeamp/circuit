package codeamp_resolvers

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
}

type PaginatorResolver interface {
	Page() int32
	Count() int32
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
}

func EntryHelper(params PaginatorInput, entries interface{}) (interface{}, error) {
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

	for _, row := range filteredRows.([]Release) {
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


// Secrets
func (r *SecretListResolver) Entries() []*SecretResolver {
	var filteredRows []Secret
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
		var secretValue SecretValue

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
	ServiceList    []Service
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

// Services
func (r *ServiceListResolver) Entries() []*ServiceResolver {
	var filteredRows []Service
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
	FeatureList    []Feature
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

func (r *FeatureListResolver) Entries() []*FeatureResolver {
	var filteredRows []Feature
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
	ProjectList    []Project
	PaginatorInput *PaginatorInput
	DB             *gorm.DB
}

func (r *ProjectListResolver) Entries() []*ProjectResolver {
	var filteredRows []Project
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
