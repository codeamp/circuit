// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// CreateDatabaseUserRequest create database user request
// swagger:model CreateDatabaseUserRequest
type CreateDatabaseUserRequest struct {
	DatabaseUser

	// password
	Password string `json:"password,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *CreateDatabaseUserRequest) UnmarshalJSON(raw []byte) error {

	var aO0 DatabaseUser
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.DatabaseUser = aO0

	var data struct {
		Password string `json:"password,omitempty"`
	}
	if err := swag.ReadJSON(raw, &data); err != nil {
		return err
	}

	m.Password = data.Password

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m CreateDatabaseUserRequest) MarshalJSON() ([]byte, error) {
	var _parts [][]byte

	aO0, err := swag.WriteJSON(m.DatabaseUser)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)

	var data struct {
		Password string `json:"password,omitempty"`
	}

	data.Password = m.Password

	jsonData, err := swag.WriteJSON(data)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, jsonData)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this create database user request
func (m *CreateDatabaseUserRequest) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.DatabaseUser.Validate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// MarshalBinary interface implementation
func (m *CreateDatabaseUserRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CreateDatabaseUserRequest) UnmarshalBinary(b []byte) error {
	var res CreateDatabaseUserRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
