package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kv "gopkg.in/Clever/kayvee-go.v6"
)

type keyVal map[string]interface{}

// takes two strings (which are assumed to be JSON)
func compareJSONStrings(t *testing.T, expected string, actual string) {
	actualJSON := map[string]interface{}{}
	expectedJSON := map[string]interface{}{}
	err := json.Unmarshal([]byte(actual), &actualJSON)
	if err != nil {
		panic(fmt.Sprint("failed to json unmarshal `actual`:", actual))
	}
	err = json.Unmarshal([]byte(expected), &expectedJSON)
	if err != nil {
		panic(fmt.Sprint("failed to json unmarshal `expected`:", expected))
	}

	expectedJSON["deploy_env"] = "testing"
	expectedJSON["wf_id"] = "abc123"

	assert.Equal(t, expectedJSON, actualJSON)
}

func assertLogFormatAndCompareContent(t *testing.T, logline, expected string) {
	rx := regexp.MustCompile(`\.*({.*})`)
	require.Regexp(t, rx, logline)
	actual := rx.FindStringSubmatch(logline)[1]
	compareJSONStrings(t, expected, actual)
}

func TestLogTrace(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Trace("testlogTrace")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.Format(
		map[string]interface{}{"source": "logger-tester", "level": Trace.String(), "title": "testlogTrace"}))
	buf.Reset()
	logger.TraceD("testlogTrace", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.Format(
		map[string]interface{}{"source": "logger-tester", "level": Trace.String(), "title": "testlogTrace", "key1": "val1", "key2": "val2"}))
}

func TestLogDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Debug("testlogdebug")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.Format(
		map[string]interface{}{"source": "logger-tester", "level": Debug.String(), "title": "testlogdebug"}))
	buf.Reset()
	logger.DebugD("testlogdebug", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.Format(
		map[string]interface{}{"source": "logger-tester", "level": Debug.String(), "title": "testlogdebug", "key1": "val1", "key2": "val2"}))
}

func TestLogInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Info("testloginfo")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testloginfo", map[string]interface{}{}))
	buf.Reset()
	logger.InfoD("testloginfo", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testloginfo", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Warn("testlogwarning")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Warning, "testlogwarning", map[string]interface{}{}))
	buf.Reset()
	logger.WarnD("testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Warning, "testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Error("testlogerror")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Error, "testlogerror", map[string]interface{}{}))
	buf.Reset()
	logger.ErrorD("testlogerror", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Error, "testlogerror", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogCritical(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Critical("testlogcritical")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Critical, "testlogcritical", map[string]interface{}{}))
	buf.Reset()
	logger.CriticalD("testlogcritical", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Critical, "testlogcritical", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogCounter(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Counter("testlogcounter")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testlogcounter", map[string]interface{}{"type": "counter", "value": 1}))
	buf.Reset()
	logger.CounterD("testlogcounter", 2, map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testlogcounter", map[string]interface{}{"key1": "val1", "key2": "val2", "type": "counter", "value": 2}))
}

func TestLogGaugeInt(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.GaugeInt("testloggauge", 0)
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testloggauge", map[string]interface{}{"type": "gauge", "value": 0}))
	buf.Reset()
	logger.GaugeIntD("testloggauge", 4, map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testloggauge", map[string]interface{}{"key1": "val1", "key2": "val2", "type": "gauge", "value": 4}))
}

func TestLogGaugeFloat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.GaugeFloat("testloggauge", 0.0)
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testloggauge", map[string]interface{}{"type": "gauge", "value": 0.0}))
	buf.Reset()
	logger.GaugeFloatD("testloggauge", 4.0, map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), kv.FormatLog(
		"logger-tester", kv.Info, "testloggauge", map[string]interface{}{"key1": "val1", "key2": "val2", "type": "gauge", "value": 4.0}))
}

func TestDiffOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.InfoD("testloginfo", map[string]interface{}{"key1": "val1", "key2": "val2"})
	infoLog := string(buf.Bytes())
	buf2 := &bytes.Buffer{}
	logger.SetOutput(buf2)
	logger.WarnD("testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assert.NotEqual(t, string(buf.Bytes()), string(buf2.Bytes()))
	assert.Equal(t, infoLog, string(buf.Bytes()))
}

func TestHiddenLog(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetLogLevel(Warning)
	logger.SetOutput(buf)
	logger.Debug("testlogdebug")
	assert.Equal(t, "", string(buf.Bytes()))

	buf.Reset()
	logger.Info("testloginfo")
	assert.Equal(t, "", string(buf.Bytes()))

	buf.Reset()
	logger.Warn("testlogwarning")
	assert.NotEqual(t, "", string(buf.Bytes()))

	buf.Reset()
	logger.Error("testlogerror")
	assert.NotEqual(t, "", string(buf.Bytes()))

	buf.Reset()
	logger.Critical("testlogcritical")
	assert.NotEqual(t, "", string(buf.Bytes()))
}

func TestDiffFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.SetFormatter(func(data map[string]interface{}) string { return "This is a test" })
	logger.WarnD("testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assert.Equal(t, "This is a test\n", string(buf.Bytes()))
}

func TestMultipleLoggers(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	logger1 := New("logger-tester1")
	logger2 := New("logger-tester2")
	logger1.SetOutput(buf1)
	logger2.SetOutput(buf2)
	logger1.WarnD("testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"})
	logger2.Info("testloginfo")
	logOutput1 := string(buf1.Bytes())
	assertLogFormatAndCompareContent(t, logOutput1, kv.FormatLog(
		"logger-tester1", kv.Warning, "testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"}))
	assertLogFormatAndCompareContent(t, string(buf2.Bytes()), kv.FormatLog(
		"logger-tester2", kv.Info, "testloginfo", map[string]interface{}{}))

	logger2.SetOutput(buf1)
	logger2.Info("testloginfo")
	assert.NotEqual(t, logOutput1, string(buf1.Bytes()))
}

func TestAddContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)
	logger.Info("1")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()),
		kv.FormatLog("logger-tester", kv.Info, "1", M{}))
	buf.Reset()
	logger.AddContext("a", "b")
	logger.Info("2")
	assertLogFormatAndCompareContent(t, string(buf.Bytes()),
		kv.FormatLog("logger-tester", kv.Info, "2", M{"a": "b"}))
}

func TestFailAddReservedContext(t *testing.T) {
	logger, ok := New("logger-tester").(*Logger)
	assert.True(t, ok)

	reservedKeyNames := map[string]bool{
		"title":  true,
		"source": true,
		"value":  true,
		"type":   true,
		"level":  true,
	}
	testVal := "testingvalue"
	for k := range reservedKeyNames {
		updateContextMapIfNotReserved(logger.globals, k, testVal)
		v := logger.globals[k]
		msg := "Should not be able to set key " + k
		assert.NotEqual(t, testVal, v, msg)
	}
}

type MockRouter struct {
	t      *testing.T
	called bool
}

func TestRouter(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New("logger-tester")
	logger.SetOutput(buf)

	m := MockRouter{t, false}
	logger.SetRouter(&m)
	logger.InfoD("testloginfo", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assert.True(t, m.called)
	expected := kv.FormatLog("logger-tester", kv.Info, "testloginfo", M{
		"key1":    "val1",
		"key2":    "val2",
		"_kvmeta": M{"routekey": 42},
	})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), expected)
}
func (m *MockRouter) Route(msg map[string]interface{}) map[string]interface{} {
	assert.False(m.t, m.called)
	m.called = true
	expected := kv.FormatLog("logger-tester", kv.Info, "testloginfo", M{
		"key1": "val1",
		"key2": "val2",
	})
	assertLogFormatAndCompareContent(m.t, kv.Format(msg), expected)
	return map[string]interface{}{"routekey": 42}
}

func TestLoggerImplementsKayveeLogger(t *testing.T) {
	assert.Implements(t, (*KayveeLogger)(nil), &Logger{}, "*Logger should implement KayveeLogger")
}
