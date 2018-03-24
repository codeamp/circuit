package transistor

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"time"

	json "github.com/bww/go-json"
	log "github.com/codeamp/logger"
	"github.com/satori/go.uuid"
)

type Event struct {
	ID           uuid.UUID   `json:"id"`
	ParentID     uuid.UUID   `json:"parentId"`
	Name         string      `json:"name"`
	Payload      interface{} `json:"payload"`
	PayloadModel string      `json:"payloadModel"`
	Error        error       `json:"error"`
	CreatedAt    time.Time   `json:"createdAt"`
	Caller       Caller      `json:"caller"`
	Artifacts    []Artifact  `json:"artifacts"`
}

type Caller struct {
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
}

type Artifact struct {
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Secret bool        `json:"secret"`
}

func (a *Artifact) GetString() string {
	return a.Value.(string)
}

func (a *Artifact) GetInt() int {
	i, err := strconv.Atoi(a.Value.(string))
	if err != nil {
		log.Error(err)
	}

	return i
}

func (a *Artifact) GetStringMap() map[string]interface{} {
	return a.Value.(map[string]interface{})
}

func name(payload interface{}) string {
	s := reflect.ValueOf(payload)

	if s.Kind() != reflect.Struct {
		return reflect.TypeOf(payload).String()
	}

	name := reflect.TypeOf(payload).String()

	f := s.FieldByName("Action")
	if f.IsValid() {
		action := f.String()
		if action != "" {
			name = fmt.Sprintf("%v:%v", name, action)
		}
	}

	f = s.FieldByName("Slug")
	if f.IsValid() {
		slug := f.String()
		if slug != "" {
			name = fmt.Sprintf("%v:%v", name, slug)
		}
	}

	return name
}

func NewEvent(payload interface{}, err error) Event {
	event := Event{
		ID:           uuid.NewV4(),
		Name:         name(payload),
		Payload:      payload,
		PayloadModel: reflect.TypeOf(payload).String(),
		Error:        err,
		CreatedAt:    time.Now(),
	}

	// for debugging purposes
	_, file, no, ok := runtime.Caller(1)
	if ok {
		event.Caller = Caller{
			File:       file,
			LineNumber: no,
		}
	}

	return event
}

func (e *Event) NewEvent(payload interface{}, err error) Event {
	event := Event{
		ID:           uuid.NewV4(),
		ParentID:     e.ID,
		Name:         name(payload),
		Payload:      payload,
		PayloadModel: reflect.TypeOf(payload).String(),
		Error:        err,
		CreatedAt:    time.Now(),
	}

	// for debugging purposes
	_, file, no, ok := runtime.Caller(1)
	if ok {
		event.Caller = Caller{
			File:       file,
			LineNumber: no,
		}
	}

	return event
}

func (e *Event) Dump() {
	event, _ := json.MarshalRole("dummy", e)
	log.Info(string(event))
}

func (e *Event) Matches(name string) bool {
	matched, err := regexp.MatchString(name, e.Name)
	if err != nil {
		log.InfoWithFields("Event regex match encountered an error", log.Fields{
			"regex":  name,
			"string": e.Name,
			"error":  err,
		})
	}

	if matched {
		return true
	}

	log.DebugWithFields("Event regex not matched", log.Fields{
		"regex":  name,
		"string": e.Name,
	})

	return false
}

func (e *Event) AddArtifact(key string, value interface{}, secret bool) {
	artifact := Artifact{
		Key:    key,
		Value:  value,
		Secret: secret,
	}
	e.Artifacts = append(e.Artifacts, artifact)
}

func (e *Event) GetArtifact(key string) (Artifact, error) {
	for _, artifact := range e.Artifacts {
		if artifact.Key == key {
			return artifact, nil
		}
	}

	return Artifact{}, errors.New(fmt.Sprintf("Artifact %s not found", key))
}
