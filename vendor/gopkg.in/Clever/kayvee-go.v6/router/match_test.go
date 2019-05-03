package router

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SortableOutputs []map[string]interface{}

func (r SortableOutputs) Len() int {
	return len(r)
}
func (r SortableOutputs) Less(i, j int) bool {
	return r[i]["rule"].(string) < r[j]["rule"].(string)
}
func (r SortableOutputs) Swap(i, j int) {
	tmp := r[j]
	r[j] = r[i]
	r[i] = tmp
}

func TestMatchesSimple(t *testing.T) {
	r := Rule{
		Matchers: RuleMatchers{
			"title": []string{"hello", "hi"},
			"foo":   []string{"bar"},
		},
		Output: RuleOutput{},
	}
	msg0 := map[string]interface{}{
		"title": "hello",
		"foo":   "bar",
	}
	msg1 := map[string]interface{}{
		"title": "hi",
		"foo":   "bar",
	}
	msg2 := map[string]interface{}{
		"title": "hi",
		"foo":   "fighters",
	}
	msg3 := map[string]interface{}{
		"title": "howdy",
		"foo":   "bar",
	}
	msg4 := map[string]interface{}{
		"missing-stuff": "indeed",
	}
	assert.True(t, r.Matches(msg0))
	assert.True(t, r.Matches(msg1))
	assert.False(t, r.Matches(msg2))
	assert.False(t, r.Matches(msg3))
	assert.False(t, r.Matches(msg4))
}

func TestMatchesNested(t *testing.T) {
	r := Rule{
		Matchers: RuleMatchers{
			"foo.bar": []string{"hello", "hi"},
		},
		Output: RuleOutput{},
	}
	msg0 := map[string]interface{}{
		"title": "greeting",
		"foo": map[string]interface{}{
			"bar": "hello",
		},
	}
	msg1 := map[string]interface{}{
		"title": "greeting",
		"foo": map[string]interface{}{
			"bar": "hi",
		},
	}
	msg2 := map[string]interface{}{
		"title": "greeting",
		"foo": map[string]interface{}{
			"bar": "howdy",
		},
	}
	msg3 := map[string]interface{}{
		"title": "greeting",
		"foo": map[string]interface{}{
			"baz": "howdy",
		},
	}
	msg4 := map[string]interface{}{
		"title": "greeting",
		"boo": map[string]interface{}{
			"bar": "howdy",
		},
	}
	assert.True(t, r.Matches(msg0))
	assert.True(t, r.Matches(msg1))
	assert.False(t, r.Matches(msg2))
	assert.False(t, r.Matches(msg3))
	assert.False(t, r.Matches(msg4))
}

func TestWildcardMatches(t *testing.T) {
	assert := assert.New(t)
	r := Rule{
		Matchers: RuleMatchers{"any": []string{"*"}},
		Output:   RuleOutput{},
	}

	tests := []struct {
		Description string
		Message     map[string]interface{}
		DoesMatch   bool
	}{
		{
			Description: "Matches any bool",
			Message:     map[string]interface{}{"any": false},
			DoesMatch:   true,
		},
		{
			Description: "Matches any number",
			Message:     map[string]interface{}{"any": 5},
			DoesMatch:   true,
		},
		{
			Description: "Matches any string",
			Message:     map[string]interface{}{"any": "hello"},
			DoesMatch:   true,
		},
		{
			Description: "Matches any object",
			Message: map[string]interface{}{
				"any": map[string]interface{}{
					"baz": "howdy",
				},
			},
			DoesMatch: true,
		},
		{
			Description: "Does not matches empty string",
			Message:     map[string]interface{}{"any": ""},
			DoesMatch:   false,
		},
		{
			Description: "Does not matches nil",
			Message:     map[string]interface{}{"any": nil},
			DoesMatch:   false,
		},
		{
			Description: "Does not match message without correct field",
			Message: map[string]interface{}{
				"title": "greeting",
				"boo": map[string]interface{}{
					"bar": "howdy",
				},
			},
			DoesMatch: false,
		},
	}
	for _, test := range tests {
		t.Log(test.Description)
		if test.DoesMatch {
			assert.True(r.Matches(test.Message))
		} else {
			assert.False(r.Matches(test.Message))
		}
	}
}

func TestBooleanMatches(t *testing.T) {
	assert := assert.New(t)
	r := Rule{
		Matchers: RuleMatchers{"bull": []string{"true"}},
		Output:   RuleOutput{},
	}

	tests := []struct {
		Description string
		Message     map[string]interface{}
		DoesMatch   bool
	}{
		{
			Description: "Simple match",
			Message:     map[string]interface{}{"bull": true},
			DoesMatch:   true,
		},
		{
			Description: "Match with multiple fields",
			Message:     map[string]interface{}{"any": false, "bull": true},
			DoesMatch:   true,
		},
		{
			Description: "Bool that doesn't match",
			Message:     map[string]interface{}{"bull": false},
			DoesMatch:   false,
		},
		{
			Description: "Messsge that doesn't have correct field",
			Message: map[string]interface{}{
				"title": "greeting",
				"foo":   map[string]string{"bar": "howdy"},
			},
			DoesMatch: false,
		},
	}
	for _, test := range tests {
		t.Log(test.Description)
		if test.DoesMatch {
			assert.True(r.Matches(test.Message))
		} else {
			assert.False(r.Matches(test.Message))
		}
	}
}

func TestSubstitution(t *testing.T) {
	r := Rule{
		Name:     "myrule",
		Matchers: RuleMatchers{},
		Output: RuleOutput{
			"channel":    "#-%{foo}-",
			"dimensions": []string{"-%{foo}-", "-%{bar.baz}-"},
			"msg": "%{an-int}, %{an-int32}, %{an-int64}, " +
				"%{a-bool}, %{a-float32}, %{a-float64}, %{a-string}, %{bar}",
		},
	}
	msg := map[string]interface{}{
		"title":     "greeting",
		"foo":       "partner",
		"an-int":    int(100),
		"an-int32":  int(132),
		"an-int64":  int(164),
		"a-bool":    true,
		"a-string":  "hihi",
		"a-float32": float32(12.3456),
		"a-float64": float64(120.3456),
		"bar": map[string]interface{}{
			"baz": "nest egg",
		},
	}
	expected := map[string]interface{}{
		"rule":       "myrule",
		"channel":    "#-partner-",
		"dimensions": []string{"-partner-", "-nest egg-"},
		"msg":        "100, 132, 164, true, 12.3456, 120.3456, hihi, UNKNOWN_VALUE_TYPE",
	}
	actual := r.OutputFor(msg)
	assert.Equal(t, expected, actual)
}

func TestRoute(t *testing.T) {
	router := RuleRouter{rules: []Rule{
		Rule{
			Name: "rule-one",
			Matchers: RuleMatchers{
				"title": []string{"hello", "hi"},
				"foo":   []string{"bar", "baz"},
			},
			Output: RuleOutput{
				"channel":    "#-%{foo}-",
				"dimensions": []string{"-%{foo}-"},
			},
		},
		Rule{
			Name: "rule-two",
			Matchers: RuleMatchers{
				"bing.bong": []string{"buzz"},
			},
			Output: RuleOutput{
				"series": "x",
			},
		},
	}}

	msg0 := map[string]interface{}{
		"title": "hi",
		"foo":   "bar",
	}
	expected0 := []map[string]interface{}{
		map[string]interface{}{
			"rule":       "rule-one",
			"channel":    "#-bar-",
			"dimensions": []string{"-bar-"},
		},
	}
	actual0 := router.Route(msg0)["routes"].([]map[string]interface{})
	assert.Equal(t, expected0, actual0)

	msg1 := map[string]interface{}{
		"title": "hi",
		"bing": map[string]interface{}{
			"bong": "buzz",
		},
	}
	expected1 := []map[string]interface{}{
		map[string]interface{}{
			"rule":   "rule-two",
			"series": "x",
		},
	}
	actual1 := router.Route(msg1)["routes"].([]map[string]interface{})
	assert.Equal(t, expected1, actual1)

	msg2 := map[string]interface{}{
		"title": "hello",
		"foo":   "baz",
		"bing": map[string]interface{}{
			"bong": "buzz",
		},
	}
	expected2 := SortableOutputs([]map[string]interface{}{
		map[string]interface{}{
			"rule":       "rule-one",
			"channel":    "#-baz-",
			"dimensions": []string{"-baz-"},
		},
		map[string]interface{}{
			"rule":   "rule-two",
			"series": "x",
		},
	})
	actual2 := SortableOutputs(router.Route(msg2)["routes"].([]map[string]interface{}))
	sort.Sort(expected2)
	sort.Sort(actual2)
	assert.Equal(t, expected2, actual2)
}
