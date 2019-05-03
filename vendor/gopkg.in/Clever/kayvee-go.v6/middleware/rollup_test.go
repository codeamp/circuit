package middleware

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type RollupLoggerCall struct {
	Title string
	Data  map[string]interface{}
}

type MockRollupLogger struct {
	mu          sync.Mutex
	infoDCalls  []RollupLoggerCall
	errorDCalls []RollupLoggerCall
}

func (m *MockRollupLogger) InfoD(title string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := RollupLoggerCall{title, data}
	m.infoDCalls = append(m.infoDCalls, call)
}

func (m *MockRollupLogger) InfoDCalls() []RollupLoggerCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	var calls []RollupLoggerCall
	for _, call := range m.infoDCalls {
		calls = append(calls, call)
	}
	return calls
}

func (m *MockRollupLogger) ErrorD(title string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := RollupLoggerCall{title, data}
	m.errorDCalls = append(m.errorDCalls, call)
}

func (m *MockRollupLogger) ErrorDCalls() []RollupLoggerCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	var calls []RollupLoggerCall
	for _, call := range m.errorDCalls {
		calls = append(calls, call)
	}
	return calls
}

func TestProcess(t *testing.T) {
	mockLogger := &MockRollupLogger{}
	reportingDelay := 1 * time.Second
	rr := NewRollupRouter(context.Background(), mockLogger, reportingDelay)

	// send a bunch of data to the rollup router in parallel (since logging can
	// happen from multiple goroutines) and you should see it logged as a rollup
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rr.Process(map[string]interface{}{
				"status-code":   200,
				"op":            "healthCheck",
				"method":        "GET",
				"canary":        false,
				"response-time": 100 * time.Millisecond,
			})
		}()
	}
	wg.Wait()

	// check in shortly after reporting delay
	time.Sleep(reportingDelay + 500*time.Millisecond)

	assert.Equal(t, mockLogger.InfoDCalls(), []RollupLoggerCall{
		{
			Title: "request-finished-rollup",
			Data: map[string]interface{}{
				"canary":               false,
				"count":                int64(100),
				"op":                   "healthCheck",
				"method":               "GET",
				"response-time-ms":     int64(100),
				"response-time-ms-sum": int64(100 * 100),
				"status-code":          200,
				"via":                  "kayvee-middleware",
			},
		},
	})
}

func TestShouldRollup(t *testing.T) {
	mockLogger := &MockRollupLogger{}
	reportingDelay := 1 * time.Second
	rr := NewRollupRouter(context.Background(), mockLogger, reportingDelay)

	// if a request is a 200 or is too slow, it should not get rolled up
	// additionally, there needs to be an "op" field
	for _, falseyInput := range []map[string]interface{}{
		map[string]interface{}{
			"status-code":   200,
			"op":            "getApps",
			"method":        "GET",
			"canary":        true,
			"response-time": 600 * time.Millisecond, // too slow
		},
		map[string]interface{}{
			"status-code":   500, // not a 200
			"op":            "getApps",
			"method":        "GET",
			"canary":        true,
			"response-time": 100 * time.Millisecond,
		},
		map[string]interface{}{
			"status-code": 200,
			// no "op" or "method" field
			"canary":        true,
			"response-time": 50 * time.Millisecond,
		},
	} {
		assert.Equal(t, rr.ShouldRollup(falseyInput), false, "expected false return: %v", falseyInput)
	}

	// 200s that are fast enough should get rolled up
	for _, truthyInput := range []map[string]interface{}{
		map[string]interface{}{
			"status-code":   200,
			"op":            "getAppByID",
			"method":        "GET",
			"canary":        true,
			"response-time": 100 * time.Millisecond,
		},
		map[string]interface{}{
			"status-code":   200,
			"op":            "getAdminByID",
			"method":        "GET",
			"canary":        true,
			"response-time": 400 * time.Millisecond,
		},
	} {
		assert.Equal(t, rr.ShouldRollup(truthyInput), true, "expected true return: %v", truthyInput)
	}
}
