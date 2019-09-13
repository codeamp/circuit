package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	kv "gopkg.in/Clever/kayvee-go.v6"
	"gopkg.in/Clever/kayvee-go.v6/logger"
)

type bufferWriter struct {
	bytes.Buffer
	status int
}

func (b *bufferWriter) WriteHeader(status int) {
	b.status = status
}

func (b *bufferWriter) Header() http.Header {
	return nil
}

func TestMiddleware(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		handler        func(http.ResponseWriter, *http.Request)
		expectedSize   int
		expectedStatus int
		expectedLog    map[string]interface{}
	}{
		{
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write(make([]byte, 10, 10))
				w.Write(make([]byte, 5, 5))
			},
			// Only the logs that vary based on the handler, the rest are tested in the test runner
			expectedLog: map[string]interface{}{
				"level": "info",
				// Floats because json decoding treats all numbers as floats
				"response-size": 15.0,
				"status-code":   200.0,
			},
		},
		{
			// Empty handler is totally valid and should send back 200 with a response size of 0
			handler: func(w http.ResponseWriter, r *http.Request) {},
			expectedLog: map[string]interface{}{
				"level":         "info",
				"response-size": 0.0,
				"status-code":   200.0,
			},
		},
		{
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			},
			expectedLog: map[string]interface{}{
				"level":         "error",
				"response-size": 0.0,
				"status-code":   500.0,
			},
		},
	}
	for _, test := range tests {
		out := &bytes.Buffer{}
		wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
			// inject buffer to capture logs
			logger.FromContext(r.Context()).SetConfig("my-source", logger.Info, kv.Format, out)
			test.handler(w, r)
		}
		handler := New(http.HandlerFunc(wrappedHandler), "my-source")
		rw := &bufferWriter{}
		handler.ServeHTTP(rw, &http.Request{
			Method: "GET",
			URL: &url.URL{
				Host:     "trollhost.com",
				Path:     "path",
				RawQuery: "key=val&key2=val2",
			},
			Header: http.Header{"X-Forwarded-For": {"192.168.0.1"}},
		})

		var result map[string]interface{}
		assert.Nil(json.NewDecoder(out).Decode(&result))

		log.Printf("%#v", result)

		// response-time changes each run, so just check that it's more than zero
		if result["response-time"].(float64) < 1 {
			t.Fatalf("invalid response-time %d", result["response-time"])
		}
		// check that response-time-ms exists, it's usually 0 in these tests
		_, ok := result["response-time-ms"].(float64)
		assert.True(ok, "response-time-ms in log")

		delete(result, "response-time")
		delete(result, "response-time-ms")

		test.expectedLog["ip"] = "192.168.0.1"
		test.expectedLog["path"] = "path"
		test.expectedLog["method"] = "GET"
		test.expectedLog["title"] = "request-finished"
		test.expectedLog["via"] = "kayvee-middleware"
		test.expectedLog["params"] = "key=val&key2=val2"
		test.expectedLog["source"] = "my-source"
		test.expectedLog["count"] = float64(1)
		test.expectedLog["deploy_env"] = "testing"
		test.expectedLog["wf_id"] = "abc123"
		test.expectedLog["canary"] = false
		assert.Equal(test.expectedLog, result)
	}
}

func TestMiddlewareCanaryFlag(t *testing.T) {
	os.Setenv("_CANARY", "1")
	defer os.Unsetenv("_CANARY")

	assert := assert.New(t)

	out := &bytes.Buffer{}

	handler := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.FromContext(r.Context()).SetConfig("my-source", logger.Info, kv.Format, out)
	}), "my-source")

	rw := &bufferWriter{}
	handler.ServeHTTP(rw, &http.Request{
		Method: "GET",
		URL: &url.URL{
			Host: "trollhost.com",
			Path: "path",
		},
	})

	var result map[string]interface{}
	assert.Nil(json.NewDecoder(out).Decode(&result))

	log.Printf("%#v", result)

	assert.Equal(true, result["canary"])

}

func TestMiddlewareIsAddedToContext(t *testing.T) {
	assert := assert.New(t)

	out := &bytes.Buffer{}
	handler := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.FromContext(r.Context()).SetConfig("my-source", logger.Info, kv.Format, out)
		logger.FromContext(r.Context()).Info("logging with context!")
		w.WriteHeader(200)
	}), "my-source")

	rw := &bufferWriter{}
	handler.ServeHTTP(rw, &http.Request{
		Method: "GET",
		URL: &url.URL{
			Host:     "trollhost.com",
			Path:     "path",
			RawQuery: "key=val&key2=val2",
		},
		Header: http.Header{"X-Forwarded-For": {"192.168.0.1"}},
	})

	logLines := strings.Split(out.String(), "\n")
	if len(logLines) != 3 /* one extra blank "line" from trailing newline */ {
		t.Fatalf("expected 2 logs, got %d: %#v", len(logLines)-1, logLines)
	}
	var result map[string]interface{}
	assert.Nil(json.NewDecoder(strings.NewReader(logLines[0])).Decode(&result))

	if result["title"].(string) != "logging with context!" {
		t.Fatalf("invalid log title %s", result["title"])
	}
}
