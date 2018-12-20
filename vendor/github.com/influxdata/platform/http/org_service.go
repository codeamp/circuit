package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/influxdata/platform"
	kerrors "github.com/influxdata/platform/kit/errors"
	"github.com/julienschmidt/httprouter"
)

// OrgHandler represents an HTTP API handler for orgs.
type OrgHandler struct {
	*httprouter.Router

	OrganizationService             platform.OrganizationService
	OrganizationOperationLogService platform.OrganizationOperationLogService
	BucketService                   platform.BucketService
	UserResourceMappingService      platform.UserResourceMappingService
}

const (
	organizationsPath            = "/api/v2/orgs"
	organizationsIDPath          = "/api/v2/orgs/:id"
	organizationsIDLogPath       = "/api/v2/orgs/:id/log"
	organizationsIDMembersPath   = "/api/v2/orgs/:id/members"
	organizationsIDMembersIDPath = "/api/v2/orgs/:id/members/:organizationID"
	organizationsIDOwnersPath    = "/api/v2/orgs/:id/owners"
	organizationsIDOwnersIDPath  = "/api/v2/orgs/:id/owners/:organizationID"
)

// NewOrgHandler returns a new instance of OrgHandler.
func NewOrgHandler(mappingService platform.UserResourceMappingService) *OrgHandler {
	h := &OrgHandler{
		Router:                     httprouter.New(),
		UserResourceMappingService: mappingService,
	}

	h.HandlerFunc("POST", organizationsPath, h.handlePostOrg)
	h.HandlerFunc("GET", organizationsPath, h.handleGetOrgs)
	h.HandlerFunc("GET", organizationsIDPath, h.handleGetOrg)
	h.HandlerFunc("GET", organizationsIDLogPath, h.handleGetOrgLog)
	h.HandlerFunc("PATCH", organizationsIDPath, h.handlePatchOrg)
	h.HandlerFunc("DELETE", organizationsIDPath, h.handleDeleteOrg)

	h.HandlerFunc("POST", organizationsIDMembersPath, newPostMemberHandler(h.UserResourceMappingService, platform.OrgResourceType, platform.Member))
	h.HandlerFunc("GET", organizationsIDMembersPath, newGetMembersHandler(h.UserResourceMappingService, platform.Member))
	h.HandlerFunc("DELETE", organizationsIDMembersIDPath, newDeleteMemberHandler(h.UserResourceMappingService, platform.Member))

	h.HandlerFunc("POST", organizationsIDOwnersPath, newPostMemberHandler(h.UserResourceMappingService, platform.OrgResourceType, platform.Owner))
	h.HandlerFunc("GET", organizationsIDOwnersPath, newGetMembersHandler(h.UserResourceMappingService, platform.Owner))
	h.HandlerFunc("DELETE", organizationsIDOwnersIDPath, newDeleteMemberHandler(h.UserResourceMappingService, platform.Owner))

	return h
}

type orgsResponse struct {
	Links         map[string]string `json:"links"`
	Organizations []*orgResponse    `json:"orgs"`
}

func (o orgsResponse) ToPlatform() []*platform.Organization {
	orgs := make([]*platform.Organization, len(o.Organizations))
	for i := range o.Organizations {
		orgs[i] = &o.Organizations[i].Organization
	}
	return orgs
}

func newOrgsResponse(orgs []*platform.Organization) *orgsResponse {
	res := orgsResponse{
		Links: map[string]string{
			"self": "/api/v2/orgs",
		},
		Organizations: []*orgResponse{},
	}
	for _, org := range orgs {
		res.Organizations = append(res.Organizations, newOrgResponse(org))
	}
	return &res
}

type orgResponse struct {
	Links map[string]string `json:"links"`
	platform.Organization
}

func newOrgResponse(o *platform.Organization) *orgResponse {
	return &orgResponse{
		Links: map[string]string{
			"self":       fmt.Sprintf("/api/v2/orgs/%s", o.ID),
			"log":        fmt.Sprintf("/api/v2/orgs/%s/log", o.ID),
			"members":    fmt.Sprintf("/api/v2/orgs/%s/members", o.ID),
			"buckets":    fmt.Sprintf("/api/v2/buckets?org=%s", o.Name),
			"tasks":      fmt.Sprintf("/api/v2/tasks?org=%s", o.Name),
			"dashboards": fmt.Sprintf("/api/v2/dashboards?org=%s", o.Name),
		},
		Organization: *o,
	}
}

// handlePostOrg is the HTTP handler for the POST /api/v2/orgs route.
func (h *OrgHandler) handlePostOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodePostOrgRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := h.OrganizationService.CreateOrganization(ctx, req.Org); err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusCreated, newOrgResponse(req.Org)); err != nil {
		EncodeError(ctx, err, w)
		return
	}
}

type postOrgRequest struct {
	Org *platform.Organization
}

func decodePostOrgRequest(ctx context.Context, r *http.Request) (*postOrgRequest, error) {
	o := &platform.Organization{}
	if err := json.NewDecoder(r.Body).Decode(o); err != nil {
		return nil, err
	}

	return &postOrgRequest{
		Org: o,
	}, nil
}

// handleGetOrg is the HTTP handler for the GET /api/v2/orgs/:id route.
func (h *OrgHandler) handleGetOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeGetOrgRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	b, err := h.OrganizationService.FindOrganizationByID(ctx, req.OrgID)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, newOrgResponse(b)); err != nil {
		EncodeError(ctx, err, w)
		return
	}
}

type getOrgRequest struct {
	OrgID platform.ID
}

func decodeGetOrgRequest(ctx context.Context, r *http.Request) (*getOrgRequest, error) {
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return nil, kerrors.InvalidDataf("url missing id")
	}

	var i platform.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	req := &getOrgRequest{
		OrgID: i,
	}

	return req, nil
}

// handleGetOrgs is the HTTP handler for the GET /api/v2/orgs route.
func (h *OrgHandler) handleGetOrgs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeGetOrgsRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	orgs, _, err := h.OrganizationService.FindOrganizations(ctx, req.filter)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, newOrgsResponse(orgs)); err != nil {
		EncodeError(ctx, err, w)
		return
	}
}

type getOrgsRequest struct {
	filter platform.OrganizationFilter
}

func decodeGetOrgsRequest(ctx context.Context, r *http.Request) (*getOrgsRequest, error) {
	qp := r.URL.Query()
	req := &getOrgsRequest{}

	if orgID := qp.Get("id"); orgID != "" {
		id, err := platform.IDFromString(orgID)
		if err != nil {
			return nil, err
		}
		req.filter.ID = id
	}

	if name := qp.Get("name"); name != "" {
		req.filter.Name = &name
	}

	return req, nil
}

// handleDeleteOrganization is the HTTP handler for the DELETE /api/v2/orgs/:id route.
func (h *OrgHandler) handleDeleteOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeDeleteOrganizationRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := h.OrganizationService.DeleteOrganization(ctx, req.OrganizationID); err != nil {
		EncodeError(ctx, err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type deleteOrganizationRequest struct {
	OrganizationID platform.ID
}

func decodeDeleteOrganizationRequest(ctx context.Context, r *http.Request) (*deleteOrganizationRequest, error) {
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return nil, kerrors.InvalidDataf("url missing id")
	}

	var i platform.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}
	req := &deleteOrganizationRequest{
		OrganizationID: i,
	}

	return req, nil
}

// handlePatchOrg is the HTTP handler for the PATH /api/v2/orgs route.
func (h *OrgHandler) handlePatchOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodePatchOrgRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	o, err := h.OrganizationService.UpdateOrganization(ctx, req.OrgID, req.Update)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, newOrgResponse(o)); err != nil {
		EncodeError(ctx, err, w)
		return
	}
}

type patchOrgRequest struct {
	Update platform.OrganizationUpdate
	OrgID  platform.ID
}

func decodePatchOrgRequest(ctx context.Context, r *http.Request) (*patchOrgRequest, error) {
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return nil, kerrors.InvalidDataf("url missing id")
	}

	var i platform.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	var upd platform.OrganizationUpdate
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		return nil, err
	}

	return &patchOrgRequest{
		Update: upd,
		OrgID:  i,
	}, nil
}

const (
	organizationPath = "/api/v2/orgs"
)

// OrganizationService connects to Influx via HTTP using tokens to manage organizations.
type OrganizationService struct {
	Addr               string
	Token              string
	InsecureSkipVerify bool
}

// FindOrganizationByID gets a single organization with a given id using HTTP.
func (s *OrganizationService) FindOrganizationByID(ctx context.Context, id platform.ID) (*platform.Organization, error) {
	filter := platform.OrganizationFilter{ID: &id}
	return s.FindOrganization(ctx, filter)
}

// FindOrganization gets a single organization matching the filter using HTTP.
func (s *OrganizationService) FindOrganization(ctx context.Context, filter platform.OrganizationFilter) (*platform.Organization, error) {
	os, n, err := s.FindOrganizations(ctx, filter)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFound
	}

	return os[0], nil
}

// FindOrganizations returns all organizations that match the filter via HTTP.
func (s *OrganizationService) FindOrganizations(ctx context.Context, filter platform.OrganizationFilter, opt ...platform.FindOptions) ([]*platform.Organization, int, error) {
	url, err := newURL(s.Addr, organizationPath)
	if err != nil {
		return nil, 0, err
	}
	qp := url.Query()

	if filter.Name != nil {
		qp.Add("name", *filter.Name)
	}
	if filter.ID != nil {
		qp.Add("id", filter.ID.String())
	}
	url.RawQuery = qp.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, 0, err
	}

	SetToken(s.Token, req)
	hc := newClient(url.Scheme, s.InsecureSkipVerify)

	resp, err := hc.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if err := CheckError(resp); err != nil {
		return nil, 0, err
	}

	var os orgsResponse
	if err := json.NewDecoder(resp.Body).Decode(&os); err != nil {
		return nil, 0, err
	}

	orgs := os.ToPlatform()
	return orgs, len(orgs), nil

}

// CreateOrganization creates an organization.
func (s *OrganizationService) CreateOrganization(ctx context.Context, o *platform.Organization) error {
	if o.Name == "" {
		return kerrors.InvalidDataf("organization name is required")
	}

	url, err := newURL(s.Addr, organizationPath)
	if err != nil {
		return err
	}

	octets, err := json.Marshal(o)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url.String(), bytes.NewReader(octets))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	SetToken(s.Token, req)

	hc := newClient(url.Scheme, s.InsecureSkipVerify)

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	// TODO(jsternberg): Should this check for a 201 explicitly?
	if err := CheckError(resp); err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(o); err != nil {
		return err
	}

	return nil
}

// UpdateOrganization updates the organization over HTTP.
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id platform.ID, upd platform.OrganizationUpdate) (*platform.Organization, error) {
	u, err := newURL(s.Addr, organizationIDPath(id))
	if err != nil {
		return nil, err
	}

	octets, err := json.Marshal(upd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", u.String(), bytes.NewReader(octets))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	SetToken(s.Token, req)

	hc := newClient(u.Scheme, s.InsecureSkipVerify)

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}

	if err := CheckError(resp); err != nil {
		return nil, err
	}

	var o platform.Organization
	if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &o, nil
}

// DeleteOrganization removes organization id over HTTP.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id platform.ID) error {
	u, err := newURL(s.Addr, organizationIDPath(id))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}
	SetToken(s.Token, req)

	hc := newClient(u.Scheme, s.InsecureSkipVerify)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}

	return CheckErrorStatus(http.StatusNoContent, resp)
}

func organizationIDPath(id platform.ID) string {
	return path.Join(organizationPath, id.String())
}

// hanldeGetOrganizationLog retrieves a organization log by the organizations ID.
func (h *OrgHandler) handleGetOrgLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeGetOrganizationLogRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	log, _, err := h.OrganizationOperationLogService.GetOrganizationOperationLog(ctx, req.OrganizationID, req.opts)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	if err := encodeResponse(ctx, w, http.StatusOK, newOrganizationLogResponse(req.OrganizationID, log)); err != nil {
		EncodeError(ctx, err, w)
		return
	}
}

type getOrganizationLogRequest struct {
	OrganizationID platform.ID
	opts           platform.FindOptions
}

func decodeGetOrganizationLogRequest(ctx context.Context, r *http.Request) (*getOrganizationLogRequest, error) {
	params := httprouter.ParamsFromContext(ctx)
	id := params.ByName("id")
	if id == "" {
		return nil, kerrors.InvalidDataf("url missing id")
	}

	var i platform.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	opts := platform.DefaultOperationLogFindOptions
	qp := r.URL.Query()
	if v := qp.Get("desc"); v == "false" {
		opts.Descending = false
	}
	if v := qp.Get("limit"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		opts.Limit = i
	}
	if v := qp.Get("offset"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		opts.Offset = i
	}

	return &getOrganizationLogRequest{
		OrganizationID: i,
		opts:           opts,
	}, nil
}

func newOrganizationLogResponse(id platform.ID, es []*platform.OperationLogEntry) *operationLogResponse {
	log := make([]*operationLogEntryResponse, 0, len(es))
	for _, e := range es {
		log = append(log, newOperationLogEntryResponse(e))
	}
	return &operationLogResponse{
		Links: map[string]string{
			"self": fmt.Sprintf("/api/v2/organizations/%s/log", id),
		},
		Log: log,
	}
}
