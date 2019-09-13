package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateJSONFormatFailsForInvalidJSON(t *testing.T) {
	errors := ValidateJSONFormat(`["connection lost", "invalid request"]`)
	assert.IsType(t, &InvalidJSONError{}, errors)
}

func TestValidateJSONFormatFailsForMissingRequiredFields(t *testing.T) {
	assert.Equal(
		t,
		ValidateJSONFormat(`{"title":"an_event", "level":"error"}`),
		&MissingRequiredFieldError{Field: "source"},
	)

	assert.Equal(
		t,
		ValidateJSONFormat(`{"source":"an-app", "level":"error"}`),
		&MissingRequiredFieldError{Field: "title"},
	)

	assert.Equal(
		t,
		ValidateJSONFormat(`{"source":"an-app", "title":"an_event"}`),
		&MissingRequiredFieldError{Field: "level"},
	)

	assert.Equal(
		t,
		ValidateJSONFormat(`{"source":"an-app", "title":"an_event", "level":"error", "type":"gauge"}`),
		&MissingRequiredFieldError{Field: "value"},
	)
}

func TestValidateJSONFormatFailsForInvalidLogType(t *testing.T) {
	assert.Equal(
		t,
		ValidateJSONFormat(`{
			"source":"an-app",
			"title":"an_event",
			"level":"error",
			"type":100
		}`),
		&InvalidValueTypeError{Field: "type", ExpectedType: StringType},
	)
}

func TestValidateJSONFormatFailsForInvalidNumericValueType(t *testing.T) {
	assert.Equal(
		t,
		ValidateJSONFormat(`{
			"source":"an-app",
			"title":"an_event",
			"level":"error",
			"type":"gauge",
			"value":"500"
		}`),
		&InvalidValueTypeError{Field: "value", ExpectedType: NumberType},
	)
}

func TestValidateJSONFormatFailsForInvalidLogLevel(t *testing.T) {
	assert.Equal(
		t,
		ValidateJSONFormat(`{"source":"an-app", "title":"an_event", "level":"shlerror"}`),
		&InvalidValueError{Field: "level", Value: "shlerror"},
	)
}

func TestValidateJSONFormatPassesForValidLogLines(t *testing.T) {
	assert.NoError(t, ValidateJSONFormat(`{"source":"an-app", "title":"an_event", "level":"error"}`))

	assert.NoError(
		t,
		ValidateJSONFormat(`{
			"source":"an-app",
			"title":"an_event",
			"level":"error",
			"type":"counter"
		}`),
	)

	// Also testing graceful handling of leading/trailing whitespace here.
	assert.NoError(
		t,
		ValidateJSONFormat(`	  {
			"source":"an-app",
			"title":"an_event",
			"level":"error",
			"type":"gauge",
			"value":24
		}

		`,
		),
	)
}
