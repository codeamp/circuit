// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// Granularity granularity
// swagger:model Granularity
type Granularity string

const (
	// GranularityPT1M captures enum value "PT1M"
	GranularityPT1M Granularity = "PT1M"
	// GranularityPT5M captures enum value "PT5M"
	GranularityPT5M Granularity = "PT5M"
	// GranularityPT1H captures enum value "PT1H"
	GranularityPT1H Granularity = "PT1H"
	// GranularityP1D captures enum value "P1D"
	GranularityP1D Granularity = "P1D"
)

// for schema
var granularityEnum []interface{}

func init() {
	var res []Granularity
	if err := json.Unmarshal([]byte(`["PT1M","PT5M","PT1H","P1D"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		granularityEnum = append(granularityEnum, v)
	}
}

func (m Granularity) validateGranularityEnum(path, location string, value Granularity) error {
	if err := validate.Enum(path, location, value, granularityEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this granularity
func (m Granularity) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateGranularityEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
