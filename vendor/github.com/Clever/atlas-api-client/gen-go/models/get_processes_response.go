// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// GetProcessesResponse get processes response
// swagger:model GetProcessesResponse
type GetProcessesResponse struct {

	// links
	Links []*Link `json:"links"`

	// results
	Results []*Process `json:"results"`

	// total count
	TotalCount int64 `json:"totalCount,omitempty"`
}

// Validate validates this get processes response
func (m *GetProcessesResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLinks(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateResults(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *GetProcessesResponse) validateLinks(formats strfmt.Registry) error {

	if swag.IsZero(m.Links) { // not required
		return nil
	}

	for i := 0; i < len(m.Links); i++ {

		if swag.IsZero(m.Links[i]) { // not required
			continue
		}

		if m.Links[i] != nil {

			if err := m.Links[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("links" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *GetProcessesResponse) validateResults(formats strfmt.Registry) error {

	if swag.IsZero(m.Results) { // not required
		return nil
	}

	for i := 0; i < len(m.Results); i++ {

		if swag.IsZero(m.Results[i]) { // not required
			continue
		}

		if m.Results[i] != nil {

			if err := m.Results[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("results" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *GetProcessesResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GetProcessesResponse) UnmarshalBinary(b []byte) error {
	var res GetProcessesResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
