package router

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SortableRules []Rule

func (r SortableRules) Len() int {
	return len(r)
}
func (r SortableRules) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}
func (r SortableRules) Swap(i, j int) {
	tmp := r[j]
	r[j] = r[i]
	r[i] = tmp
}

func TestParsesWellFormatedConfig(t *testing.T) {
	conf := []byte(`
routes:
  rule-one:
    matchers:
      title: ["authorize-app", true]
    output:
      type: "notifications"
      channel: "%{foo.bar}"
      icon: ":rocket:"
      message: "authorized %{foo.bar} in ${SCHOOL}"
      user: "@fishman"
  rule-two:
    matchers:
      foo.bar: ["multiple", "matches"]
      baz: ["whatever"]
    output:
      type: "alerts"
      series: "other-series"
      dimensions: ["baz"]
      stat_type: "gauge"
  rule-three:
    matchers:
      foo.bar: ["multiple", "matches"]
      baz: ["whatever"]
    output:
      type: "alerts"
      series: "other-series"
      dimensions: ["baz"]
      stat_type: "counter"
      value_field: "hello"
  rule-four:
    matchers:
      foo.bar: ["multiple", "matches"]
      baz: ["whatever"]
    output:
      type: "alerts"
      series: "other-series"
      dimensions: []
      stat_type: "counter"
  rule-five:
    matchers:
      foo.bar: [true]
      baz: [false]
    output:
      type: "alerts"
      series: "other-series"
      dimensions: []
      stat_type: "gauge"
`)
	expected := SortableRules{
		Rule{
			Name:     "rule-one",
			Matchers: RuleMatchers{"title": []string{"authorize-app", "true"}},
			Output: RuleOutput{
				"type":    "notifications",
				"channel": `%{foo.bar}`,
				"icon":    ":rocket:",
				"message": `authorized %{foo.bar} in Hogwarts`,
				"user":    "@fishman",
			},
		},
		Rule{
			Name: "rule-two",
			Matchers: RuleMatchers{
				"foo.bar": []string{"multiple", "matches"},
				"baz":     []string{"whatever"},
			},
			Output: RuleOutput{
				"type":        "alerts",
				"series":      "other-series",
				"dimensions":  []interface{}{"baz"},
				"stat_type":   "gauge",
				"value_field": "value",
			},
		},
		Rule{
			Name: "rule-three",
			Matchers: RuleMatchers{
				"foo.bar": []string{"multiple", "matches"},
				"baz":     []string{"whatever"},
			},
			Output: RuleOutput{
				"type":        "alerts",
				"series":      "other-series",
				"dimensions":  []interface{}{"baz"},
				"stat_type":   "counter",
				"value_field": "hello",
			},
		},
		Rule{
			Name: "rule-four",
			Matchers: RuleMatchers{
				"foo.bar": []string{"multiple", "matches"},
				"baz":     []string{"whatever"},
			},
			Output: RuleOutput{
				"type":        "alerts",
				"series":      "other-series",
				"dimensions":  []interface{}{},
				"stat_type":   "counter",
				"value_field": "value",
			},
		},
		Rule{
			Name: "rule-five",
			Matchers: RuleMatchers{
				"foo.bar": []string{"true"},
				"baz":     []string{"false"},
			},
			Output: RuleOutput{
				"type":        "alerts",
				"series":      "other-series",
				"dimensions":  []interface{}{},
				"stat_type":   "gauge",
				"value_field": "value",
			},
		},
	}

	err := os.Setenv("SCHOOL", "Hogwarts")
	assert.Nil(t, err)

	router, err := NewFromConfigBytes(conf)
	assert.Nil(t, err)

	r, ok := router.(*RuleRouter)
	assert.True(t, ok)
	actual := SortableRules(r.rules)
	sort.Sort(expected)
	sort.Sort(actual)
	assert.Equal(t, expected, actual)
}

func TestOnlyNonemptyStringMatcherValues(t *testing.T) {
	confTmpl := `
routes:
  non-string-values:
    matchers:
      no-numbers: [%s]
    output:
      type: "analytics"
      series: "fun"
`

	// Make sure the template works
	conf := []byte(fmt.Sprintf(confTmpl, "\"valid\""))
	_, err := NewFromConfigBytes(conf)
	assert.Nil(t, err)

	for _, invalidVal := range []string{"5", "[]", "{}", `""`} {
		conf := []byte(fmt.Sprintf(confTmpl, invalidVal))
		_, err := NewFromConfigBytes(conf)
		assert.Error(t, err)
	}
}

func TestNoSpecialsInMatcher(t *testing.T) {
	confFieldTmpl := `
routes:
  complicated-fields:
    matchers:
      "%s": ["hallo?"]
    output:
      type: "analytics"
      series: "fun"
`
	confValTmpl := `
routes:
  complicated-values:
    matchers:
      title: ["%s"]
    output:
      type: "analytics"
      series: "fun"
`

	// Make sure templates work
	for _, tmpl := range []string{confFieldTmpl, confValTmpl} {
		conf := []byte(fmt.Sprintf(tmpl, "valid"))
		_, err := NewFromConfigBytes(conf)
		assert.Nil(t, err)
	}

	invalids := []string{"${wut}", "%{wut}", "$huh", "}ok?", "nope{", `100% fail`}
	for _, invalid := range invalids {
		for _, tmpl := range []string{confFieldTmpl, confValTmpl} {
			conf := []byte(fmt.Sprintf(tmpl, invalid))
			_, err := NewFromConfigBytes(conf)
			assert.Error(t, err)
		}
	}
}

func TestNoDupMatchers(t *testing.T) {
	confTmpl := `
routes:
  sloppy:
    matchers:
      title: [%s]
    output:
      type: "analytics"
      series: "fun"
`

	validConf := []byte(fmt.Sprintf(confTmpl, `"non-repeated", "name"`))
	_, err := NewFromConfigBytes(validConf)
	assert.Nil(t, err)

	invalidConf := []byte(fmt.Sprintf(confTmpl, `"repeated", "repeated", "name"`))
	_, err = NewFromConfigBytes(invalidConf)
	assert.Error(t, err)
}

func TestErrorsThrownWithTypeOCaught(t *testing.T) {
	assert := assert.New(t)

	config := `
route: # Shouldn't routes (plural)
  string-values:
    matchers:
      errors: [ "type-o" ]
    output:
      type: "analytics"
      series: "fun"
`
	_, err := NewFromConfigBytes([]byte(config))
	assert.Error(err)

	config = `
routes:
  string-values:
    matcher: # Shouldn't matches (plural)
      errors: [ "type-o" ]
    output:
      type: "analytics"
      series: "fun"
`
	_, err = NewFromConfigBytes([]byte(config))
	assert.Error(err)

	config = `
routes:
  string-values:
    matchers:
      errors: [ "type-o" ]
    outputs: # Should be output (singular)
      type: "analytics"
      series: "fun"
`
	_, err = NewFromConfigBytes([]byte(config))
	assert.Error(err)

	config = `
routes:
  $invalid-string-values: # Invalid rule name
    matchers:
      errors: [ "type-o" ]
    output:
      type: "analytics"
      series: "fun"
`
	_, err = NewFromConfigBytes([]byte(config))
	assert.Error(err)

	config = `
routes:
  string-values:
    matchers:
      errors: [ "*", "type-o" ] # A wildcard cannot exist with other matchers
    output:
      type: "analytics"
      series: "fun"
`
	_, err = NewFromConfigBytes([]byte(config))
	assert.Error(err)
}

func TestOutputRequiresCorrectTypes(t *testing.T) {
	confTmpl := `
routes:
  wrong:
    matchers:
      title: ["test"]
    output:
      type: "alerts"
      series: %s
      dimensions: %s
      value_field: "hello"
      stat_type: "gauge"
`

	validConf := []byte(fmt.Sprintf(confTmpl, `"my-series"`, `["dim1", "dim2"]`))
	_, err := NewFromConfigBytes(validConf)
	assert.Nil(t, err)

	invalidConf0 := []byte(fmt.Sprintf(confTmpl, `["my-series"]`, `["dim1", "dim2"]`))
	_, err = NewFromConfigBytes(invalidConf0)
	assert.Error(t, err)

	invalidConf1 := []byte(fmt.Sprintf(confTmpl, `"my-series"`, `"dim1"`))
	_, err = NewFromConfigBytes(invalidConf1)
	assert.Error(t, err)
}

func TestOutputRequiresAllKeys(t *testing.T) {
	confTmpl := `
routes:
  wrong:
    matchers:
      title: ["test"]
    output:
      type: "alerts"%s
      dimensions: ["dim1", "dim2"]
      stat_type: "gauge"
`

	validConf := []byte(fmt.Sprintf(confTmpl, `
      series: "whatever"`))
	_, err := NewFromConfigBytes(validConf)
	assert.Nil(t, err)

	invalidConf := []byte(fmt.Sprintf(confTmpl, ``))
	_, err = NewFromConfigBytes(invalidConf)
	assert.Error(t, err)
}

func TestOutputNoExtraKeysAllowed(t *testing.T) {
	confTmpl := `
routes:
  wrong:
    matchers:
      title: ["test"]
    output:
      type: "metrics"%s
      dimensions: ["dim1", "dim2"]
      value_field: "hihi"
`

	validConf := []byte(fmt.Sprintf(confTmpl, `
      series: "whatever"`))
	_, err := NewFromConfigBytes(validConf)
	assert.Nil(t, err)

	invalidConf := []byte(fmt.Sprintf(confTmpl, `
      series: "whatever"
      something-else: "hi there"`))
	_, err = NewFromConfigBytes(invalidConf)
	assert.Error(t, err)
}
