package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// These imports may not be used depending on the input parameters
var _ = json.Marshal
var _ = fmt.Sprintf
var _ = url.QueryEscape
var _ = strconv.FormatInt
var _ = strings.Replace
var _ = validate.Maximum
var _ = strfmt.NewFormats

// GetClustersInput holds the input parameters for a getClusters operation.
type GetClustersInput struct {
	GroupID string
}

// ValidateGetClustersInput returns an error if the input parameter doesn't
// satisfy the requirements in the swagger yml file.
func ValidateGetClustersInput(groupID string) error {

	return nil
}

// GetClustersInputPath returns the URI path for the input.
func GetClustersInputPath(groupID string) (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/clusters"
	urlVals := url.Values{}

	pathgroupID := groupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// CreateClusterInput holds the input parameters for a createCluster operation.
type CreateClusterInput struct {
	GroupID                      string
	CreateOrUpdateClusterRequest *CreateOrUpdateClusterRequest
}

// Validate returns an error if any of the CreateClusterInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i CreateClusterInput) Validate() error {

	if i.CreateOrUpdateClusterRequest != nil {
		if err := i.CreateOrUpdateClusterRequest.Validate(nil); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the URI path for the input.
func (i CreateClusterInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/clusters"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// DeleteClusterInput holds the input parameters for a deleteCluster operation.
type DeleteClusterInput struct {
	GroupID     string
	ClusterName string
}

// Validate returns an error if any of the DeleteClusterInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i DeleteClusterInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i DeleteClusterInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/clusters/{clusterName}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathclusterName := i.ClusterName
	if pathclusterName == "" {
		err := fmt.Errorf("clusterName cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{clusterName}", pathclusterName, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetClusterInput holds the input parameters for a getCluster operation.
type GetClusterInput struct {
	GroupID     string
	ClusterName string
}

// Validate returns an error if any of the GetClusterInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetClusterInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i GetClusterInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/clusters/{clusterName}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathclusterName := i.ClusterName
	if pathclusterName == "" {
		err := fmt.Errorf("clusterName cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{clusterName}", pathclusterName, -1)

	return path + "?" + urlVals.Encode(), nil
}

// UpdateClusterInput holds the input parameters for a updateCluster operation.
type UpdateClusterInput struct {
	GroupID                      string
	ClusterName                  string
	CreateOrUpdateClusterRequest *CreateOrUpdateClusterRequest
}

// Validate returns an error if any of the UpdateClusterInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i UpdateClusterInput) Validate() error {

	if i.CreateOrUpdateClusterRequest != nil {
		if err := i.CreateOrUpdateClusterRequest.Validate(nil); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the URI path for the input.
func (i UpdateClusterInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/clusters/{clusterName}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathclusterName := i.ClusterName
	if pathclusterName == "" {
		err := fmt.Errorf("clusterName cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{clusterName}", pathclusterName, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetContainersInput holds the input parameters for a getContainers operation.
type GetContainersInput struct {
	GroupID string
}

// ValidateGetContainersInput returns an error if the input parameter doesn't
// satisfy the requirements in the swagger yml file.
func ValidateGetContainersInput(groupID string) error {

	return nil
}

// GetContainersInputPath returns the URI path for the input.
func GetContainersInputPath(groupID string) (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/containers"
	urlVals := url.Values{}

	pathgroupID := groupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// CreateContainerInput holds the input parameters for a createContainer operation.
type CreateContainerInput struct {
	GroupID                        string
	CreateOrUpdateContainerRequest *CreateOrUpdateContainerRequest
}

// Validate returns an error if any of the CreateContainerInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i CreateContainerInput) Validate() error {

	if i.CreateOrUpdateContainerRequest != nil {
		if err := i.CreateOrUpdateContainerRequest.Validate(nil); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the URI path for the input.
func (i CreateContainerInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/containers"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetContainerInput holds the input parameters for a getContainer operation.
type GetContainerInput struct {
	GroupID     string
	ContainerID string
}

// Validate returns an error if any of the GetContainerInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetContainerInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i GetContainerInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/containers/{containerID}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathcontainerID := i.ContainerID
	if pathcontainerID == "" {
		err := fmt.Errorf("containerID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{containerID}", pathcontainerID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// UpdateContainerInput holds the input parameters for a updateContainer operation.
type UpdateContainerInput struct {
	GroupID                        string
	ContainerID                    string
	CreateOrUpdateContainerRequest *CreateOrUpdateContainerRequest
}

// Validate returns an error if any of the UpdateContainerInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i UpdateContainerInput) Validate() error {

	if i.CreateOrUpdateContainerRequest != nil {
		if err := i.CreateOrUpdateContainerRequest.Validate(nil); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the URI path for the input.
func (i UpdateContainerInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/containers/{containerID}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathcontainerID := i.ContainerID
	if pathcontainerID == "" {
		err := fmt.Errorf("containerID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{containerID}", pathcontainerID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetDatabaseUsersInput holds the input parameters for a getDatabaseUsers operation.
type GetDatabaseUsersInput struct {
	GroupID string
}

// ValidateGetDatabaseUsersInput returns an error if the input parameter doesn't
// satisfy the requirements in the swagger yml file.
func ValidateGetDatabaseUsersInput(groupID string) error {

	return nil
}

// GetDatabaseUsersInputPath returns the URI path for the input.
func GetDatabaseUsersInputPath(groupID string) (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/databaseUsers"
	urlVals := url.Values{}

	pathgroupID := groupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// CreateDatabaseUserInput holds the input parameters for a createDatabaseUser operation.
type CreateDatabaseUserInput struct {
	GroupID                   string
	CreateDatabaseUserRequest *CreateDatabaseUserRequest
}

// Validate returns an error if any of the CreateDatabaseUserInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i CreateDatabaseUserInput) Validate() error {

	if i.CreateDatabaseUserRequest != nil {
		if err := i.CreateDatabaseUserRequest.Validate(nil); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the URI path for the input.
func (i CreateDatabaseUserInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/databaseUsers"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// DeleteDatabaseUserInput holds the input parameters for a deleteDatabaseUser operation.
type DeleteDatabaseUserInput struct {
	GroupID  string
	Username string
}

// Validate returns an error if any of the DeleteDatabaseUserInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i DeleteDatabaseUserInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i DeleteDatabaseUserInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/databaseUsers/admin/{username}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathusername := i.Username
	if pathusername == "" {
		err := fmt.Errorf("username cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{username}", pathusername, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetDatabaseUserInput holds the input parameters for a getDatabaseUser operation.
type GetDatabaseUserInput struct {
	GroupID  string
	Username string
}

// Validate returns an error if any of the GetDatabaseUserInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetDatabaseUserInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i GetDatabaseUserInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/databaseUsers/admin/{username}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathusername := i.Username
	if pathusername == "" {
		err := fmt.Errorf("username cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{username}", pathusername, -1)

	return path + "?" + urlVals.Encode(), nil
}

// UpdateDatabaseUserInput holds the input parameters for a updateDatabaseUser operation.
type UpdateDatabaseUserInput struct {
	GroupID                   string
	Username                  string
	UpdateDatabaseUserRequest *UpdateDatabaseUserRequest
}

// Validate returns an error if any of the UpdateDatabaseUserInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i UpdateDatabaseUserInput) Validate() error {

	if i.UpdateDatabaseUserRequest != nil {
		if err := i.UpdateDatabaseUserRequest.Validate(nil); err != nil {
			return err
		}
	}
	return nil
}

// Path returns the URI path for the input.
func (i UpdateDatabaseUserInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/databaseUsers/admin/{username}"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathusername := i.Username
	if pathusername == "" {
		err := fmt.Errorf("username cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{username}", pathusername, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetProcessesInput holds the input parameters for a getProcesses operation.
type GetProcessesInput struct {
	GroupID string
}

// ValidateGetProcessesInput returns an error if the input parameter doesn't
// satisfy the requirements in the swagger yml file.
func ValidateGetProcessesInput(groupID string) error {

	return nil
}

// GetProcessesInputPath returns the URI path for the input.
func GetProcessesInputPath(groupID string) (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/processes"
	urlVals := url.Values{}

	pathgroupID := groupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	return path + "?" + urlVals.Encode(), nil
}

// GetProcessDatabasesInput holds the input parameters for a getProcessDatabases operation.
type GetProcessDatabasesInput struct {
	GroupID      string
	Host         string
	Port         int64
	PageNum      *int64
	ItemsPerPage *int64
}

// Validate returns an error if any of the GetProcessDatabasesInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetProcessDatabasesInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i GetProcessDatabasesInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/databases"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathhost := i.Host
	if pathhost == "" {
		err := fmt.Errorf("host cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{host}", pathhost, -1)

	pathport := strconv.FormatInt(i.Port, 10)
	if pathport == "" {
		err := fmt.Errorf("port cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{port}", pathport, -1)

	if i.PageNum != nil {
		urlVals.Add("pageNum", strconv.FormatInt(*i.PageNum, 10))
	}

	if i.ItemsPerPage != nil {
		urlVals.Add("itemsPerPage", strconv.FormatInt(*i.ItemsPerPage, 10))
	}

	return path + "?" + urlVals.Encode(), nil
}

// GetProcessDatabaseMeasurementsInput holds the input parameters for a getProcessDatabaseMeasurements operation.
type GetProcessDatabaseMeasurementsInput struct {
	GroupID      string
	Host         string
	Port         int64
	DatabaseID   string
	Granularity  string
	Period       *string
	Start        *strfmt.DateTime
	End          *strfmt.DateTime
	M            []string
	PageNum      *int64
	ItemsPerPage *int64
}

// Validate returns an error if any of the GetProcessDatabaseMeasurementsInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetProcessDatabaseMeasurementsInput) Validate() error {

	if err := validate.Enum("granularity", "query", i.Granularity, []interface{}{"PT1M", "PT5M", "PT1H", "P1D"}); err != nil {
		return err
	}

	if i.Start != nil {
		if err := validate.FormatOf("start", "query", "date-time", (*i.Start).String(), strfmt.Default); err != nil {
			return err
		}
	}

	if i.End != nil {
		if err := validate.FormatOf("end", "query", "date-time", (*i.End).String(), strfmt.Default); err != nil {
			return err
		}
	}

	return nil
}

// Path returns the URI path for the input.
func (i GetProcessDatabaseMeasurementsInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/databases/{databaseID}/measurements"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathhost := i.Host
	if pathhost == "" {
		err := fmt.Errorf("host cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{host}", pathhost, -1)

	pathport := strconv.FormatInt(i.Port, 10)
	if pathport == "" {
		err := fmt.Errorf("port cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{port}", pathport, -1)

	pathdatabaseID := i.DatabaseID
	if pathdatabaseID == "" {
		err := fmt.Errorf("databaseID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{databaseID}", pathdatabaseID, -1)

	urlVals.Add("granularity", i.Granularity)

	if i.Period != nil {
		urlVals.Add("period", *i.Period)
	}

	if i.Start != nil {
		urlVals.Add("start", (*i.Start).String())
	}

	if i.End != nil {
		urlVals.Add("end", (*i.End).String())
	}

	for _, v := range i.M {
		urlVals.Add("m", v)
	}

	if i.PageNum != nil {
		urlVals.Add("pageNum", strconv.FormatInt(*i.PageNum, 10))
	}

	if i.ItemsPerPage != nil {
		urlVals.Add("itemsPerPage", strconv.FormatInt(*i.ItemsPerPage, 10))
	}

	return path + "?" + urlVals.Encode(), nil
}

// GetProcessDisksInput holds the input parameters for a getProcessDisks operation.
type GetProcessDisksInput struct {
	GroupID      string
	Host         string
	Port         int64
	PageNum      *int64
	ItemsPerPage *int64
}

// Validate returns an error if any of the GetProcessDisksInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetProcessDisksInput) Validate() error {

	return nil
}

// Path returns the URI path for the input.
func (i GetProcessDisksInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/disks"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathhost := i.Host
	if pathhost == "" {
		err := fmt.Errorf("host cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{host}", pathhost, -1)

	pathport := strconv.FormatInt(i.Port, 10)
	if pathport == "" {
		err := fmt.Errorf("port cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{port}", pathport, -1)

	if i.PageNum != nil {
		urlVals.Add("pageNum", strconv.FormatInt(*i.PageNum, 10))
	}

	if i.ItemsPerPage != nil {
		urlVals.Add("itemsPerPage", strconv.FormatInt(*i.ItemsPerPage, 10))
	}

	return path + "?" + urlVals.Encode(), nil
}

// GetProcessDiskMeasurementsInput holds the input parameters for a getProcessDiskMeasurements operation.
type GetProcessDiskMeasurementsInput struct {
	GroupID      string
	Host         string
	Port         int64
	DiskName     string
	Granularity  string
	Period       *string
	Start        *strfmt.DateTime
	End          *strfmt.DateTime
	M            []string
	PageNum      *int64
	ItemsPerPage *int64
}

// Validate returns an error if any of the GetProcessDiskMeasurementsInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetProcessDiskMeasurementsInput) Validate() error {

	if err := validate.Enum("granularity", "query", i.Granularity, []interface{}{"PT1M", "PT5M", "PT1H", "P1D"}); err != nil {
		return err
	}

	if i.Start != nil {
		if err := validate.FormatOf("start", "query", "date-time", (*i.Start).String(), strfmt.Default); err != nil {
			return err
		}
	}

	if i.End != nil {
		if err := validate.FormatOf("end", "query", "date-time", (*i.End).String(), strfmt.Default); err != nil {
			return err
		}
	}

	return nil
}

// Path returns the URI path for the input.
func (i GetProcessDiskMeasurementsInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/disks/{diskName}/measurements"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathhost := i.Host
	if pathhost == "" {
		err := fmt.Errorf("host cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{host}", pathhost, -1)

	pathport := strconv.FormatInt(i.Port, 10)
	if pathport == "" {
		err := fmt.Errorf("port cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{port}", pathport, -1)

	pathdiskName := i.DiskName
	if pathdiskName == "" {
		err := fmt.Errorf("diskName cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{diskName}", pathdiskName, -1)

	urlVals.Add("granularity", i.Granularity)

	if i.Period != nil {
		urlVals.Add("period", *i.Period)
	}

	if i.Start != nil {
		urlVals.Add("start", (*i.Start).String())
	}

	if i.End != nil {
		urlVals.Add("end", (*i.End).String())
	}

	for _, v := range i.M {
		urlVals.Add("m", v)
	}

	if i.PageNum != nil {
		urlVals.Add("pageNum", strconv.FormatInt(*i.PageNum, 10))
	}

	if i.ItemsPerPage != nil {
		urlVals.Add("itemsPerPage", strconv.FormatInt(*i.ItemsPerPage, 10))
	}

	return path + "?" + urlVals.Encode(), nil
}

// GetProcessMeasurementsInput holds the input parameters for a getProcessMeasurements operation.
type GetProcessMeasurementsInput struct {
	GroupID      string
	Host         string
	Port         int64
	Granularity  string
	Period       *string
	Start        *strfmt.DateTime
	End          *strfmt.DateTime
	M            []string
	PageNum      *int64
	ItemsPerPage *int64
}

// Validate returns an error if any of the GetProcessMeasurementsInput parameters don't satisfy the
// requirements from the swagger yml file.
func (i GetProcessMeasurementsInput) Validate() error {

	if err := validate.Enum("granularity", "query", i.Granularity, []interface{}{"PT1M", "PT5M", "PT1H", "P1D"}); err != nil {
		return err
	}

	if i.Start != nil {
		if err := validate.FormatOf("start", "query", "date-time", (*i.Start).String(), strfmt.Default); err != nil {
			return err
		}
	}

	if i.End != nil {
		if err := validate.FormatOf("end", "query", "date-time", (*i.End).String(), strfmt.Default); err != nil {
			return err
		}
	}

	return nil
}

// Path returns the URI path for the input.
func (i GetProcessMeasurementsInput) Path() (string, error) {
	path := "/api/atlas/v1.0/groups/{groupID}/processes/{host}:{port}/measurements"
	urlVals := url.Values{}

	pathgroupID := i.GroupID
	if pathgroupID == "" {
		err := fmt.Errorf("groupID cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{groupID}", pathgroupID, -1)

	pathhost := i.Host
	if pathhost == "" {
		err := fmt.Errorf("host cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{host}", pathhost, -1)

	pathport := strconv.FormatInt(i.Port, 10)
	if pathport == "" {
		err := fmt.Errorf("port cannot be empty because it's a path parameter")
		if err != nil {
			return "", err
		}
	}
	path = strings.Replace(path, "{port}", pathport, -1)

	urlVals.Add("granularity", i.Granularity)

	if i.Period != nil {
		urlVals.Add("period", *i.Period)
	}

	if i.Start != nil {
		urlVals.Add("start", (*i.Start).String())
	}

	if i.End != nil {
		urlVals.Add("end", (*i.End).String())
	}

	for _, v := range i.M {
		urlVals.Add("m", v)
	}

	if i.PageNum != nil {
		urlVals.Add("pageNum", strconv.FormatInt(*i.PageNum, 10))
	}

	if i.ItemsPerPage != nil {
		urlVals.Add("itemsPerPage", strconv.FormatInt(*i.ItemsPerPage, 10))
	}

	return path + "?" + urlVals.Encode(), nil
}
