package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	router "gopkg.in/Clever/kayvee-go.v6/router"
)

func TestMockLoggerImplementsKayveeLogger(t *testing.T) {
	assert.Implements(t, (*KayveeLogger)(nil), &MockRouteCountLogger{}, "*MockRouteCountLogger should implement KayveeLogger")
}

func TestRouteCountsWithMockLogger(t *testing.T) {
	routes := map[string](router.Rule){
		"rule-one": router.Rule{
			Matchers: router.RuleMatchers{
				"foo": []string{"bar", "baz"},
			},
			Output: router.RuleOutput{
				"out": "#-%{foo}-",
			},
		},
		"rule-two": router.Rule{
			Matchers: router.RuleMatchers{
				"abc": []string{"def"},
			},
			Output: router.RuleOutput{
				"more": "x",
			},
		},
	}
	testRouter, err := router.NewFromRoutes(routes)
	assert.NoError(t, err)

	mockLogger := NewMockCountLogger("testing")
	mockLogger.SetRouter(testRouter)

	t.Log("log0")
	data0 := M{
		"wrong": "stuff",
	}
	mockLogger.InfoD("log0", data0)

	t.Log("log0 -- verify rule counts")
	actualCounts0 := mockLogger.RuleCounts()
	expectedCounts0 := map[string]int{}
	assert.Equal(t, expectedCounts0, actualCounts0)

	t.Log("log0 -- verify rule matches")
	actualRoutes0 := mockLogger.RuleOutputs()
	expectedRoutes0 := map[string][]router.RuleOutput{}
	assert.Equal(t, expectedRoutes0, actualRoutes0)

	t.Log("log1")
	data1 := M{
		"foo": "bar",
	}
	mockLogger.InfoD("log1", data1)

	t.Log("log1 -- verify rule counts")
	actualCounts1 := mockLogger.RuleCounts()
	expectedCounts1 := map[string]int{"rule-one": 1}
	assert.Equal(t, expectedCounts1, actualCounts1)

	t.Log("log1 -- verify rule matches")
	actualRoutes1 := mockLogger.RuleOutputs()
	expectedRoutes1 := map[string][]router.RuleOutput{
		"rule-one": []router.RuleOutput{
			router.RuleOutput{
				"rule": "rule-one",
				"out":  "#-bar-",
			},
		},
	}
	assert.Equal(t, expectedRoutes1, actualRoutes1)

	t.Log("log2")
	data2 := M{
		"foo": "baz",
		"abc": "def",
	}
	mockLogger.InfoD("log2", data2)

	t.Log("log2 -- verify rule counts")
	actualCounts2 := mockLogger.RuleCounts()
	expectedCounts2 := map[string]int{
		"rule-one": 2,
		"rule-two": 1,
	}
	assert.Equal(t, expectedCounts2, actualCounts2)

	t.Log("log2 -- verify rule matches")

	expectedRoutes2 := map[string][]router.RuleOutput{
		"rule-one": []router.RuleOutput{
			router.RuleOutput{
				"rule": "rule-one",
				"out":  "#-bar-",
			},
			router.RuleOutput{
				"rule": "rule-one",
				"out":  "#-baz-",
			},
		},
		"rule-two": []router.RuleOutput{
			router.RuleOutput{
				"rule": "rule-two",
				"more": "x",
			},
		},
	}

	actualRoutes2 := mockLogger.RuleOutputs()
	assert.Equal(t, expectedRoutes2, actualRoutes2)
}
