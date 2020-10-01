// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package klog

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	stdLog "log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// TODO: This test package should be refactored so that tests cannot
// interfere with each-other.

// Test that shortHostname works as advertised.
func TestShortHostname(t *testing.T) {
	for hostname, expect := range map[string]string{
		"":                "",
		"host":            "host",
		"host.google.com": "host",
	} {
		if got := shortHostname(hostname); expect != got {
			t.Errorf("shortHostname(%q): expected %q, got %q", hostname, expect, got)
		}
	}
}

// flushBuffer wraps a bytes.Buffer to satisfy flushSyncWriter.
type flushBuffer struct {
	bytes.Buffer
}

func (f *flushBuffer) Flush() error {
	return nil
}

func (f *flushBuffer) Sync() error {
	return nil
}

// swap sets the log writers and returns the old array.
func (l *loggingT) swap(writers [numSeverity]flushSyncWriter) (old [numSeverity]flushSyncWriter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	old = l.file
	for i, w := range writers {
		logging.file[i] = w
	}
	return
}

// newBuffers sets the log writers to all new byte buffers and returns the old array.
func (l *loggingT) newBuffers() [numSeverity]flushSyncWriter {
	return l.swap([numSeverity]flushSyncWriter{new(flushBuffer), new(flushBuffer), new(flushBuffer), new(flushBuffer)})
}

// contents returns the specified log value as a string.
func contents(s severity) string {
	return logging.file[s].(*flushBuffer).String()
}

// contains reports whether the string is contained in the log.
func contains(s severity, str string, t *testing.T) bool {
	return strings.Contains(contents(s), str)
}

// setFlags configures the logging flags how the test expects them.
func setFlags() {
	logging.toStderr = false
	logging.addDirHeader = false
}

// Test that Info works as advertised.
func TestInfo(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	Info("test")
	if !contains(infoLog, "I", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	if !contains(infoLog, "test", t) {
		t.Error("Info failed")
	}
}

func TestInfoDepth(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())

	f := func() { InfoDepth(1, "depth-test1") }

	// The next three lines must stay together
	_, _, wantLine, _ := runtime.Caller(0)
	InfoDepth(0, "depth-test0")
	f()

	msgs := strings.Split(strings.TrimSuffix(contents(infoLog), "\n"), "\n")
	if len(msgs) != 2 {
		t.Fatalf("Got %d lines, expected 2", len(msgs))
	}

	for i, m := range msgs {
		if !strings.HasPrefix(m, "I") {
			t.Errorf("InfoDepth[%d] has wrong character: %q", i, m)
		}
		w := fmt.Sprintf("depth-test%d", i)
		if !strings.Contains(m, w) {
			t.Errorf("InfoDepth[%d] missing %q: %q", i, w, m)
		}

		// pull out the line number (between : and ])
		msg := m[strings.LastIndex(m, ":")+1:]
		x := strings.Index(msg, "]")
		if x < 0 {
			t.Errorf("InfoDepth[%d]: missing ']': %q", i, m)
			continue
		}
		line, err := strconv.Atoi(msg[:x])
		if err != nil {
			t.Errorf("InfoDepth[%d]: bad line number: %q", i, m)
			continue
		}
		wantLine++
		if wantLine != line {
			t.Errorf("InfoDepth[%d]: got line %d, want %d", i, line, wantLine)
		}
	}
}

func init() {
	CopyStandardLogTo("INFO")
}

// Test that CopyStandardLogTo panics on bad input.
func TestCopyStandardLogToPanic(t *testing.T) {
	defer func() {
		if s, ok := recover().(string); !ok || !strings.Contains(s, "LOG") {
			t.Errorf(`CopyStandardLogTo("LOG") should have panicked: %v`, s)
		}
	}()
	CopyStandardLogTo("LOG")
}

// Test that using the standard log package logs to INFO.
func TestStandardLog(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	stdLog.Print("test")
	if !contains(infoLog, "I", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	if !contains(infoLog, "test", t) {
		t.Error("Info failed")
	}
}

// Test that the header has the correct format.
func TestHeader(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	defer func(previous func() time.Time) { timeNow = previous }(timeNow)
	timeNow = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.Local)
	}
	pid = 1234
	Info("test")
	var line int
	format := "I0102 15:04:05.067890    1234 klog_test.go:%d] test\n"
	n, err := fmt.Sscanf(contents(infoLog), format, &line)
	if n != 1 || err != nil {
		t.Errorf("log format error: %d elements, error %s:\n%s", n, err, contents(infoLog))
	}
	// Scanf treats multiple spaces as equivalent to a single space,
	// so check for correct space-padding also.
	want := fmt.Sprintf(format, line)
	if contents(infoLog) != want {
		t.Errorf("log format error: got:\n\t%q\nwant:\t%q", contents(infoLog), want)
	}
}

func TestHeaderWithDir(t *testing.T) {
	setFlags()
	logging.addDirHeader = true
	defer logging.swap(logging.newBuffers())
	defer func(previous func() time.Time) { timeNow = previous }(timeNow)
	timeNow = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.Local)
	}
	pid = 1234
	Info("test")
	re := regexp.MustCompile(`I0102 15:04:05.067890    1234 (klog|v2)/klog_test.go:(\d+)] test\n`)
	if !re.MatchString(contents(infoLog)) {
		t.Errorf("log format error: line does not match regex:\n\t%q\n", contents(infoLog))
	}
}

// Test that an Error log goes to Warning and Info.
// Even in the Info log, the source character will be E, so the data should
// all be identical.
func TestError(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	Error("test")
	if !contains(errorLog, "E", t) {
		t.Errorf("Error has wrong character: %q", contents(errorLog))
	}
	if !contains(errorLog, "test", t) {
		t.Error("Error failed")
	}
	str := contents(errorLog)
	if !contains(warningLog, str, t) {
		t.Error("Warning failed")
	}
	if !contains(infoLog, str, t) {
		t.Error("Info failed")
	}
}

// Test that a Warning log goes to Info.
// Even in the Info log, the source character will be W, so the data should
// all be identical.
func TestWarning(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	Warning("test")
	if !contains(warningLog, "W", t) {
		t.Errorf("Warning has wrong character: %q", contents(warningLog))
	}
	if !contains(warningLog, "test", t) {
		t.Error("Warning failed")
	}
	str := contents(warningLog)
	if !contains(infoLog, str, t) {
		t.Error("Info failed")
	}
}

// Test that a V log goes to Info.
func TestV(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	logging.verbosity.Set("2")
	defer logging.verbosity.Set("0")
	V(2).Info("test")
	if !contains(infoLog, "I", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	if !contains(infoLog, "test", t) {
		t.Error("Info failed")
	}
}

// Test that a vmodule enables a log in this file.
func TestVmoduleOn(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	logging.vmodule.Set("klog_test=2")
	defer logging.vmodule.Set("")
	if !V(1).Enabled() {
		t.Error("V not enabled for 1")
	}
	if !V(2).Enabled() {
		t.Error("V not enabled for 2")
	}
	if V(3).Enabled() {
		t.Error("V enabled for 3")
	}
	V(2).Info("test")
	if !contains(infoLog, "I", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	if !contains(infoLog, "test", t) {
		t.Error("Info failed")
	}
}

// Test that a vmodule of another file does not enable a log in this file.
func TestVmoduleOff(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	logging.vmodule.Set("notthisfile=2")
	defer logging.vmodule.Set("")
	for i := 1; i <= 3; i++ {
		if V(Level(i)).Enabled() {
			t.Errorf("V enabled for %d", i)
		}
	}
	V(2).Info("test")
	if contents(infoLog) != "" {
		t.Error("V logged incorrectly")
	}
}

func TestSetOutputDataRace(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	var wg sync.WaitGroup
	for i := 1; i <= 50; i++ {
		go func() {
			logging.flushDaemon()
		}()
	}
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			SetOutput(ioutil.Discard)
		}()
	}
	for i := 1; i <= 50; i++ {
		go func() {
			logging.flushDaemon()
		}()
	}
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			SetOutputBySeverity("INFO", ioutil.Discard)
		}()
	}
	for i := 1; i <= 50; i++ {
		go func() {
			logging.flushDaemon()
		}()
	}
	wg.Wait()
}

func TestLogToOutput(t *testing.T) {
	logging.toStderr = true
	defer logging.swap(logging.newBuffers())
	buf := new(bytes.Buffer)
	SetOutput(buf)
	LogToStderr(false)

	Info("Does logging to an output work?")

	str := buf.String()
	if !strings.Contains(str, "Does logging to an output work?") {
		t.Fatalf("Expected %q to contain \"Does logging to an output work?\"", str)
	}
}

// vGlobs are patterns that match/don't match this file at V=2.
var vGlobs = map[string]bool{
	// Easy to test the numeric match here.
	"klog_test=1": false, // If -vmodule sets V to 1, V(2) will fail.
	"klog_test=2": true,
	"klog_test=3": true, // If -vmodule sets V to 1, V(3) will succeed.
	// These all use 2 and check the patterns. All are true.
	"*=2":           true,
	"?l*=2":         true,
	"????_*=2":      true,
	"??[mno]?_*t=2": true,
	// These all use 2 and check the patterns. All are false.
	"*x=2":         false,
	"m*=2":         false,
	"??_*=2":       false,
	"?[abc]?_*t=2": false,
}

// Test that vmodule globbing works as advertised.
func testVmoduleGlob(pat string, match bool, t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	defer logging.vmodule.Set("")
	logging.vmodule.Set(pat)
	if V(2).Enabled() != match {
		t.Errorf("incorrect match for %q: got %t expected %t", pat, V(2), match)
	}
}

// Test that a vmodule globbing works as advertised.
func TestVmoduleGlob(t *testing.T) {
	for glob, match := range vGlobs {
		testVmoduleGlob(glob, match, t)
	}
}

func TestRollover(t *testing.T) {
	setFlags()
	var err error
	defer func(previous func(error)) { logExitFunc = previous }(logExitFunc)
	logExitFunc = func(e error) {
		err = e
	}
	defer func(previous uint64) { MaxSize = previous }(MaxSize)
	MaxSize = 512
	Info("x") // Be sure we have a file.
	info, ok := logging.file[infoLog].(*syncBuffer)
	if !ok {
		t.Fatal("info wasn't created")
	}
	if err != nil {
		t.Fatalf("info has initial error: %v", err)
	}
	fname0 := info.file.Name()
	Info(strings.Repeat("x", int(MaxSize))) // force a rollover
	if err != nil {
		t.Fatalf("info has error after big write: %v", err)
	}

	// Make sure the next log file gets a file name with a different
	// time stamp.
	//
	// TODO: determine whether we need to support subsecond log
	// rotation.  C++ does not appear to handle this case (nor does it
	// handle Daylight Savings Time properly).
	time.Sleep(1 * time.Second)

	Info("x") // create a new file
	if err != nil {
		t.Fatalf("error after rotation: %v", err)
	}
	fname1 := info.file.Name()
	if fname0 == fname1 {
		t.Errorf("info.f.Name did not change: %v", fname0)
	}
	if info.nbytes >= info.maxbytes {
		t.Errorf("file size was not reset: %d", info.nbytes)
	}
}

func TestOpenAppendOnStart(t *testing.T) {
	const (
		x string = "xxxxxxxxxx"
		y string = "yyyyyyyyyy"
	)

	setFlags()
	var err error
	defer func(previous func(error)) { logExitFunc = previous }(logExitFunc)
	logExitFunc = func(e error) {
		err = e
	}

	f, err := ioutil.TempFile("", "test_klog_OpenAppendOnStart")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	logging.logFile = f.Name()

	// Erase files created by prior tests,
	for i := range logging.file {
		logging.file[i] = nil
	}

	// Logging creates the file
	Info(x)
	_, ok := logging.file[infoLog].(*syncBuffer)
	if !ok {
		t.Fatal("info wasn't created")
	}
	if err != nil {
		t.Fatalf("info has initial error: %v", err)
	}
	// ensure we wrote what we expected
	logging.flushAll()
	b, err := ioutil.ReadFile(logging.logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(b), x) {
		t.Fatalf("got %s, missing expected Info log: %s", string(b), x)
	}

	// Set the file to nil so it gets "created" (opened) again on the next write.
	for i := range logging.file {
		logging.file[i] = nil
	}

	// Logging agagin should open the file again with O_APPEND instead of O_TRUNC
	Info(y)
	// ensure we wrote what we expected
	logging.lockAndFlushAll()
	b, err = ioutil.ReadFile(logging.logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(b), y) {
		t.Fatalf("got %s, missing expected Info log: %s", string(b), y)
	}
	// The initial log message should be preserved across create calls.
	logging.lockAndFlushAll()
	b, err = ioutil.ReadFile(logging.logFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(b), x) {
		t.Fatalf("got %s, missing expected Info log: %s", string(b), x)
	}
}

func TestLogBacktraceAt(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	// The peculiar style of this code simplifies line counting and maintenance of the
	// tracing block below.
	var infoLine string
	setTraceLocation := func(file string, line int, ok bool, delta int) {
		if !ok {
			t.Fatal("could not get file:line")
		}
		_, file = filepath.Split(file)
		infoLine = fmt.Sprintf("%s:%d", file, line+delta)
		err := logging.traceLocation.Set(infoLine)
		if err != nil {
			t.Fatal("error setting log_backtrace_at: ", err)
		}
	}
	{
		// Start of tracing block. These lines know about each other's relative position.
		_, file, line, ok := runtime.Caller(0)
		setTraceLocation(file, line, ok, +2) // Two lines between Caller and Info calls.
		Info("we want a stack trace here")
	}
	numAppearances := strings.Count(contents(infoLog), infoLine)
	if numAppearances < 2 {
		// Need 2 appearances, one in the log header and one in the trace:
		//   log_test.go:281: I0511 16:36:06.952398 02238 log_test.go:280] we want a stack trace here
		//   ...
		//   k8s.io/klog/klog_test.go:280 (0x41ba91)
		//   ...
		// We could be more precise but that would require knowing the details
		// of the traceback format, which may not be dependable.
		t.Fatal("got no trace back; log is ", contents(infoLog))
	}
}

func BenchmarkHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf, _, _ := logging.header(infoLog, 0)
		logging.putBuffer(buf)
	}
}

func BenchmarkHeaderWithDir(b *testing.B) {
	logging.addDirHeader = true
	for i := 0; i < b.N; i++ {
		buf, _, _ := logging.header(infoLog, 0)
		logging.putBuffer(buf)
	}
}

func BenchmarkLogs(b *testing.B) {
	setFlags()
	defer logging.swap(logging.newBuffers())

	testFile, err := ioutil.TempFile("", "test.log")
	if err != nil {
		b.Error("unable to create temporary file")
	}
	defer os.Remove(testFile.Name())

	logging.verbosity.Set("0")
	logging.toStderr = false
	logging.alsoToStderr = false
	logging.stderrThreshold = fatalLog
	logging.logFile = testFile.Name()
	logging.swap([numSeverity]flushSyncWriter{nil, nil, nil, nil})

	for i := 0; i < b.N; i++ {
		Error("error")
		Warning("warning")
		Info("info")
	}
	logging.flushAll()
}

// Test the logic on checking log size limitation.
func TestFileSizeCheck(t *testing.T) {
	setFlags()
	testData := map[string]struct {
		testLogFile          string
		testLogFileMaxSizeMB uint64
		testCurrentSize      uint64
		expectedResult       bool
	}{
		"logFile not specified, exceeds max size": {
			testLogFile:          "",
			testLogFileMaxSizeMB: 1,
			testCurrentSize:      1024 * 1024 * 2000, //exceeds the maxSize
			expectedResult:       true,
		},

		"logFile not specified, not exceeds max size": {
			testLogFile:          "",
			testLogFileMaxSizeMB: 1,
			testCurrentSize:      1024 * 1024 * 1000, //smaller than the maxSize
			expectedResult:       false,
		},
		"logFile specified, exceeds max size": {
			testLogFile:          "/tmp/test.log",
			testLogFileMaxSizeMB: 500,                // 500MB
			testCurrentSize:      1024 * 1024 * 1000, //exceeds the logFileMaxSizeMB
			expectedResult:       true,
		},
		"logFile specified, not exceeds max size": {
			testLogFile:          "/tmp/test.log",
			testLogFileMaxSizeMB: 500,               // 500MB
			testCurrentSize:      1024 * 1024 * 300, //smaller than the logFileMaxSizeMB
			expectedResult:       false,
		},
	}

	for name, test := range testData {
		logging.logFile = test.testLogFile
		logging.logFileMaxSizeMB = test.testLogFileMaxSizeMB
		actualResult := test.testCurrentSize >= CalculateMaxSize()
		if test.expectedResult != actualResult {
			t.Fatalf("Error on test case '%v': Was expecting result equals %v, got %v",
				name, test.expectedResult, actualResult)
		}
	}
}

func TestInitFlags(t *testing.T) {
	fs1 := flag.NewFlagSet("test1", flag.PanicOnError)
	InitFlags(fs1)
	fs1.Set("log_dir", "/test1")
	fs1.Set("log_file_max_size", "1")
	fs2 := flag.NewFlagSet("test2", flag.PanicOnError)
	InitFlags(fs2)
	if logging.logDir != "/test1" {
		t.Fatalf("Expected log_dir to be %q, got %q", "/test1", logging.logDir)
	}
	fs2.Set("log_file_max_size", "2048")
	if logging.logFileMaxSizeMB != 2048 {
		t.Fatal("Expected log_file_max_size to be 2048")
	}
}

func TestInfoObjectRef(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())

	tests := []struct {
		name string
		ref  ObjectRef
		want string
	}{
		{
			name: "with ns",
			ref: ObjectRef{
				Name:      "test-name",
				Namespace: "test-ns",
			},
			want: "test-ns/test-name",
		},
		{
			name: "without ns",
			ref: ObjectRef{
				Name:      "test-name",
				Namespace: "",
			},
			want: "test-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Info(tt.ref)
			if !contains(infoLog, tt.want, t) {
				t.Errorf("expected %v, got %v", tt.want, contents(infoLog))
			}
		})
	}
}

type mockKmeta struct {
	name, ns string
}

func (m mockKmeta) GetName() string {
	return m.name
}
func (m mockKmeta) GetNamespace() string {
	return m.ns
}

func TestKObj(t *testing.T) {
	tests := []struct {
		name string
		obj  KMetadata
		want ObjectRef
	}{
		{
			name: "with ns",
			obj:  mockKmeta{"test-name", "test-ns"},
			want: ObjectRef{
				Name:      "test-name",
				Namespace: "test-ns",
			},
		},
		{
			name: "without ns",
			obj:  mockKmeta{"test-name", ""},
			want: ObjectRef{
				Name: "test-name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if KObj(tt.obj) != tt.want {
				t.Errorf("expected %v, got %v", tt.want, KObj(tt.obj))
			}
		})
	}
}

func TestKRef(t *testing.T) {
	tests := []struct {
		testname  string
		name      string
		namespace string
		want      ObjectRef
	}{
		{
			testname:  "with ns",
			name:      "test-name",
			namespace: "test-ns",
			want: ObjectRef{
				Name:      "test-name",
				Namespace: "test-ns",
			},
		},
		{
			testname: "without ns",
			name:     "test-name",
			want: ObjectRef{
				Name: "test-name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			if KRef(tt.namespace, tt.name) != tt.want {
				t.Errorf("expected %v, got %v", tt.want, KRef(tt.namespace, tt.name))
			}
		})
	}
}

// Test that InfoS works as advertised.
func TestInfoS(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	timeNow = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.Local)
	}
	pid = 1234
	var testDataInfo = []struct {
		msg        string
		format     string
		keysValues []interface{}
	}{
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" pod=\"kubedns\"\n",
			keysValues: []interface{}{"pod", "kubedns"},
		},
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" replicaNum=20\n",
			keysValues: []interface{}{"replicaNum", 20},
		},
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" err=\"test error\"\n",
			keysValues: []interface{}{"err", errors.New("test error")},
		},
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" err=\"test error\"\n",
			keysValues: []interface{}{"err", errors.New("test error")},
		},
	}

	for _, data := range testDataInfo {
		logging.file[infoLog] = &flushBuffer{}
		InfoS(data.msg, data.keysValues...)
		var line int
		n, err := fmt.Sscanf(contents(infoLog), data.format, &line)
		if n != 1 || err != nil {
			t.Errorf("log format error: %d elements, error %s:\n%s", n, err, contents(infoLog))
		}
		want := fmt.Sprintf(data.format, line)
		if contents(infoLog) != want {
			t.Errorf("InfoS has wrong format: \n got:\t%s\nwant:\t%s", contents(infoLog), want)
		}
	}
}

// Test that Verbose.InfoS works as advertised.
func TestVInfoS(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	timeNow = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.Local)
	}
	pid = 1234
	var testDataInfo = []struct {
		msg        string
		format     string
		keysValues []interface{}
	}{
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" pod=\"kubedns\"\n",
			keysValues: []interface{}{"pod", "kubedns"},
		},
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" replicaNum=20\n",
			keysValues: []interface{}{"replicaNum", 20},
		},
		{
			msg:        "test",
			format:     "I0102 15:04:05.067890    1234 klog_test.go:%d] \"test\" err=\"test error\"\n",
			keysValues: []interface{}{"err", errors.New("test error")},
		},
	}

	logging.verbosity.Set("2")
	defer logging.verbosity.Set("0")

	for l := Level(0); l < Level(4); l++ {
		for _, data := range testDataInfo {
			logging.file[infoLog] = &flushBuffer{}

			V(l).InfoS(data.msg, data.keysValues...)

			var want string
			var line int
			if l <= 2 {
				n, err := fmt.Sscanf(contents(infoLog), data.format, &line)
				if n != 1 || err != nil {
					t.Errorf("log format error: %d elements, error %s:\n%s", n, err, contents(infoLog))
				}

				want = fmt.Sprintf(data.format, line)
			} else {
				want = ""
			}
			if contents(infoLog) != want {
				t.Errorf("V(%d).InfoS has unexpected output: \n got:\t%s\nwant:\t%s", l, contents(infoLog), want)
			}
		}
	}
}

// Test that ErrorS works as advertised.
func TestErrorS(t *testing.T) {
	setFlags()
	defer logging.swap(logging.newBuffers())
	timeNow = func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.Local)
	}
	logging.logFile = ""
	pid = 1234
	ErrorS(fmt.Errorf("update status failed"), "Failed to update pod status", "pod", "kubedns")
	var line int
	format := "E0102 15:04:05.067890    1234 klog_test.go:%d] \"Failed to update pod status\" err=\"update status failed\" pod=\"kubedns\"\n"
	n, err := fmt.Sscanf(contents(errorLog), format, &line)
	if n != 1 || err != nil {
		t.Errorf("log format error: %d elements, error %s:\n%s", n, err, contents(errorLog))
	}
	want := fmt.Sprintf(format, line)
	if contents(errorLog) != want {
		t.Errorf("ErrorS has wrong format: \n got:\t%s\nwant:\t%s", contents(errorLog), want)
	}
}

// Test that kvListFormat works as advertised.
func TestKvListFormat(t *testing.T) {
	var testKVList = []struct {
		keysValues []interface{}
		want       string
	}{
		{
			keysValues: []interface{}{"pod", "kubedns"},
			want:       " pod=\"kubedns\"",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "update", true},
			want:       " pod=\"kubedns\" update=true",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "spec", struct {
				X int
				Y string
				N time.Time
			}{X: 76, Y: "strval", N: time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.UTC)}},
			want: " pod=\"kubedns\" spec={X:76 Y:strval N:2006-01-02 15:04:05.06789 +0000 UTC}",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "values", []int{8, 6, 7, 5, 3, 0, 9}},
			want:       " pod=\"kubedns\" values=[8 6 7 5 3 0 9]",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "values", []string{"deployment", "svc", "configmap"}},
			want:       " pod=\"kubedns\" values=[deployment svc configmap]",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "maps", map[string]int{"three": 4}},
			want:       " pod=\"kubedns\" maps=map[three:4]",
		},
		{
			keysValues: []interface{}{"pod", KRef("kube-system", "kubedns"), "status", "ready"},
			want:       " pod=\"kube-system/kubedns\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", KRef("", "kubedns"), "status", "ready"},
			want:       " pod=\"kubedns\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", KObj(mockKmeta{"test-name", "test-ns"}), "status", "ready"},
			want:       " pod=\"test-ns/test-name\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", KObj(mockKmeta{"test-name", ""}), "status", "ready"},
			want:       " pod=\"test-name\" status=\"ready\"",
		},
	}

	for _, d := range testKVList {
		b := &bytes.Buffer{}
		kvListFormat(b, d.keysValues...)
		if b.String() != d.want {
			t.Errorf("kvlist format error:\n got:\n\t%s\nwant:\t%s", b.String(), d.want)
		}
	}
}

func createTestValueOfLoggingT() *loggingT {
	l := new(loggingT)
	l.toStderr = true
	l.alsoToStderr = false
	l.stderrThreshold = errorLog
	l.verbosity = Level(0)
	l.skipHeaders = false
	l.skipLogHeaders = false
	l.addDirHeader = false
	return l
}

func createTestValueOfModulePat(p string, li bool, le Level) modulePat {
	m := modulePat{}
	m.pattern = p
	m.literal = li
	m.level = le
	return m
}

func compareModuleSpec(a, b moduleSpec) bool {
	if len(a.filter) != len(b.filter) {
		return false
	}

	for i := 0; i < len(a.filter); i++ {
		if a.filter[i] != b.filter[i] {
			return false
		}
	}

	return true
}

func TestSetVState(t *testing.T) {
	//Target loggingT value
	want := createTestValueOfLoggingT()
	want.verbosity = Level(3)
	want.vmodule.filter = []modulePat{
		createTestValueOfModulePat("recordio", true, Level(2)),
		createTestValueOfModulePat("file", true, Level(1)),
		createTestValueOfModulePat("gfs*", false, Level(3)),
		createTestValueOfModulePat("gopher*", false, Level(3)),
	}
	want.filterLength = 4

	//loggingT value to which test is run
	target := createTestValueOfLoggingT()

	tf := []modulePat{
		createTestValueOfModulePat("recordio", true, Level(2)),
		createTestValueOfModulePat("file", true, Level(1)),
		createTestValueOfModulePat("gfs*", false, Level(3)),
		createTestValueOfModulePat("gopher*", false, Level(3)),
	}

	target.setVState(Level(3), tf, true)

	if want.verbosity != target.verbosity || !compareModuleSpec(want.vmodule, target.vmodule) || want.filterLength != target.filterLength {
		t.Errorf("setVState method doesn't configure loggingT values' verbosity, vmodule or filterLength:\nwant:\n\tverbosity:\t%v\n\tvmodule:\t%v\n\tfilterLength:\t%v\ngot:\n\tverbosity:\t%v\n\tvmodule:\t%v\n\tfilterLength:\t%v", want.verbosity, want.vmodule, want.filterLength, target.verbosity, target.vmodule, target.filterLength)
	}
}
